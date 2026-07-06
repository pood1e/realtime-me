package gateway

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type metricDefinition struct {
	Name        string
	Type        string
	Unit        string
	Description string
	OTelName    string
}

var metricDefinitions = []metricDefinition{
	{Name: "realtime_device_last_update_time_seconds", Type: "gauge", Unit: "s", OTelName: "realtime.device.last_update", Description: "Unix timestamp of the latest accepted device update."},
	{Name: "realtime_device_battery_level_ratio", Type: "gauge", Unit: "1", OTelName: "realtime.device.battery.level", Description: "Device battery level as a fraction of total capacity."},
	{Name: "realtime_device_charging", Type: "gauge", Unit: "1", OTelName: "realtime.device.charging", Description: "Device charging state: 1 for charging, 0 otherwise."},
	{Name: "realtime_device_network_state", Type: "gauge", Unit: "1", OTelName: "realtime.device.network.state", Description: "Phone network state labelled by network_type; current state is 1."},
	{Name: "realtime_host_cpu_cores", Type: "gauge", Unit: "{cpu}", OTelName: "system.cpu.logical.count", Description: "Host CPU logical core count."},
	{Name: "realtime_host_cpu_usage_ratio", Type: "gauge", Unit: "1", OTelName: "system.cpu.utilization", Description: "Host CPU utilization as a fraction."},
	{Name: "realtime_host_memory_usage_bytes", Type: "gauge", Unit: "By", OTelName: "system.memory.usage", Description: "Host memory usage in bytes."},
	{Name: "realtime_host_memory_limit_bytes", Type: "gauge", Unit: "By", OTelName: "system.memory.limit", Description: "Host memory capacity in bytes."},
	{Name: "realtime_host_filesystem_usage_bytes", Type: "gauge", Unit: "By", OTelName: "system.filesystem.usage", Description: "Host filesystem usage in bytes."},
	{Name: "realtime_host_filesystem_limit_bytes", Type: "gauge", Unit: "By", OTelName: "system.filesystem.limit", Description: "Host filesystem capacity in bytes."},
	{Name: "realtime_host_filesystem_usage_ratio", Type: "gauge", Unit: "1", OTelName: "system.filesystem.utilization", Description: "Host filesystem utilization as a fraction."},
	{Name: "realtime_host_vm_state", Type: "gauge", Unit: "1", OTelName: "realtime.host.vm.state", Description: "Virtual machine state labelled by vm_name and state; current state is 1."},
	{Name: "realtime_watch_heart_rate_beats_per_minute", Type: "gauge", Unit: "{beat}/min", OTelName: "realtime.watch.heart_rate", Description: "Latest watch heart rate."},
	{Name: "realtime_watch_steps", Type: "gauge", Unit: "{step}", OTelName: "realtime.watch.steps", Description: "Latest watch local-day step count."},
	{Name: "realtime_watch_wrist_state", Type: "gauge", Unit: "1", OTelName: "realtime.watch.wrist.state", Description: "Watch wrist state labelled by wrist_state; current state is 1."},
	{Name: "realtime_agent_state", Type: "gauge", Unit: "1", OTelName: "realtime.agent.state", Description: "Agent state labelled by state; current state is 1."},
	{Name: "realtime_agent_budget_remaining_ratio", Type: "gauge", Unit: "1", OTelName: "realtime.agent.budget.remaining", Description: "Agent budget remaining as a fraction."},
	{Name: "realtime_github_status_sync_state", Type: "gauge", Unit: "1", OTelName: "realtime.github.status.sync.state", Description: "GitHub status sync state labelled by state; current state is 1."},
}

func RenderMetrics(snapshot GatewayStateSnapshot) string {
	lines := make([]string, 0, 64)
	appendMetricMetadata(&lines)
	if snapshot.Mobile != nil {
		appendMobileMetrics(&lines, *snapshot.Mobile)
	}
	for _, agent := range snapshot.Agents {
		appendAgentMetrics(&lines, agent)
	}
	for _, device := range snapshot.Devices {
		appendReportedDeviceMetrics(&lines, device)
	}
	appendGitHubMetrics(&lines, snapshot.GitHub)
	return strings.Join(append(lines, ""), "\n")
}

func appendMetricMetadata(lines *[]string) {
	for _, metric := range metricDefinitions {
		*lines = append(*lines,
			fmt.Sprintf("# HELP %s %s OpenTelemetry name: %s.", metric.Name, metric.Description, metric.OTelName),
			fmt.Sprintf("# TYPE %s %s", metric.Name, metric.Type),
			fmt.Sprintf("# UNIT %s %s", metric.Name, metric.Unit),
		)
	}
}

func appendMobileMetrics(lines *[]string, status StoredMobileStatus) {
	appendDeviceTimestamp(lines, status.DeviceID, "phone", status.ReceivedAt)
	if status.Phone != nil {
		appendBattery(lines, status.DeviceID, "phone", status.Phone.BatteryPercent)
		if status.Phone.ChargeState != "" {
			appendSample(lines, "realtime_device_charging", deviceLabels(status.DeviceID, "phone"), boolFloat(status.Phone.ChargeState == ChargeCharging))
		}
		if status.Phone.Network != "" {
			appendState(lines, "realtime_device_network_state", map[string]string{"device_id": status.DeviceID}, "network_type", status.Phone.Network, []string{"offline", "wifi", "cellular", "vpn", "online", "unknown"})
		}
	}
	if status.Watch == nil {
		return
	}
	appendDeviceTimestamp(lines, status.DeviceID, "watch", status.ReceivedAt)
	appendBattery(lines, status.DeviceID, "watch", status.Watch.BatteryPercent)
	if status.Watch.ChargeState != "" {
		appendSample(lines, "realtime_device_charging", deviceLabels(status.DeviceID, "watch"), boolFloat(status.Watch.ChargeState == ChargeCharging))
	}
	if status.Watch.HeartRate != nil && status.Watch.WristState != WristOffWrist {
		appendSample(lines, "realtime_watch_heart_rate_beats_per_minute", map[string]string{"device_id": status.DeviceID}, float64(*status.Watch.HeartRate))
	}
	if status.Watch.Steps != nil {
		appendSample(lines, "realtime_watch_steps", map[string]string{"device_id": status.DeviceID}, float64(*status.Watch.Steps))
	}
	if status.Watch.WristState != "" {
		appendState(lines, "realtime_watch_wrist_state", map[string]string{"device_id": status.DeviceID}, "wrist_state", string(status.Watch.WristState), []string{string(WristUnknown), string(WristOnWrist), string(WristOffWrist)})
	}
}

func appendDeviceTimestamp(lines *[]string, deviceID string, deviceType string, receivedAt string) {
	appendSample(lines, "realtime_device_last_update_time_seconds", deviceLabels(deviceID, deviceType), float64(unixSeconds(receivedAt)))
}

func appendBattery(lines *[]string, deviceID string, deviceType string, value *int) {
	if value == nil {
		return
	}
	appendSample(lines, "realtime_device_battery_level_ratio", deviceLabels(deviceID, deviceType), float64(*value)/100)
}

func appendAgentMetrics(lines *[]string, status StoredAgentStatus) {
	labels := map[string]string{"agent_id": status.AgentID}
	if status.DeviceID != "" {
		labels["device_id"] = status.DeviceID
	}
	if status.DeviceName != "" {
		labels["device_name"] = status.DeviceName
	}
	appendState(lines, "realtime_agent_state", labels, "state", status.State, []string{"idle", "running", "failed"})
	if status.BudgetRemainingPercent != nil {
		appendSample(lines, "realtime_agent_budget_remaining_ratio", labels, float64(*status.BudgetRemainingPercent)/100)
	}
}

func appendReportedDeviceMetrics(lines *[]string, status StoredDeviceStatus) {
	labels := map[string]string{"device_id": status.DeviceID}
	if status.Role != "" {
		labels["role"] = status.Role
	}
	appendDeviceTimestamp(lines, status.DeviceID, status.Kind, status.ReceivedAt)
	for _, metric := range status.Metrics {
		appendReportedMetric(lines, labels, metric)
	}
	for _, child := range status.Children {
		childLabels := map[string]string{
			"device_id":        child.DeviceID,
			"parent_device_id": status.DeviceID,
		}
		if child.Kind != "" {
			childLabels["child_kind"] = child.Kind
		}
		appendDeviceTimestamp(lines, child.DeviceID, child.Kind, child.UpdatedAt)
		appendState(lines, "realtime_host_vm_state", childLabels, "state", child.State, []string{"running", "shut off", "paused", "unknown"})
		for _, metric := range child.Metrics {
			appendReportedMetric(lines, childLabels, metric)
		}
	}
}

func appendReportedMetric(lines *[]string, baseLabels map[string]string, metric MetricSample) {
	labels := copyLabels(baseLabels)
	for key, value := range metric.Attributes {
		labels[key] = value
	}
	switch metric.Name {
	case metricSystemCPULogicalCount:
		appendSample(lines, "realtime_host_cpu_cores", labels, metric.Value)
	case metricSystemCPUUtilization:
		appendSample(lines, "realtime_host_cpu_usage_ratio", labels, metric.Value)
	case metricSystemMemoryUsage:
		appendSample(lines, "realtime_host_memory_usage_bytes", labels, metric.Value)
	case metricSystemMemoryLimit:
		appendSample(lines, "realtime_host_memory_limit_bytes", labels, metric.Value)
	case metricSystemFilesystemUsage:
		appendSample(lines, "realtime_host_filesystem_usage_bytes", labels, metric.Value)
	case metricSystemFilesystemLimit:
		appendSample(lines, "realtime_host_filesystem_limit_bytes", labels, metric.Value)
	case metricSystemFilesystemUsagePct:
		appendSample(lines, "realtime_host_filesystem_usage_ratio", labels, metric.Value)
	}
}

func appendGitHubMetrics(lines *[]string, status GitHubSyncStatus) {
	appendState(lines, "realtime_github_status_sync_state", nil, "state", string(status.State), []string{string(GitHubSyncDisabled), string(GitHubSyncPending), string(GitHubSyncOK), string(GitHubSyncError)})
}

func appendState(lines *[]string, metric string, baseLabels map[string]string, stateLabel string, current string, states []string) {
	seen := map[string]struct{}{}
	for _, state := range states {
		seen[state] = struct{}{}
		labels := copyLabels(baseLabels)
		labels[stateLabel] = state
		appendSample(lines, metric, labels, boolFloat(current == state))
	}
	if _, ok := seen[current]; current != "" && !ok {
		labels := copyLabels(baseLabels)
		labels[stateLabel] = current
		appendSample(lines, metric, labels, 1)
	}
}

func appendSample(lines *[]string, name string, labels map[string]string, value float64) {
	*lines = append(*lines, fmt.Sprintf("%s%s %s", name, labelSet(labels), strconv.FormatFloat(value, 'f', -1, 64)))
}

func deviceLabels(deviceID string, deviceType string) map[string]string {
	return map[string]string{"device_id": deviceID, "device_type": deviceType}
}

func copyLabels(labels map[string]string) map[string]string {
	copy := map[string]string{}
	for key, value := range labels {
		copy[key] = value
	}
	return copy
}

func unixSeconds(value string) int64 {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return 0
	}
	return parsed.Unix()
}

func labelSet(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		value := labels[key]
		pairs = append(pairs, fmt.Sprintf("%s=%s", prometheusLabelName(key), strconv.Quote(value)))
	}
	return "{" + strings.Join(pairs, ",") + "}"
}

func prometheusLabelName(value string) string {
	var builder strings.Builder
	for index, character := range value {
		valid := character == '_' || character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || index > 0 && character >= '0' && character <= '9'
		if valid {
			builder.WriteRune(character)
		} else {
			builder.WriteByte('_')
		}
	}
	if builder.Len() == 0 {
		return "label"
	}
	name := builder.String()
	if name[0] >= '0' && name[0] <= '9' {
		return "label_" + name
	}
	return name
}

func boolFloat(value bool) float64 {
	if value {
		return 1
	}
	return 0
}
