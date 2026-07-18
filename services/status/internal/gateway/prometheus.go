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
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	mev1 "github.com/pood1e/realtime-me/gen/go/realtime/me/status/v1"
)

const (
	metricSystemCPULogicalCount    = "system.cpu.logical.count"
	metricSystemCPUUtilization     = "system.cpu.utilization"
	metricSystemMemoryUsage        = "system.memory.usage"
	metricSystemMemoryLimit        = "system.memory.limit"
	metricSystemFilesystemUsage    = "system.filesystem.usage"
	metricSystemFilesystemLimit    = "system.filesystem.limit"
	metricSystemFilesystemUsagePct = "system.filesystem.utilization"
	maxPrometheusResponseSize      = 4 * 1024 * 1024

	// maxConcurrentPrometheusQueries sizes the idle-connection pool to the widest
	// concurrent fan-out a single status assembly performs.
	maxConcurrentPrometheusQueries = 10

	// maxAgentsPerSample bounds how many agents one exporter sample may claim to
	// be running. The count decides how many messages the unauthenticated status
	// assembly allocates, and the exporter that supplies it is a process on a
	// probe host rather than something this gateway controls.
	maxAgentsPerSample = 64
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
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// A status assembly issues its queries concurrently against one host; the
	// default of two idle connections would serialize them onto new sockets.
	transport.MaxIdleConnsPerHost = maxConcurrentPrometheusQueries
	return &PrometheusClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout:   3 * time.Second,
			Transport: transport,
		},
	}
}

// parallel runs every function concurrently and waits for all of them. The
// status document is assembled from many independent Prometheus queries; run
// serially they add up to more than the server's write timeout.
func parallel(functions ...func()) {
	var group sync.WaitGroup
	group.Add(len(functions))
	for _, function := range functions {
		go func() {
			defer group.Done()
			function()
		}()
	}
	group.Wait()
}

// ServerStatus describes the always-on server. Media and accessories are passed
// in because the whole fleet's are read in one pair of queries, not once per
// device group.
func (client *PrometheusClient) ServerStatus(ctx context.Context, media map[string]*mev1.MediaStatus, accessories map[string][]*mev1.Accessory) *mev1.DeviceState {
	var up, cpuCores, memoryTotal, memoryAvailable, diskTotal, diskAvailable *float64
	var cpuUsage *float64
	var osInfo, unameInfo []prometheusSample
	parallel(
		func() { up = client.queryScalar(ctx, `max(up{job="node-exporter",instance="server"})`) },
		func() {
			cpuUsage = client.queryRatio(ctx, `1 - avg(rate(node_cpu_seconds_total{job="node-exporter",instance="server",mode="idle"}[2m]))`)
		},
		func() {
			cpuCores = client.queryScalar(ctx, `count(count by (cpu) (node_cpu_seconds_total{job="node-exporter",instance="server",mode="idle"}))`)
		},
		func() {
			memoryTotal = client.queryScalar(ctx, `node_memory_MemTotal_bytes{job="node-exporter",instance="server"}`)
		},
		func() {
			memoryAvailable = client.queryScalar(ctx, `node_memory_MemAvailable_bytes{job="node-exporter",instance="server"}`)
		},
		func() {
			diskTotal = client.queryScalar(ctx, `node_filesystem_size_bytes{job="node-exporter",instance="server",mountpoint="/",fstype!~"tmpfs|overlay|squashfs"}`)
		},
		func() {
			diskAvailable = client.queryScalar(ctx, `node_filesystem_avail_bytes{job="node-exporter",instance="server",mountpoint="/",fstype!~"tmpfs|overlay|squashfs"}`)
		},
		func() { osInfo = client.queryVector(ctx, `node_os_info{job="node-exporter",instance="server"}`) },
		func() { unameInfo = client.queryVector(ctx, `node_uname_info{job="node-exporter",instance="server"}`) },
	)
	models := nodeModels(osInfo, unameInfo)

	return &mev1.DeviceState{
		DeviceUid:   "server",
		Kind:        mev1.DeviceKind_DEVICE_KIND_HOST,
		Role:        mev1.DeviceRole_DEVICE_ROLE_SERVER,
		State:       onlineState(up != nil && *up > 0),
		UpdateTime:  timestamppb.New(time.Now().UTC()),
		Model:       models["server"],
		Metrics:     systemMetrics(cpuCores, cpuUsage, memoryTotal, memoryAvailable, diskTotal, diskAvailable),
		Media:       media["server"],
		Accessories: accessories["server"],
	}
}

func (client *PrometheusClient) NodeExporterStatuses(ctx context.Context, media map[string]*mev1.MediaStatus, accessories map[string][]*mev1.Accessory) []*mev1.DeviceState {
	up := client.queryVector(ctx, `up{job=~"node-exporter|vm-node-exporter",instance!="server"}`)
	if len(up) == 0 {
		return nil
	}
	var cpuUsage, cpuCores, memoryTotal, memoryAvailable, diskTotal, diskAvailable map[string]*float64
	var osInfo, unameInfo []prometheusSample
	parallel(
		func() {
			cpuUsage = samplesByInstance(client.queryVector(ctx, `1 - avg by (instance) (rate(node_cpu_seconds_total{job=~"node-exporter|vm-node-exporter",instance!="server",mode="idle"}[2m]))`))
		},
		func() {
			cpuCores = samplesByInstance(client.queryVector(ctx, `count by (instance) (count by (instance, cpu) (node_cpu_seconds_total{job=~"node-exporter|vm-node-exporter",instance!="server",mode="idle"}))`))
		},
		// Linux node_exporter names come from /proc/meminfo; the darwin build uses
		// its own metric names, so fall back to those for macOS nodes. MemAvailable
		// has no darwin equivalent, so approximate it from reclaimable memory.
		func() {
			memoryTotal = samplesByInstance(client.queryVector(ctx, `node_memory_MemTotal_bytes{job=~"node-exporter|vm-node-exporter",instance!="server"} or node_memory_total_bytes{job=~"node-exporter|vm-node-exporter",instance!="server"}`))
		},
		func() {
			memoryAvailable = samplesByInstance(client.queryVector(ctx, `node_memory_MemAvailable_bytes{job=~"node-exporter|vm-node-exporter",instance!="server"} or (node_memory_free_bytes{job=~"node-exporter|vm-node-exporter",instance!="server"} + ignoring(__name__) node_memory_inactive_bytes{job=~"node-exporter|vm-node-exporter",instance!="server"})`))
		},
		func() {
			diskTotal = samplesByInstance(client.queryVector(ctx, `node_filesystem_size_bytes{job=~"node-exporter|vm-node-exporter",instance!="server",mountpoint="/",fstype!~"tmpfs|overlay|squashfs"}`))
		},
		func() {
			diskAvailable = samplesByInstance(client.queryVector(ctx, `node_filesystem_avail_bytes{job=~"node-exporter|vm-node-exporter",instance!="server",mountpoint="/",fstype!~"tmpfs|overlay|squashfs"}`))
		},
		func() {
			osInfo = client.queryVector(ctx, `node_os_info{job=~"node-exporter|vm-node-exporter",instance!="server"}`)
		},
		func() {
			unameInfo = client.queryVector(ctx, `node_uname_info{job=~"node-exporter|vm-node-exporter",instance!="server"}`)
		},
	)
	models := nodeModels(osInfo, unameInfo)

	devices := make([]*mev1.DeviceState, 0, len(up))
	for _, sample := range up {
		instance := sample.Metric["instance"]
		if instance == "" {
			continue
		}
		deviceUID, deviceName, ok := prometheusDeviceIdentity(sample.Metric)
		if !ok {
			continue
		}
		kind := parseDeviceKind(firstNonEmpty(sample.Metric["device_kind"], "host"))
		role := parseDeviceRole(firstNonEmpty(sample.Metric["device_role"], "desktop"))
		if sample.Metric["job"] == "vm-node-exporter" {
			kind = parseDeviceKind(firstNonEmpty(sample.Metric["device_kind"], "virtual_machine"))
			role = parseDeviceRole(firstNonEmpty(sample.Metric["device_role"], "vm"))
		}
		devices = append(devices, &mev1.DeviceState{
			DeviceUid:   deviceUID,
			DisplayName: deviceName,
			Model:       firstNonEmpty(sample.Metric["device_model"], models[instance]),
			Kind:        kind,
			Role:        role,
			State:       onlineState(sample.Value > 0),
			UpdateTime:  timestamppb.New(time.Now().UTC()),
			Metrics: systemMetrics(
				cpuCores[instance],
				cpuUsage[instance],
				memoryTotal[instance],
				memoryAvailable[instance],
				diskTotal[instance],
				diskAvailable[instance],
			),
			Media:       media[deviceUID],
			Accessories: accessories[deviceUID],
		})
	}
	sort.Slice(devices, func(left int, right int) bool {
		return devices[left].GetDeviceUid() < devices[right].GetDeviceUid()
	})
	return devices
}

func (client *PrometheusClient) DeviceMediaStatuses(ctx context.Context) map[string]*mev1.MediaStatus {
	samples := client.queryVector(ctx, `realtime_device_media_playing{job="device-exporter"} == 1`)
	statuses := make(map[string]*mev1.MediaStatus, len(samples))
	for _, sample := range samples {
		title := sample.Metric["title"]
		if title == "" {
			continue
		}
		deviceID := firstNonEmpty(sample.Metric["device_uid"], sample.Metric["device_id"], sample.Metric["instance"])
		if deviceID == "" {
			continue
		}
		statuses[deviceID] = &mev1.MediaStatus{
			Title:  title,
			Artist: sample.Metric["artist"],
		}
	}
	return statuses
}

func (client *PrometheusClient) DeviceAccessoryStatuses(ctx context.Context) map[string][]*mev1.Accessory {
	connected := client.queryVector(ctx, `realtime_device_accessory_connected{job="device-exporter"} == 1`)
	batteries := client.queryVector(ctx, `realtime_device_accessory_battery_level_ratio{job="device-exporter"}`)
	statuses := make(map[string][]*mev1.Accessory, len(connected))
	byKey := make(map[string]*mev1.Accessory, len(connected)+len(batteries))

	for _, sample := range connected {
		deviceID, accessory, ok := accessoryFromMetric(sample.Metric)
		if !ok {
			continue
		}
		byKey[deviceID+"\x00"+accessoryKey(accessory)] = accessory
	}
	for _, sample := range batteries {
		deviceID, accessory, ok := accessoryFromMetric(sample.Metric)
		if !ok {
			continue
		}
		percent := int32(math.Round(clampRatio(sample.Value) * 100))
		key := deviceID + "\x00" + accessoryKey(accessory)
		if existing, exists := byKey[key]; exists {
			existing.BatteryPercent = &percent
		} else {
			accessory.BatteryPercent = &percent
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

// AgentStatuses expands the exporter's per-model counts into one Agent per agent
// working right now. A host can drive several agents of one kind at once, and
// each is its own card: the exporter counts them, and the gateway names them.
// The budget and the sub-agents are reported for the kind rather than for one
// agent among several, so they are carried by the first agent of that kind.
func (client *PrometheusClient) AgentStatuses(ctx context.Context) []*mev1.Agent {
	var runningCounts, budgetSamples, subagentSamples []prometheusSample
	parallel(
		func() {
			runningCounts = client.queryVector(ctx, `realtime_agent_running_count{job="agent-exporter"} > 0`)
		},
		func() {
			budgetSamples = client.queryVector(ctx, `realtime_agent_budget_remaining_ratio{job="agent-exporter"}`)
		},
		func() {
			subagentSamples = client.queryVector(ctx, `realtime_agent_subagents_running{job="agent-exporter"}`)
		},
	)
	if len(runningCounts) == 0 {
		return nil
	}
	budgets := samplesByAgent(budgetSamples)
	subagents := agentSubagents(subagentSamples)
	now := timestamppb.New(time.Now().UTC())

	agents := make([]*mev1.Agent, 0, len(runningCounts))
	for _, sample := range runningCounts {
		kind := firstNonEmpty(sample.Metric["agent_kind"], sample.Metric["agent_id"])
		count := sampleCount(sample.Value)
		if kind == "" || count == 0 {
			continue
		}
		deviceUID := firstNonEmpty(sample.Metric["device_uid"], sample.Metric["device_id"])
		model := sample.Metric["model"]
		for ordinal := 0; ordinal < count; ordinal++ {
			agents = append(agents, &mev1.Agent{
				Uid:         agentUID(deviceUID, kind, model, ordinal),
				Kind:        kind,
				DeviceUid:   deviceUID,
				DisplayName: sample.Metric["device_name"],
				State:       mev1.AgentState_AGENT_STATE_RUNNING,
				UpdateTime:  now,
				Model:       model,
			})
		}
	}
	sort.Slice(agents, func(left int, right int) bool {
		if agents[left].GetDeviceUid() != agents[right].GetDeviceUid() {
			return agents[left].GetDeviceUid() < agents[right].GetDeviceUid()
		}
		if agents[left].GetKind() != agents[right].GetKind() {
			return agents[left].GetKind() < agents[right].GetKind()
		}
		if agents[left].GetModel() != agents[right].GetModel() {
			return agents[left].GetModel() < agents[right].GetModel()
		}
		return agents[left].GetUid() < agents[right].GetUid()
	})

	// Once sorted, the first agent of a kind is the one that carries what the
	// exporter could only report for the kind as a whole.
	seen := make(map[string]struct{}, len(agents))
	for _, agent := range agents {
		key := agent.GetDeviceUid() + "/" + agent.GetKind()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		agent.Subagents = subagents[key]
		if value, ok := budgets[key]; ok {
			percent := int32(math.Round(clampRatio(value) * 100))
			agent.BudgetRemainingPercent = &percent
		}
	}
	return agents
}

func accessoryFromMetric(labels map[string]string) (string, *mev1.Accessory, bool) {
	deviceID := firstNonEmpty(labels["device_uid"], labels["device_id"], labels["instance"])
	kind := labels["accessory_kind"]
	name := labels["accessory_name"]
	if deviceID == "" || kind == "" || name == "" {
		return "", nil, false
	}
	return deviceID, &mev1.Accessory{
		Kind:        kind,
		DisplayName: name,
		Model:       labels["accessory_model"],
	}, true
}

func accessoryKey(accessory *mev1.Accessory) string {
	return accessory.GetKind() + "\x00" + accessory.GetDisplayName() + "\x00" + accessory.GetModel()
}

func sortAccessories(accessories []*mev1.Accessory) {
	sort.Slice(accessories, func(left int, right int) bool {
		if accessories[left].GetKind() != accessories[right].GetKind() {
			return accessories[left].GetKind() < accessories[right].GetKind()
		}
		return accessories[left].GetDisplayName() < accessories[right].GetDisplayName()
	})
}

// QueryRange samples one expression over a time range. The expression is always
// built by the gateway from a MetricSeries; it never originates with a caller.
func (client *PrometheusClient) QueryRange(ctx context.Context, query string, start time.Time, end time.Time, step time.Duration) ([]*mev1.MetricPoint, error) {
	endpoint, err := url.Parse(client.baseURL + "/api/v1/query_range")
	if err != nil {
		return nil, err
	}
	values := endpoint.Query()
	values.Set("query", query)
	values.Set("start", strconv.FormatInt(start.Unix(), 10))
	values.Set("end", strconv.FormatInt(end.Unix(), 10))
	values.Set("step", strconv.FormatFloat(step.Seconds(), 'f', -1, 64))
	endpoint.RawQuery = values.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")

	response, err := client.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil, fmt.Errorf("prometheus returned %d", response.StatusCode)
	}

	var payload prometheusRangeResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, maxPrometheusResponseSize)).Decode(&payload); err != nil {
		return nil, err
	}
	if len(payload.Data.Result) == 0 {
		return nil, nil
	}
	return metricPoints(payload.Data.Result[0].Values), nil
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
	if err := json.NewDecoder(io.LimitReader(response.Body, maxPrometheusResponseSize)).Decode(&payload); err != nil {
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
		// Prometheus renders NaN and ±Inf as samples like any other. A missing
		// sample is the honest answer for both: NaN survives every arithmetic
		// guard downstream -- it clamps to NaN, rounds to NaN, and reaches the
		// page as a full gauge captioned "NaN%" -- so it is refused here, where
		// the range path already refuses it.
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
			continue
		}
		samples = append(samples, prometheusSample{Metric: result.Metric, Value: parsed})
	}
	return samples
}

type prometheusRangeResponse struct {
	Data struct {
		Result []struct {
			Values [][]any `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// metricPoints converts Prometheus's [unixSeconds, "value"] pairs, skipping any
// pair that is malformed or not a finite number.
func metricPoints(values [][]any) []*mev1.MetricPoint {
	points := make([]*mev1.MetricPoint, 0, len(values))
	for _, pair := range values {
		if len(pair) < 2 {
			continue
		}
		seconds, ok := pair[0].(float64)
		if !ok {
			continue
		}
		raw, ok := pair[1].(string)
		if !ok {
			continue
		}
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
			continue
		}
		points = append(points, &mev1.MetricPoint{
			Time:  timestamppb.New(time.Unix(int64(seconds), 0).UTC()),
			Value: value,
		})
	}
	return points
}

type prometheusQueryResponse struct {
	Data struct {
		Result []struct {
			Metric map[string]string `json:"metric"`
			Value  []any             `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func systemMetrics(cpuCores *float64, cpuUsage *float64, memoryTotal *float64, memoryAvailable *float64, diskTotal *float64, diskAvailable *float64) []*mev1.MetricSample {
	metrics := make([]*mev1.MetricSample, 0, 7)
	metrics = appendMetric(metrics, metricSystemCPULogicalCount, "{cpu}", cpuCores, nil)
	metrics = appendMetric(metrics, metricSystemCPUUtilization, "1", clampRatioPointer(cpuUsage), nil)
	metrics = appendMetric(metrics, metricSystemMemoryUsage, "By", subtract(memoryTotal, memoryAvailable), map[string]string{"system.memory.state": "used"})
	metrics = appendMetric(metrics, metricSystemMemoryLimit, "By", memoryTotal, nil)
	metrics = appendMetric(metrics, metricSystemFilesystemUsage, "By", subtract(diskTotal, diskAvailable), map[string]string{"mountpoint": "/"})
	metrics = appendMetric(metrics, metricSystemFilesystemLimit, "By", diskTotal, map[string]string{"mountpoint": "/"})
	metrics = appendMetric(metrics, metricSystemFilesystemUsagePct, "1", ratio(subtract(diskTotal, diskAvailable), diskTotal), map[string]string{"mountpoint": "/"})
	return metrics
}

func appendMetric(metrics []*mev1.MetricSample, name string, unit string, value *float64, attributes map[string]string) []*mev1.MetricSample {
	if value == nil {
		return metrics
	}
	return append(metrics, &mev1.MetricSample{
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

// agentSubagents expands the per-model counts into one Subagent per worker, so a
// caller can render a sub-agent without knowing how the count was labelled. The
// exporter emits one series per model, and a zero keeps the series alive rather
// than describing a worker.
func agentSubagents(samples []prometheusSample) map[string][]*mev1.Subagent {
	subagents := make(map[string][]*mev1.Subagent, len(samples))
	for _, sample := range samples {
		kind := firstNonEmpty(sample.Metric["agent_kind"], sample.Metric["agent_id"])
		count := sampleCount(sample.Value)
		if kind == "" || count == 0 {
			continue
		}
		key := firstNonEmpty(sample.Metric["device_uid"], sample.Metric["device_id"]) + "/" + kind
		for index := 0; index < count; index++ {
			subagents[key] = append(subagents[key], &mev1.Subagent{Model: sample.Metric["model"]})
		}
	}
	for key := range subagents {
		sort.Slice(subagents[key], func(left int, right int) bool {
			return subagents[key][left].GetModel() < subagents[key][right].GetModel()
		})
	}
	return subagents
}

func samplesByAgent(samples []prometheusSample) map[string]float64 {
	values := make(map[string]float64, len(samples))
	for _, sample := range samples {
		kind := firstNonEmpty(sample.Metric["agent_kind"], sample.Metric["agent_id"])
		if kind == "" {
			continue
		}
		deviceUID := firstNonEmpty(sample.Metric["device_uid"], sample.Metric["device_id"])
		values[deviceUID+"/"+kind] = sample.Value
	}
	return values
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
	deviceID := firstPublicNetworkLabel(labels["device_uid"], labels["device_id"], labels["device_name"], labels["instance"])
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

// sampleCount reads a counting gauge as the number of things it counts. The
// count sizes an allocation on an unauthenticated path, so a sample claiming
// more than any host could be running is taken for as much of it as the gateway
// is willing to draw, and one that counts nothing at all counts nothing.
func sampleCount(value float64) int {
	count := math.Round(value)
	if math.IsNaN(count) || count < 1 {
		return 0
	}
	return int(math.Min(maxAgentsPerSample, count))
}

func roundMetric(value float64) float64 {
	return math.Round(value*1000) / 1000
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
