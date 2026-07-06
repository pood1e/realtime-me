package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
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
	maxPrometheusProxyResponseSize = 4 * 1024 * 1024
)

type PrometheusClient struct {
	baseURL string
	client  *http.Client
}

type prometheusSample struct {
	Metric map[string]string
	Value  float64
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
	up := client.queryScalar(ctx, `max(up{job="node-exporter",instance="server"})`)
	cpuUsage := client.queryRatio(ctx, `1 - avg(rate(node_cpu_seconds_total{job="node-exporter",instance="server",mode="idle"}[2m]))`)
	cpuCores := client.queryScalar(ctx, `count(count by (cpu) (node_cpu_seconds_total{job="node-exporter",instance="server",mode="idle"}))`)
	memoryTotal := client.queryScalar(ctx, `node_memory_MemTotal_bytes{job="node-exporter",instance="server"}`)
	memoryAvailable := client.queryScalar(ctx, `node_memory_MemAvailable_bytes{job="node-exporter",instance="server"}`)
	diskTotal := client.queryScalar(ctx, `node_filesystem_size_bytes{job="node-exporter",instance="server",mountpoint="/",fstype!~"tmpfs|overlay|squashfs"}`)
	diskAvailable := client.queryScalar(ctx, `node_filesystem_avail_bytes{job="node-exporter",instance="server",mountpoint="/",fstype!~"tmpfs|overlay|squashfs"}`)

	return DeviceStatus{
		DeviceID:  "server",
		Kind:      "host",
		Role:      "server",
		State:     onlineState(up),
		UpdatedAt: utcTime(),
		Metrics:   systemMetrics(cpuCores, cpuUsage, memoryTotal, memoryAvailable, diskTotal, diskAvailable),
	}
}

func (client *PrometheusClient) VirtualMachineStatuses(ctx context.Context) []DeviceStatus {
	up := client.queryVector(ctx, `up{job="vm-node-exporter"}`)
	if len(up) == 0 {
		return nil
	}
	cpuUsage := samplesByInstance(client.queryVector(ctx, `1 - avg by (instance) (rate(node_cpu_seconds_total{job="vm-node-exporter",mode="idle"}[2m]))`))
	cpuCores := samplesByInstance(client.queryVector(ctx, `count by (instance) (count by (instance, cpu) (node_cpu_seconds_total{job="vm-node-exporter",mode="idle"}))`))
	memoryTotal := samplesByInstance(client.queryVector(ctx, `node_memory_MemTotal_bytes{job="vm-node-exporter"}`))
	memoryAvailable := samplesByInstance(client.queryVector(ctx, `node_memory_MemAvailable_bytes{job="vm-node-exporter"}`))
	diskTotal := samplesByInstance(client.queryVector(ctx, `node_filesystem_size_bytes{job="vm-node-exporter",mountpoint="/",fstype!~"tmpfs|overlay|squashfs"}`))
	diskAvailable := samplesByInstance(client.queryVector(ctx, `node_filesystem_avail_bytes{job="vm-node-exporter",mountpoint="/",fstype!~"tmpfs|overlay|squashfs"}`))
	models := vmModels(client.queryVector(ctx, `node_os_info{job="vm-node-exporter"}`), client.queryVector(ctx, `node_uname_info{job="vm-node-exporter"}`))

	devices := make([]DeviceStatus, 0, len(up))
	for _, sample := range up {
		instance := sample.Metric["instance"]
		if instance == "" {
			continue
		}
		devices = append(devices, DeviceStatus{
			DeviceID:    instance,
			DeviceName:  instance,
			DeviceModel: models[instance],
			Kind:        "virtual_machine",
			Role:        "vm",
			State:       onlineState(&sample.Value),
			UpdatedAt:   utcTime(),
			Metrics: systemMetrics(
				cpuCores[instance],
				cpuUsage[instance],
				memoryTotal[instance],
				memoryAvailable[instance],
				diskTotal[instance],
				diskAvailable[instance],
			),
		})
	}
	sort.Slice(devices, func(left int, right int) bool {
		return devices[left].DeviceID < devices[right].DeviceID
	})
	return devices
}

func (client *PrometheusClient) Proxy(ctx context.Context, path string, values url.Values) ([]byte, int, error) {
	endpoint, err := url.Parse(client.baseURL + path)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}
	endpoint.RawQuery = values.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}
	request.Header.Set("Accept", "application/json")

	response, err := client.client.Do(request)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, maxPrometheusProxyResponseSize+1))
	if err != nil {
		return nil, http.StatusBadGateway, err
	}
	if len(body) > maxPrometheusProxyResponseSize {
		return nil, http.StatusBadGateway, fmt.Errorf("prometheus response exceeded %d bytes", maxPrometheusProxyResponseSize)
	}
	return body, response.StatusCode, nil
}

func (client *PrometheusClient) queryRatio(ctx context.Context, query string) *float64 {
	value := client.queryScalar(ctx, query)
	if value == nil {
		return nil
	}
	clamped := clampRatio(*value)
	return &clamped
}

func (client *PrometheusClient) queryScalar(ctx context.Context, query string) *float64 {
	samples := client.queryVector(ctx, query)
	if len(samples) == 0 {
		return nil
	}
	return &samples[0].Value
}

func (client *PrometheusClient) queryVector(ctx context.Context, query string) []prometheusSample {
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

	samples := make([]prometheusSample, 0, len(payload.Data.Result))
	for _, result := range payload.Data.Result {
		if len(result.Value) < 2 {
			continue
		}
		value, ok := result.Value[1].(string)
		if !ok {
			continue
		}
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			continue
		}
		samples = append(samples, prometheusSample{Metric: result.Metric, Value: parsed})
	}
	return samples
}

type prometheusQueryResponse struct {
	Data struct {
		Result []struct {
			Metric map[string]string `json:"metric"`
			Value  []any             `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func systemMetrics(cpuCores *float64, cpuUsage *float64, memoryTotal *float64, memoryAvailable *float64, diskTotal *float64, diskAvailable *float64) []MetricSample {
	metrics := make([]MetricSample, 0, 7)
	metrics = appendMetric(metrics, metricSystemCPULogicalCount, "{cpu}", cpuCores, nil)
	metrics = appendMetric(metrics, metricSystemCPUUtilization, "1", clampRatioPointer(cpuUsage), nil)
	metrics = appendMetric(metrics, metricSystemMemoryUsage, "By", subtract(memoryTotal, memoryAvailable), map[string]string{"system.memory.state": "used"})
	metrics = appendMetric(metrics, metricSystemMemoryLimit, "By", memoryTotal, nil)
	metrics = appendMetric(metrics, metricSystemFilesystemUsage, "By", subtract(diskTotal, diskAvailable), map[string]string{"mountpoint": "/"})
	metrics = appendMetric(metrics, metricSystemFilesystemLimit, "By", diskTotal, map[string]string{"mountpoint": "/"})
	metrics = appendMetric(metrics, metricSystemFilesystemUsagePct, "1", ratio(subtract(diskTotal, diskAvailable), diskTotal), map[string]string{"mountpoint": "/"})
	return metrics
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

func samplesByInstance(samples []prometheusSample) map[string]*float64 {
	values := make(map[string]*float64, len(samples))
	for _, sample := range samples {
		instance := sample.Metric["instance"]
		if instance == "" {
			continue
		}
		value := sample.Value
		values[instance] = &value
	}
	return values
}

func vmModels(osInfo []prometheusSample, unameInfo []prometheusSample) map[string]string {
	models := map[string]string{}
	for _, sample := range unameInfo {
		instance := sample.Metric["instance"]
		if instance == "" {
			continue
		}
		models[instance] = joinNonEmpty(" ", sample.Metric["sysname"], sample.Metric["release"], sample.Metric["machine"])
	}
	for _, sample := range osInfo {
		instance := sample.Metric["instance"]
		if instance == "" {
			continue
		}
		models[instance] = firstNonEmpty(sample.Metric["pretty_name"], sample.Metric["name"], models[instance])
	}
	return models
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
	ratio := clampRatio(*value / *total)
	return &ratio
}

func onlineState(up *float64) string {
	if up != nil && *up > 0 {
		return "online"
	}
	return "offline"
}

func clampRatioPointer(value *float64) *float64 {
	if value == nil {
		return nil
	}
	clamped := clampRatio(*value)
	return &clamped
}

func clampRatio(value float64) float64 {
	return math.Max(0, math.Min(1, value))
}

func roundMetric(value float64) float64 {
	return math.Round(value*1000) / 1000
}

func utcTime() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func joinNonEmpty(separator string, values ...string) string {
	result := ""
	for _, value := range values {
		if value == "" {
			continue
		}
		if result != "" {
			result += separator
		}
		result += value
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
