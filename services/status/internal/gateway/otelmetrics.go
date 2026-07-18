package gateway

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/protobuf/types/known/timestamppb"

	mev1 "github.com/pood1e/realtime-me/gen/go/realtime/me/status/v1"
)

// metricDefinition names one gauge exported to Prometheus. Names and units
// follow OpenTelemetry semantic conventions; the OTel Prometheus exporter turns
// them into the stable Prometheus series consumed by the status page.
type metricDefinition struct {
	Name        string
	Unit        string
	Description string
}

// metricDefinitions covers only what the gateway itself observes: the phone and
// watch it accepts pushes from, and its own GitHub sync state. Host, VM, and
// agent series are exported by the server's node-exporter or each host's unified
// probe, which Prometheus scrapes directly; re-publishing them would duplicate series.
var metricDefinitions = []metricDefinition{
	{Name: "realtime_device_last_update_time_seconds", Unit: "s", Description: "Unix timestamp of the latest accepted device update."},
	{Name: "realtime_device_battery_level_ratio", Unit: "1", Description: "Device battery level as a fraction of total capacity."},
	{Name: "realtime_device_charging", Unit: "1", Description: "Device charging state: 1 for charging, 0 otherwise."},
	{Name: "realtime_device_network_state", Unit: "1", Description: "Phone network state labelled by network_type; current state is 1."},
	{Name: "realtime_device_accessory_connected", Unit: "1", Description: "Connected accessory state labelled by accessory name and kind."},
	{Name: "realtime_device_accessory_battery_level_ratio", Unit: "1", Description: "Accessory battery level as a fraction of total capacity."},
	{Name: "realtime_watch_heart_rate_beats_per_minute", Unit: "{beat}/min", Description: "Latest watch heart rate."},
	{Name: "realtime_watch_steps", Unit: "{step}", Description: "Latest watch local-day step count."},
	{Name: "realtime_github_status_sync_state", Unit: "1", Description: "GitHub status sync state labelled by state; current state is 1."},
}

// MetricsExporter serves /metrics via the OpenTelemetry SDK and its Prometheus
// exporter. A single observable callback reads one store snapshot per scrape and
// bridges the pushed status into the pull-based Prometheus model.
type MetricsExporter struct {
	store   *StatusStore
	handler http.Handler
	gauges  map[string]otelmetric.Float64ObservableGauge
}

func NewMetricsExporter(store *StatusStore) (*MetricsExporter, error) {
	registry := prometheus.NewRegistry()
	exporter, err := otelprom.New(
		otelprom.WithRegisterer(registry),
		otelprom.WithoutUnits(),
		otelprom.WithoutScopeInfo(),
		otelprom.WithoutTargetInfo(),
		otelprom.WithoutCounterSuffixes(),
	)
	if err != nil {
		return nil, err
	}
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	meter := provider.Meter("realtime-me/status-gateway")

	exporterMetrics := &MetricsExporter{
		store:  store,
		gauges: make(map[string]otelmetric.Float64ObservableGauge, len(metricDefinitions)),
	}
	instruments := make([]otelmetric.Observable, 0, len(metricDefinitions))
	for _, definition := range metricDefinitions {
		gauge, err := meter.Float64ObservableGauge(
			definition.Name,
			otelmetric.WithUnit(definition.Unit),
			otelmetric.WithDescription(definition.Description),
		)
		if err != nil {
			return nil, err
		}
		exporterMetrics.gauges[definition.Name] = gauge
		instruments = append(instruments, gauge)
	}
	if _, err := meter.RegisterCallback(exporterMetrics.observe, instruments...); err != nil {
		return nil, err
	}
	exporterMetrics.handler = promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	return exporterMetrics, nil
}

func (exporter *MetricsExporter) Handler() http.Handler {
	return exporter.handler
}

// observe exports only what the gateway itself owns: the phones' pushed status
// and the GitHub sync state. Host, VM, and agent series come from the unified
// host probes, which Prometheus scrapes directly.
func (exporter *MetricsExporter) observe(_ context.Context, observer otelmetric.Observer) error {
	snapshot := exporter.store.Snapshot()
	for _, mobile := range freshMobiles(snapshot.Mobiles, time.Now().UTC()) {
		exporter.observeMobile(observer, mobile)
	}
	exporter.observeGitHub(observer, snapshot.GitHub)
	return nil
}

func (exporter *MetricsExporter) gauge(observer otelmetric.Observer, name string, value float64, labels map[string]string) {
	gauge, ok := exporter.gauges[name]
	if !ok {
		return
	}
	observer.ObserveFloat64(gauge, value, attributeOption(labels))
}

func (exporter *MetricsExporter) state(observer otelmetric.Observer, name string, base map[string]string, stateKey string, current string, states []string) {
	seen := false
	for _, state := range states {
		if state == current {
			seen = true
		}
		exporter.gauge(observer, name, boolFloat(current == state), mergeLabels(base, stateKey, state))
	}
	if current != "" && !seen {
		exporter.gauge(observer, name, 1, mergeLabels(base, stateKey, current))
	}
}

func (exporter *MetricsExporter) observeMobile(observer otelmetric.Observer, mobile *mev1.MobileState) {
	deviceID := mobile.GetDeviceUid()
	update := float64(unixSeconds(mobile.GetUpdateTime()))
	exporter.gauge(observer, "realtime_device_last_update_time_seconds", update, deviceLabels(deviceID, "phone"))
	if phone := mobile.GetPhone(); phone != nil {
		if phone.BatteryPercent != nil {
			exporter.gauge(observer, "realtime_device_battery_level_ratio", float64(phone.GetBatteryPercent())/100, deviceLabels(deviceID, "phone"))
		}
		if phone.GetChargeState() != mev1.ChargeState_CHARGE_STATE_UNSPECIFIED {
			exporter.gauge(observer, "realtime_device_charging", boolFloat(phone.GetChargeState() == mev1.ChargeState_CHARGE_STATE_CHARGING), deviceLabels(deviceID, "phone"))
		}
		if phone.GetNetwork() != mev1.NetworkState_NETWORK_STATE_UNSPECIFIED {
			exporter.state(observer, "realtime_device_network_state", map[string]string{"device_id": deviceID}, "network_type", networkStateString(phone.GetNetwork()), []string{"offline", "wifi", "cellular", "vpn", "online", "unknown"})
		}
		exporter.observeAccessories(observer, deviceLabels(deviceID, "phone"), phone.GetAccessories())
	}

	watch := mobile.GetWatch()
	if watch == nil {
		return
	}
	state := watch.GetWatchState()
	exporter.gauge(observer, "realtime_device_last_update_time_seconds", update, deviceLabels(deviceID, "watch"))
	if state != nil {
		exporter.gauge(observer, "realtime_device_battery_level_ratio", float64(state.GetBatteryPercent())/100, deviceLabels(deviceID, "watch"))
		if state.GetChargeState() != mev1.ChargeState_CHARGE_STATE_UNSPECIFIED {
			exporter.gauge(observer, "realtime_device_charging", boolFloat(state.GetChargeState() == mev1.ChargeState_CHARGE_STATE_CHARGING), deviceLabels(deviceID, "watch"))
		}
	}
	if watch.GetHeartRate() != nil {
		exporter.gauge(observer, "realtime_watch_heart_rate_beats_per_minute", float64(watch.GetHeartRate().GetBeatsPerMinute()), map[string]string{"device_id": deviceID})
	}
	if watch.GetActivityTotals() != nil {
		exporter.gauge(observer, "realtime_watch_steps", float64(watch.GetActivityTotals().GetSteps()), map[string]string{"device_id": deviceID})
	}
}

func (exporter *MetricsExporter) observeAccessories(observer otelmetric.Observer, base map[string]string, accessories []*mev1.Accessory) {
	for _, accessory := range accessories {
		if accessory.GetKind() == "" || accessory.GetDisplayName() == "" {
			continue
		}
		labels := copyLabels(base)
		labels["accessory_kind"] = accessory.GetKind()
		labels["accessory_name"] = accessory.GetDisplayName()
		if accessory.GetModel() != "" {
			labels["accessory_model"] = accessory.GetModel()
		}
		exporter.gauge(observer, "realtime_device_accessory_connected", 1, labels)
		if accessory.BatteryPercent != nil {
			exporter.gauge(observer, "realtime_device_accessory_battery_level_ratio", float64(accessory.GetBatteryPercent())/100, labels)
		}
	}
}

func (exporter *MetricsExporter) observeGitHub(observer otelmetric.Observer, status *mev1.GithubSyncDetail) {
	exporter.state(observer, "realtime_github_status_sync_state", nil, "state", githubStateString(status.GetState()), []string{"disabled", "pending", "ok", "error"})
}

func attributeOption(labels map[string]string) otelmetric.MeasurementOption {
	pairs := make([]attribute.KeyValue, 0, len(labels))
	for key, value := range labels {
		pairs = append(pairs, attribute.String(key, value))
	}
	return otelmetric.WithAttributes(pairs...)
}

func mergeLabels(base map[string]string, key string, value string) map[string]string {
	labels := copyLabels(base)
	labels[key] = value
	return labels
}

func deviceLabels(deviceID string, deviceType string) map[string]string {
	return map[string]string{"device_id": deviceID, "device_type": deviceType}
}

func copyLabels(labels map[string]string) map[string]string {
	clone := make(map[string]string, len(labels))
	for key, value := range labels {
		clone[key] = value
	}
	return clone
}

func unixSeconds(value *timestamppb.Timestamp) int64 {
	if value == nil {
		return 0
	}
	return value.AsTime().Unix()
}

func boolFloat(value bool) float64 {
	if value {
		return 1
	}
	return 0
}
