package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/netip"
	"net/url"
	"sort"
	"strconv"
	"strings"
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
	media := client.DeviceMediaStatuses(ctx)
	accessories := client.DeviceAccessoryStatuses(ctx)

	return DeviceStatus{
		DeviceID:    "server",
		Kind:        "host",
		Role:        "server",
		State:       onlineState(up),
		UpdatedAt:   utcTime(),
		Metrics:     systemMetrics(cpuCores, cpuUsage, memoryTotal, memoryAvailable, diskTotal, diskAvailable),
		Media:       media["server"],
		Accessories: accessories["server"],
	}
}

func (client *PrometheusClient) NodeExporterStatuses(ctx context.Context) []DeviceStatus {
	up := client.queryVector(ctx, `up{job=~"node-exporter|vm-node-exporter",instance!="server"}`)
	if len(up) == 0 {
		return nil
	}
	cpuUsage := samplesByInstance(client.queryVector(ctx, `1 - avg by (instance) (rate(node_cpu_seconds_total{job=~"node-exporter|vm-node-exporter",instance!="server",mode="idle"}[2m]))`))
	cpuCores := samplesByInstance(client.queryVector(ctx, `count by (instance) (count by (instance, cpu) (node_cpu_seconds_total{job=~"node-exporter|vm-node-exporter",instance!="server",mode="idle"}))`))
	memoryTotal := samplesByInstance(client.queryVector(ctx, `node_memory_MemTotal_bytes{job=~"node-exporter|vm-node-exporter",instance!="server"}`))
	memoryAvailable := samplesByInstance(client.queryVector(ctx, `node_memory_MemAvailable_bytes{job=~"node-exporter|vm-node-exporter",instance!="server"}`))
	diskTotal := samplesByInstance(client.queryVector(ctx, `node_filesystem_size_bytes{job=~"node-exporter|vm-node-exporter",instance!="server",mountpoint="/",fstype!~"tmpfs|overlay|squashfs"}`))
	diskAvailable := samplesByInstance(client.queryVector(ctx, `node_filesystem_avail_bytes{job=~"node-exporter|vm-node-exporter",instance!="server",mountpoint="/",fstype!~"tmpfs|overlay|squashfs"}`))
	models := nodeModels(client.queryVector(ctx, `node_os_info{job=~"node-exporter|vm-node-exporter",instance!="server"}`), client.queryVector(ctx, `node_uname_info{job=~"node-exporter|vm-node-exporter",instance!="server"}`))
	media := client.DeviceMediaStatuses(ctx)
	accessories := client.DeviceAccessoryStatuses(ctx)

	devices := make([]DeviceStatus, 0, len(up))
	for _, sample := range up {
		instance := sample.Metric["instance"]
		if instance == "" {
			continue
		}
		deviceID, deviceName, ok := prometheusDeviceIdentity(sample.Metric)
		if !ok {
			continue
		}
		kind := firstNonEmpty(sample.Metric["device_kind"], "host")
		role := firstNonEmpty(sample.Metric["device_role"], "desktop")
		if sample.Metric["job"] == "vm-node-exporter" {
			kind = firstNonEmpty(sample.Metric["device_kind"], "virtual_machine")
			role = firstNonEmpty(sample.Metric["device_role"], "vm")
		}
		devices = append(devices, DeviceStatus{
			DeviceID:    deviceID,
			DeviceName:  deviceName,
			DeviceModel: firstNonEmpty(sample.Metric["device_model"], models[instance]),
			Kind:        kind,
			Role:        role,
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
			Media:       media[deviceID],
			Accessories: accessories[deviceID],
		})
	}
	sort.Slice(devices, func(left int, right int) bool {
		return devices[left].DeviceID < devices[right].DeviceID
	})
	return devices
}

func (client *PrometheusClient) DeviceMediaStatuses(ctx context.Context) map[string]*MediaStatus {
	samples := client.queryVector(ctx, `realtime_device_media_playing{job="device-exporter"} == 1`)
	statuses := make(map[string]*MediaStatus, len(samples))
	for _, sample := range samples {
		title := sample.Metric["title"]
		if title == "" {
			continue
		}
		deviceID := firstNonEmpty(sample.Metric["device_id"], sample.Metric["instance"])
		if deviceID == "" {
			continue
		}
		statuses[deviceID] = &MediaStatus{
			Title:  title,
			Artist: sample.Metric["artist"],
		}
	}
	return statuses
}

func (client *PrometheusClient) DeviceAccessoryStatuses(ctx context.Context) map[string][]AccessoryStatus {
	connected := client.queryVector(ctx, `realtime_device_accessory_connected{job="device-exporter"} == 1`)
	batteries := client.queryVector(ctx, `realtime_device_accessory_battery_level_ratio{job="device-exporter"}`)
	statuses := make(map[string][]AccessoryStatus, len(connected))
	byKey := make(map[string]AccessoryStatus, len(connected)+len(batteries))

	for _, sample := range connected {
		deviceID, accessory, ok := accessoryFromMetric(sample.Metric)
		if !ok {
			continue
		}
		key := deviceID + "\x00" + accessoryKey(accessory)
		byKey[key] = accessory
	}
	for _, sample := range batteries {
		deviceID, accessory, ok := accessoryFromMetric(sample.Metric)
		if !ok {
			continue
		}
		percent := int(math.Round(clampRatio(sample.Value) * 100))
		accessory.BatteryPercent = &percent
		key := deviceID + "\x00" + accessoryKey(accessory)
		if existing, exists := byKey[key]; exists {
			existing.BatteryPercent = &percent
			byKey[key] = existing
		} else {
			byKey[key] = accessory
		}
	}
	for key, accessory := range byKey {
		deviceID, _, ok := strings.Cut(key, "\x00")
		if !ok || deviceID == "" {
			continue
		}
		statuses[deviceID] = append(statuses[deviceID], accessory)
	}
	for deviceID := range statuses {
		sortAccessories(statuses[deviceID])
	}
	return statuses
}

func (client *PrometheusClient) AgentStatuses(ctx context.Context) []StoredAgentStatus {
	running := client.queryVector(ctx, `realtime_agent_state{job="agent-exporter",state="running"} == 1`)
	if len(running) == 0 {
		return nil
	}
	budgets := samplesByAgent(client.queryVector(ctx, `realtime_agent_budget_remaining_ratio{job="agent-exporter"}`))
	now := utcTime()
	agents := make([]StoredAgentStatus, 0, len(running))
	for _, sample := range running {
		agentID := sample.Metric["agent_id"]
		if agentID == "" {
			continue
		}
		input := AgentIngest{
			AgentID:    agentID,
			DeviceID:   sample.Metric["device_id"],
			DeviceName: sample.Metric["device_name"],
			UpdatedAt:  now,
			State:      "running",
		}
		if value, ok := budgets[agentKey(input)]; ok {
			percent := int(math.Round(clampRatio(value) * 100))
			input.BudgetRemainingPercent = &percent
		}
		agents = append(agents, StoredAgentStatus{
			AgentIngest: input,
			ReceivedAt:  now,
		})
	}
	sort.Slice(agents, func(left int, right int) bool {
		if agents[left].DeviceID != agents[right].DeviceID {
			return agents[left].DeviceID < agents[right].DeviceID
		}
		return agents[left].AgentID < agents[right].AgentID
	})
	return agents
}

func accessoryFromMetric(labels map[string]string) (string, AccessoryStatus, bool) {
	deviceID := firstNonEmpty(labels["device_id"], labels["instance"])
	kind := labels["accessory_kind"]
	name := labels["accessory_name"]
	if deviceID == "" || kind == "" || name == "" {
		return "", AccessoryStatus{}, false
	}
	return deviceID, AccessoryStatus{
		Kind:  kind,
		Name:  name,
		Model: labels["accessory_model"],
	}, true
}

func accessoryKey(accessory AccessoryStatus) string {
	return accessory.Kind + "\x00" + accessory.Name + "\x00" + accessory.Model
}

func sortAccessories(accessories []AccessoryStatus) {
	sort.Slice(accessories, func(left int, right int) bool {
		if accessories[left].Kind != accessories[right].Kind {
			return accessories[left].Kind < accessories[right].Kind
		}
		return accessories[left].Name < accessories[right].Name
	})
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

func samplesByAgent(samples []prometheusSample) map[string]float64 {
	values := make(map[string]float64, len(samples))
	for _, sample := range samples {
		agent := AgentIngest{
			AgentID:  sample.Metric["agent_id"],
			DeviceID: sample.Metric["device_id"],
		}
		if agent.AgentID == "" {
			continue
		}
		values[agentKey(agent)] = sample.Value
	}
	return values
}

func agentKey(agent AgentIngest) string {
	if agent.DeviceID == "" {
		return agent.AgentID
	}
	return agent.DeviceID + "/" + agent.AgentID
}

func nodeModels(osInfo []prometheusSample, unameInfo []prometheusSample) map[string]string {
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

func prometheusDeviceIdentity(labels map[string]string) (string, string, bool) {
	deviceID := firstPublicNetworkLabel(labels["device_id"], labels["device_name"], labels["instance"])
	if deviceID == "" {
		return "", "", false
	}
	deviceName := firstPublicNetworkLabel(labels["device_name"], deviceID)
	return deviceID, deviceName, true
}

func firstPublicNetworkLabel(values ...string) string {
	for _, value := range values {
		label := strings.TrimSpace(value)
		if label != "" && !isPrivateNetworkLabel(label) {
			return label
		}
	}
	return ""
}

func isPrivateNetworkLabel(value string) bool {
	address, err := netip.ParseAddr(value)
	return err == nil && (address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast())
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
