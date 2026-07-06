package gateway

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	metricSystemCPULogicalCount    = "system.cpu.logical.count"
	metricSystemCPUUtilization     = "system.cpu.utilization"
	metricSystemMemoryUsage        = "system.memory.usage"
	metricSystemMemoryLimit        = "system.memory.limit"
	metricSystemFilesystemUsage    = "system.filesystem.usage"
	metricSystemFilesystemLimit    = "system.filesystem.limit"
	metricSystemFilesystemUsagePct = "system.filesystem.utilization"
)

type PrometheusClient struct {
	baseURL string
	client  *http.Client
}

func NewPrometheusClient(baseURL string) *PrometheusClient {
	return &PrometheusClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (client *PrometheusClient) ServerStatus(ctx context.Context) DeviceStatus {
	up := client.queryScalar(ctx, `max(up{job="node-exporter"})`)
	cpuUsage := client.queryRatio(ctx, `1 - avg(rate(node_cpu_seconds_total{job="node-exporter",mode="idle"}[2m]))`)
	cpuCores := client.queryScalar(ctx, `count(count by (cpu) (node_cpu_seconds_total{job="node-exporter",mode="idle"}))`)
	memoryTotal := client.queryScalar(ctx, `node_memory_MemTotal_bytes{job="node-exporter"}`)
	memoryAvailable := client.queryScalar(ctx, `node_memory_MemAvailable_bytes{job="node-exporter"}`)
	diskTotal := client.queryScalar(ctx, `node_filesystem_size_bytes{job="node-exporter",mountpoint="/",fstype!~"tmpfs|overlay|squashfs"}`)
	diskAvailable := client.queryScalar(ctx, `node_filesystem_avail_bytes{job="node-exporter",mountpoint="/",fstype!~"tmpfs|overlay|squashfs"}`)

	metrics := make([]MetricSample, 0, 7)
	metrics = appendMetric(metrics, metricSystemCPULogicalCount, "{cpu}", cpuCores, nil)
	metrics = appendMetric(metrics, metricSystemCPUUtilization, "1", cpuUsage, nil)
	metrics = appendMetric(metrics, metricSystemMemoryUsage, "By", subtract(memoryTotal, memoryAvailable), map[string]string{"system.memory.state": "used"})
	metrics = appendMetric(metrics, metricSystemMemoryLimit, "By", memoryTotal, nil)
	metrics = appendMetric(metrics, metricSystemFilesystemUsage, "By", subtract(diskTotal, diskAvailable), map[string]string{"mountpoint": "/"})
	metrics = appendMetric(metrics, metricSystemFilesystemLimit, "By", diskTotal, map[string]string{"mountpoint": "/"})
	metrics = appendMetric(metrics, metricSystemFilesystemUsagePct, "1", ratio(subtract(diskTotal, diskAvailable), diskTotal), map[string]string{"mountpoint": "/"})

	return DeviceStatus{
		DeviceID:  "server",
		Kind:      "host",
		Role:      "server",
		State:     onlineState(up),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Metrics:   metrics,
	}
}

func (client *PrometheusClient) queryRatio(ctx context.Context, query string) *float64 {
	value := client.queryScalar(ctx, query)
	if value == nil {
		return nil
	}
	clamped := math.Max(0, math.Min(1, *value))
	return &clamped
}

func (client *PrometheusClient) queryScalar(ctx context.Context, query string) *float64 {
	endpoint, err := url.Parse(client.baseURL + "/api/v1/query")
	if err != nil {
		return nil
	}
	values := endpoint.Query()
	values.Set("query", query)
	endpoint.RawQuery = values.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil
	}
	response, err := client.client.Do(request)
	if err != nil {
		return nil
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil
	}

	var payload prometheusQueryResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil
	}
	if len(payload.Data.Result) == 0 || len(payload.Data.Result[0].Value) < 2 {
		return nil
	}
	value, ok := payload.Data.Result[0].Value[1].(string)
	if !ok {
		return nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

type prometheusQueryResponse struct {
	Data struct {
		Result []struct {
			Value []any `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func appendMetric(metrics []MetricSample, name string, unit string, value *float64, attributes map[string]string) []MetricSample {
	if value == nil {
		return metrics
	}
	return append(metrics, MetricSample{
		Name:       name,
		Unit:       unit,
		Value:      roundMetric(*value),
		Attributes: attributes,
	})
}

func subtract(total *float64, available *float64) *float64 {
	if total == nil || available == nil {
		return nil
	}
	used := math.Max(0, *total-*available)
	return &used
}

func ratio(value *float64, total *float64) *float64 {
	if value == nil || total == nil || *total <= 0 {
		return nil
	}
	ratio := math.Max(0, math.Min(1, *value / *total))
	return &ratio
}

func onlineState(up *float64) string {
	if up != nil && *up > 0 {
		return "online"
	}
	return "offline"
}

func roundMetric(value float64) float64 {
	return math.Round(value*1000) / 1000
}
