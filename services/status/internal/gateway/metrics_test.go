package gateway

import (
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	mev1 "github.com/pood1e/realtime-me/gen/go/realtime/me/status/v1"
)

func TestMetricSeriesQueryResolvesEverySeries(t *testing.T) {
	cases := []struct {
		name    string
		request *mev1.GetMetricRangeRequest
		want    string
	}{
		{
			name:    "host cpu",
			request: &mev1.GetMetricRangeRequest{Series: mev1.MetricSeries_METRIC_SERIES_HOST_CPU_UTILIZATION, DeviceUid: "dev_a"},
			want:    `realtime_system_cpu_utilization_ratio{job="probe-agent",instance="dev_a"} * 100`,
		},
		{
			name:    "the server uses its static instance label",
			request: &mev1.GetMetricRangeRequest{Series: mev1.MetricSeries_METRIC_SERIES_HOST_CPU_UTILIZATION, DeviceUid: "server"},
			want:    `100 * (1 - avg(rate(node_cpu_seconds_total{job="node-exporter",instance="server",mode="idle"}[2m])))`,
		},
		{
			name:    "phone battery",
			request: &mev1.GetMetricRangeRequest{Series: mev1.MetricSeries_METRIC_SERIES_PHONE_BATTERY_LEVEL, DeviceUid: "dev_a"},
			want:    `realtime_device_battery_level_ratio{device_id="dev_a",device_type="phone"} * 100`,
		},
		{
			name:    "watch battery",
			request: &mev1.GetMetricRangeRequest{Series: mev1.MetricSeries_METRIC_SERIES_WATCH_BATTERY_LEVEL, DeviceUid: "dev_a"},
			want:    `realtime_device_battery_level_ratio{device_id="dev_a",device_type="watch"} * 100`,
		},
		{
			name:    "watch heart rate",
			request: &mev1.GetMetricRangeRequest{Series: mev1.MetricSeries_METRIC_SERIES_WATCH_HEART_RATE, DeviceUid: "dev_a"},
			want:    `realtime_watch_heart_rate_beats_per_minute{device_id="dev_a"}`,
		},
		{
			name:    "watch steps",
			request: &mev1.GetMetricRangeRequest{Series: mev1.MetricSeries_METRIC_SERIES_WATCH_STEPS, DeviceUid: "dev_a"},
			want:    `realtime_watch_steps{device_id="dev_a"}`,
		},
		{
			name: "accessory battery",
			request: &mev1.GetMetricRangeRequest{
				Series:    mev1.MetricSeries_METRIC_SERIES_ACCESSORY_BATTERY_LEVEL,
				DeviceUid: "dev_a",
				Accessory: &mev1.AccessorySelector{Kind: "bluetooth_audio", DisplayName: "AirPods Pro"},
			},
			want: `realtime_device_accessory_battery_level_ratio{device_id="dev_a",accessory_kind="bluetooth_audio",accessory_name="AirPods Pro"} * 100`,
		},
		{
			// The exporter labels this series agent_kind. The old frontend query
			// used agent_id, so this chart never matched a single sample.
			name:    "agent budget uses agent_kind",
			request: &mev1.GetMetricRangeRequest{Series: mev1.MetricSeries_METRIC_SERIES_AGENT_BUDGET_REMAINING, AgentKind: "claude", DeviceUid: "dev_a"},
			want:    `realtime_agent_budget_remaining_ratio{agent_kind="claude",device_id="dev_a"} * 100`,
		},
		{
			name:    "agent budget without a device",
			request: &mev1.GetMetricRangeRequest{Series: mev1.MetricSeries_METRIC_SERIES_AGENT_BUDGET_REMAINING, AgentKind: "codex"},
			want:    `realtime_agent_budget_remaining_ratio{agent_kind="codex"} * 100`,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := metricSeriesQuery(testCase.request)
			if err != nil {
				t.Fatalf("metricSeriesQuery: %v", err)
			}
			if got != testCase.want {
				t.Errorf("\n got: %s\nwant: %s", got, testCase.want)
			}
		})
	}
}

// An unknown or unspecified series is refused, never defaulted onto some other
// series a caller did not ask for.
func TestMetricSeriesQueryRefusesUnknownSeries(t *testing.T) {
	for _, series := range []mev1.MetricSeries{mev1.MetricSeries_METRIC_SERIES_UNSPECIFIED, mev1.MetricSeries(9999)} {
		if _, err := metricSeriesQuery(&mev1.GetMetricRangeRequest{Series: series, DeviceUid: "dev_a"}); err == nil {
			t.Errorf("series %v must be refused", series)
		}
	}
}

func TestMetricSeriesQueryRequiresItsSelectors(t *testing.T) {
	cases := map[string]*mev1.GetMetricRangeRequest{
		"host without device":      {Series: mev1.MetricSeries_METRIC_SERIES_HOST_CPU_UTILIZATION},
		"watch without device":     {Series: mev1.MetricSeries_METRIC_SERIES_WATCH_HEART_RATE},
		"agent without kind":       {Series: mev1.MetricSeries_METRIC_SERIES_AGENT_BUDGET_REMAINING, DeviceUid: "dev_a"},
		"accessory without device": {Series: mev1.MetricSeries_METRIC_SERIES_ACCESSORY_BATTERY_LEVEL, Accessory: &mev1.AccessorySelector{Kind: "k", DisplayName: "n"}},
		"accessory without name":   {Series: mev1.MetricSeries_METRIC_SERIES_ACCESSORY_BATTERY_LEVEL, DeviceUid: "dev_a", Accessory: &mev1.AccessorySelector{Kind: "k"}},
		"accessory absent":         {Series: mev1.MetricSeries_METRIC_SERIES_ACCESSORY_BATTERY_LEVEL, DeviceUid: "dev_a"},
	}
	for name, request := range cases {
		if _, err := metricSeriesQuery(request); err == nil {
			t.Errorf("%s: expected rejection", name)
		}
	}
}

// A selector must never terminate the quoted label value and graft on an
// expression of the caller's choosing.
func TestPromLabelValueEscapesInjection(t *testing.T) {
	hostile := `dev_a"} or up{job="prometheus`
	query, err := metricSeriesQuery(&mev1.GetMetricRangeRequest{
		Series:    mev1.MetricSeries_METRIC_SERIES_WATCH_STEPS,
		DeviceUid: hostile,
	})
	if err != nil {
		t.Fatalf("metricSeriesQuery: %v", err)
	}
	if strings.Contains(query, `or up{`) && !strings.Contains(query, `\"`) {
		t.Fatalf("selector escaped its label value: %s", query)
	}
	want := `realtime_watch_steps{device_id="dev_a\"} or up{job=\"prometheus"}`
	if query != want {
		t.Errorf("\n got: %s\nwant: %s", query, want)
	}

	if got := promLabelValue(`back\slash`); got != `"back\\slash"` {
		t.Errorf("backslash not escaped: %s", got)
	}
	if got := promLabelValue("new\nline"); got != `"new\nline"` {
		t.Errorf("newline not escaped: %s", got)
	}
}

// The window is bounded so a single request cannot ask Prometheus for an
// unbounded scan. This keeps an authenticated Console request bounded.
func TestMetricRangeWindowBoundsTheQuery(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	window := func(start time.Time, end time.Time, step time.Duration) *mev1.GetMetricRangeRequest {
		return &mev1.GetMetricRangeRequest{
			StartTime: timestamppb.New(start),
			EndTime:   timestamppb.New(end),
			Step:      durationpb.New(step),
		}
	}

	if _, _, _, err := metricRangeWindow(window(now.Add(-time.Hour), now, 30*time.Second)); err != nil {
		t.Fatalf("a reasonable window must be accepted: %v", err)
	}

	rejected := map[string]*mev1.GetMetricRangeRequest{
		"end before start":  window(now, now.Add(-time.Hour), time.Minute),
		"zero-length range": window(now, now, time.Minute),
		"range too wide":    window(now.Add(-64*24*time.Hour), now, time.Minute),
		"step too small":    window(now.Add(-time.Hour), now, time.Millisecond),
		"too many points":   window(now.Add(-24*time.Hour), now, time.Second),
		"missing fields":    {Series: mev1.MetricSeries_METRIC_SERIES_WATCH_STEPS},
	}
	for name, request := range rejected {
		if _, _, _, err := metricRangeWindow(request); err == nil {
			t.Errorf("%s: expected rejection", name)
		}
	}
}

func TestMetricPointsSkipsMalformedSamples(t *testing.T) {
	points := metricPoints([][]any{
		{float64(1_700_000_000), "12.5"},
		{float64(1_700_000_030), "NaN"},
		{float64(1_700_000_060), "not a number"},
		{float64(1_700_000_090)},
		{"not a timestamp", "1"},
		{float64(1_700_000_120), "+Inf"},
		{float64(1_700_000_150), "7"},
	})
	if len(points) != 2 {
		t.Fatalf("got %d points, want 2 (only the two finite samples)", len(points))
	}
	if points[0].GetValue() != 12.5 || points[1].GetValue() != 7 {
		t.Errorf("unexpected values: %v, %v", points[0].GetValue(), points[1].GetValue())
	}
	if got := points[0].GetTime().AsTime().Unix(); got != 1_700_000_000 {
		t.Errorf("timestamp = %d", got)
	}
}

func TestHostMemoryQueryUsesUnifiedProbeMetrics(t *testing.T) {
	query, err := metricSeriesQuery(&mev1.GetMetricRangeRequest{
		Series:    mev1.MetricSeries_METRIC_SERIES_HOST_MEMORY_USAGE,
		DeviceUid: "dev_aaaa",
	})
	if err != nil {
		t.Fatalf("metricSeriesQuery: %v", err)
	}
	want := `realtime_system_memory_total_bytes{job="probe-agent",instance="dev_aaaa"} - realtime_system_memory_available_bytes{job="probe-agent",instance="dev_aaaa"}`
	if query != want {
		t.Errorf("\n got: %s\nwant: %s", query, want)
	}
}
