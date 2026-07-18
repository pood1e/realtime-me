package gateway

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"

	mev1 "github.com/pood1e/realtime-me/gen/go/realtime/me/status/v1"
)

const (
	// maxMetricRangePoints bounds one response. It also bounds the work
	// Prometheus does per request, since a caller cannot widen the range without
	// widening the step to match.
	maxMetricRangePoints = 1500

	minMetricRangeStep = time.Second
	maxMetricRange     = 32 * 24 * time.Hour
)

// MetricsServer implements the Connect MetricsService. It is the only place a
// PromQL expression is written: callers name a MetricSeries and the gateway
// resolves it, so no client depends on a metric name, label, or job.
type MetricsServer struct {
	prometheus *PrometheusClient
}

func NewMetricsServer(prometheus *PrometheusClient) *MetricsServer {
	return &MetricsServer{prometheus: prometheus}
}

func (server *MetricsServer) GetMetricRange(
	ctx context.Context,
	request *connect.Request[mev1.GetMetricRangeRequest],
) (*connect.Response[mev1.GetMetricRangeResponse], error) {
	message := request.Msg
	start, end, step, err := metricRangeWindow(message)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	query, err := metricSeriesQuery(message)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	points, err := server.prometheus.QueryRange(ctx, query, start, end, step)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("prometheus unavailable"))
	}
	return connect.NewResponse(&mev1.GetMetricRangeResponse{Points: points}), nil
}

// metricRangeWindow validates the requested window and returns it. The point
// count is capped so one request cannot ask Prometheus for an unbounded scan.
func metricRangeWindow(request *mev1.GetMetricRangeRequest) (time.Time, time.Time, time.Duration, error) {
	if request.GetStartTime() == nil || request.GetEndTime() == nil || request.GetStep() == nil {
		return time.Time{}, time.Time{}, 0, errors.New("start_time, end_time, and step are required")
	}
	start := request.GetStartTime().AsTime()
	end := request.GetEndTime().AsTime()
	step := request.GetStep().AsDuration()

	if !end.After(start) {
		return time.Time{}, time.Time{}, 0, errors.New("end_time must be after start_time")
	}
	if end.Sub(start) > maxMetricRange {
		return time.Time{}, time.Time{}, 0, fmt.Errorf("range exceeds %s", maxMetricRange)
	}
	if step < minMetricRangeStep {
		return time.Time{}, time.Time{}, 0, fmt.Errorf("step must be at least %s", minMetricRangeStep)
	}
	// query_range answers with both endpoints, so a window spanning n steps is
	// n+1 points.
	if end.Sub(start)/step+1 > maxMetricRangePoints {
		return time.Time{}, time.Time{}, 0, fmt.Errorf("range and step yield more than %d points", maxMetricRangePoints)
	}
	return start, end, step, nil
}

// metricSeriesQuery resolves a series and its selectors into a PromQL
// expression. Every selector value is escaped, and an unknown series is refused
// rather than defaulted, so no caller can reach an unintended series.
func metricSeriesQuery(request *mev1.GetMetricRangeRequest) (string, error) {
	switch request.GetSeries() {
	case mev1.MetricSeries_METRIC_SERIES_HOST_CPU_UTILIZATION:
		return hostCPUQuery(request.GetDeviceUid())

	case mev1.MetricSeries_METRIC_SERIES_HOST_MEMORY_USAGE:
		return hostMemoryQuery(request.GetDeviceUid())

	case mev1.MetricSeries_METRIC_SERIES_HOST_FILESYSTEM_UTILIZATION:
		return hostFilesystemQuery(request.GetDeviceUid())

	case mev1.MetricSeries_METRIC_SERIES_PHONE_BATTERY_LEVEL:
		return deviceBatteryQuery(request.GetDeviceUid(), "phone")

	case mev1.MetricSeries_METRIC_SERIES_WATCH_BATTERY_LEVEL:
		return deviceBatteryQuery(request.GetDeviceUid(), "watch")

	case mev1.MetricSeries_METRIC_SERIES_WATCH_HEART_RATE:
		device, err := requireDeviceUID(request.GetDeviceUid())
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`realtime_watch_heart_rate_beats_per_minute{device_id=%s}`, device), nil

	case mev1.MetricSeries_METRIC_SERIES_WATCH_STEPS:
		device, err := requireDeviceUID(request.GetDeviceUid())
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`realtime_watch_steps{device_id=%s}`, device), nil

	case mev1.MetricSeries_METRIC_SERIES_ACCESSORY_BATTERY_LEVEL:
		return accessoryBatteryQuery(request)

	case mev1.MetricSeries_METRIC_SERIES_AGENT_BUDGET_REMAINING:
		kind := strings.TrimSpace(request.GetAgentKind())
		if kind == "" {
			return "", errors.New("agent_kind is required")
		}
		selectors := []string{"agent_kind=" + promLabelValue(kind)}
		if device := strings.TrimSpace(request.GetDeviceUid()); device != "" {
			selectors = append(selectors, "device_id="+promLabelValue(device))
		}
		return fmt.Sprintf(`realtime_agent_budget_remaining_ratio{%s} * 100`, strings.Join(selectors, ",")), nil

	default:
		return "", errors.New("unknown metric series")
	}
}

func hostCPUQuery(deviceUID string) (string, error) {
	instance := strings.TrimSpace(deviceUID)
	if instance == "" {
		return "", errors.New("device_uid is required")
	}
	if instance == "server" {
		return `100 * (1 - avg(rate(node_cpu_seconds_total{job="node-exporter",instance="server",mode="idle"}[2m])))`, nil
	}
	selector := probeHostSelector(instance)
	return fmt.Sprintf(`realtime_system_cpu_utilization_ratio{%s} * 100`, selector), nil
}

func hostMemoryQuery(deviceUID string) (string, error) {
	instance := strings.TrimSpace(deviceUID)
	if instance == "" {
		return "", errors.New("device_uid is required")
	}
	if instance == "server" {
		selector := `job="node-exporter",instance="server"`
		return fmt.Sprintf(`node_memory_MemTotal_bytes{%s} - node_memory_MemAvailable_bytes{%s}`, selector, selector), nil
	}
	selector := probeHostSelector(instance)
	return fmt.Sprintf(
		`realtime_system_memory_total_bytes{%s} - realtime_system_memory_available_bytes{%s}`,
		selector,
		selector,
	), nil
}

func hostFilesystemQuery(deviceUID string) (string, error) {
	instance := strings.TrimSpace(deviceUID)
	if instance == "" {
		return "", errors.New("device_uid is required")
	}
	if instance == "server" {
		selector := `job="node-exporter",instance="server",mountpoint="/",fstype!~"tmpfs|overlay|squashfs"`
		return fmt.Sprintf(`100 * (1 - node_filesystem_avail_bytes{%s} / node_filesystem_size_bytes{%s})`, selector, selector), nil
	}
	selector := probeHostSelector(instance)
	return fmt.Sprintf(
		`100 * (1 - realtime_system_filesystem_available_bytes{%s} / realtime_system_filesystem_total_bytes{%s})`,
		selector,
		selector,
	), nil
}

func probeHostSelector(instance string) string {
	return probeJobSelector + `,instance=` + promLabelValue(instance)
}

func deviceBatteryQuery(deviceUID string, deviceType string) (string, error) {
	device, err := requireDeviceUID(deviceUID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`realtime_device_battery_level_ratio{device_id=%s,device_type=%s} * 100`, device, promLabelValue(deviceType)), nil
}

func accessoryBatteryQuery(request *mev1.GetMetricRangeRequest) (string, error) {
	device, err := requireDeviceUID(request.GetDeviceUid())
	if err != nil {
		return "", err
	}
	accessory := request.GetAccessory()
	kind := strings.TrimSpace(accessory.GetKind())
	name := strings.TrimSpace(accessory.GetDisplayName())
	if kind == "" || name == "" {
		return "", errors.New("accessory kind and display_name are required")
	}
	return fmt.Sprintf(
		`realtime_device_accessory_battery_level_ratio{device_id=%s,accessory_kind=%s,accessory_name=%s} * 100`,
		device, promLabelValue(kind), promLabelValue(name),
	), nil
}

func requireDeviceUID(deviceUID string) (string, error) {
	device := strings.TrimSpace(deviceUID)
	if device == "" {
		return "", errors.New("device_uid is required")
	}
	return promLabelValue(device), nil
}

// promLabelValue renders a PromQL label value, escaping it so a selector can
// never terminate the quoted string and inject an expression.
func promLabelValue(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`)
	return `"` + replacer.Replace(value) + `"`
}
