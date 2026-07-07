package gateway

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"google.golang.org/protobuf/types/known/timestamppb"

	mev1 "realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1"
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
	{Name: "realtime_device_media_playing", Type: "gauge", Unit: "1", OTelName: "realtime.device.media.playing", Description: "Device media playback state with current title and artist labels when available."},
	{Name: "realtime_device_accessory_connected", Type: "gauge", Unit: "1", OTelName: "realtime.device.accessory.connected", Description: "Connected accessory state labelled by accessory name and kind."},
	{Name: "realtime_device_accessory_battery_level_ratio", Type: "gauge", Unit: "1", OTelName: "realtime.device.accessory.battery.level", Description: "Accessory battery level as a fraction of total capacity."},
	{Name: "realtime_watch_heart_rate_beats_per_minute", Type: "gauge", Unit: "{beat}/min", OTelName: "realtime.watch.heart_rate", Description: "Latest watch heart rate."},
	{Name: "realtime_watch_steps", Type: "gauge", Unit: "{step}", OTelName: "realtime.watch.steps", Description: "Latest watch local-day step count."},
	{Name: "realtime_watch_wrist_state", Type: "gauge", Unit: "1", OTelName: "realtime.watch.wrist.state", Description: "Watch wrist state labelled by wrist_state; current state is 1."},
	{Name: "realtime_agent_state", Type: "gauge", Unit: "1", OTelName: "realtime.agent.state", Description: "Agent state labelled by state; current state is 1."},
	{Name: "realtime_agent_budget_remaining_ratio", Type: "gauge", Unit: "1", OTelName: "realtime.agent.budget.remaining", Description: "Agent budget remaining as a fraction."},
	{Name: "realtime_github_status_sync_state", Type: "gauge", Unit: "1", OTelName: "realtime.github.status.sync.state", Description: "GitHub status sync state labelled by state; current state is 1."},
}

func RenderMetrics(snapshot StatusSnapshot) string {
	lines := make([]string, 0, 64)
	appendMetricMetadata(&lines)
	if snapshot.Mobile != nil {
		appendMobileMetrics(&lines, snapshot.Mobile)
	}
	for _, agent := range snapshot.Agents {
		appendAgentMetrics(&lines, agent)
	}
	for _, device := range snapshot.Hosts {
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

func appendMobileMetrics(lines *[]string, mobile *mev1.MobileState) {
	deviceID := mobile.GetDeviceUid()
	appendDeviceTimestamp(lines, deviceID, "phone", mobile.GetUpdateTime())
	if phone := mobile.GetPhone(); phone != nil {
		if phone.BatteryPercent != nil {
			appendBattery(lines, deviceID, "phone", phone.GetBatteryPercent())
		}
		if phone.GetChargeState() != mev1.ChargeState_CHARGE_STATE_UNSPECIFIED {
			appendSample(lines, "realtime_device_charging", deviceLabels(deviceID, "phone"), boolFloat(phone.GetChargeState() == mev1.ChargeState_CHARGE_STATE_CHARGING))
		}
		if phone.GetNetwork() != mev1.NetworkState_NETWORK_STATE_UNSPECIFIED {
			appendState(lines, "realtime_device_network_state", map[string]string{"device_id": deviceID}, "network_type", networkStateString(phone.GetNetwork()), []string{"offline", "wifi", "cellular", "vpn", "online", "unknown"})
		}
		appendAccessories(lines, deviceLabels(deviceID, "phone"), phone.GetAccessories())
	}

	watch := mobile.GetWatch()
	if watch == nil {
		return
	}
	state := watch.GetWatchState()
	appendDeviceTimestamp(lines, deviceID, "watch", mobile.GetUpdateTime())
	if state != nil {
		appendBattery(lines, deviceID, "watch", state.GetBatteryPercent())
		if state.GetChargeState() != mev1.ChargeState_CHARGE_STATE_UNSPECIFIED {
			appendSample(lines, "realtime_device_charging", deviceLabels(deviceID, "watch"), boolFloat(state.GetChargeState() == mev1.ChargeState_CHARGE_STATE_CHARGING))
		}
	}
	if watch.GetHeartRate() != nil && state.GetWristState() != mev1.WristState_WRIST_STATE_OFF_WRIST {
		appendSample(lines, "realtime_watch_heart_rate_beats_per_minute", map[string]string{"device_id": deviceID}, float64(watch.GetHeartRate().GetBeatsPerMinute()))
	}
	if watch.GetActivityTotals() != nil {
		appendSample(lines, "realtime_watch_steps", map[string]string{"device_id": deviceID}, float64(watch.GetActivityTotals().GetSteps()))
	}
	if state.GetWristState() != mev1.WristState_WRIST_STATE_UNSPECIFIED {
		appendState(lines, "realtime_watch_wrist_state", map[string]string{"device_id": deviceID}, "wrist_state", wristStateString(state.GetWristState()), []string{"unknown", "on_wrist", "off_wrist"})
	}
}

func appendDeviceTimestamp(lines *[]string, deviceID string, deviceType string, updateTime *timestamppb.Timestamp) {
	appendSample(lines, "realtime_device_last_update_time_seconds", deviceLabels(deviceID, deviceType), float64(unixSeconds(updateTime)))
}

func appendBattery(lines *[]string, deviceID string, deviceType string, value int32) {
	appendSample(lines, "realtime_device_battery_level_ratio", deviceLabels(deviceID, deviceType), float64(value)/100)
}

func appendAgentMetrics(lines *[]string, agent *mev1.Agent) {
	labels := map[string]string{"agent_id": agent.GetKind()}
	if agent.GetDeviceUid() != "" {
		labels["device_id"] = agent.GetDeviceUid()
	}
	if agent.GetDisplayName() != "" {
		labels["device_name"] = agent.GetDisplayName()
	}
	appendState(lines, "realtime_agent_state", labels, "state", agentStateString(agent.GetState()), []string{"idle", "running", "failed"})
	if agent.BudgetRemainingPercent != nil {
		appendSample(lines, "realtime_agent_budget_remaining_ratio", labels, float64(agent.GetBudgetRemainingPercent())/100)
	}
}

func appendReportedDeviceMetrics(lines *[]string, device *mev1.DeviceState) {
	labels := map[string]string{"device_id": device.GetDeviceUid()}
	if role := deviceRoleString(device.GetRole()); role != "" {
		labels["role"] = role
	}
	appendDeviceTimestamp(lines, device.GetDeviceUid(), deviceKindString(device.GetKind()), device.GetUpdateTime())
	for _, metric := range device.GetMetrics() {
		appendReportedMetric(lines, labels, metric)
	}
	if device.GetMedia() != nil {
		appendMedia(lines, labels, device.GetMedia())
	}
	appendAccessories(lines, labels, device.GetAccessories())
	for _, child := range device.GetChildren() {
		childLabels := map[string]string{
			"device_id":        child.GetDeviceUid(),
			"parent_device_id": device.GetDeviceUid(),
		}
		if kind := deviceKindString(child.GetKind()); kind != "" {
			childLabels["child_kind"] = kind
		}
		appendDeviceTimestamp(lines, child.GetDeviceUid(), deviceKindString(child.GetKind()), child.GetUpdateTime())
		appendState(lines, "realtime_host_vm_state", childLabels, "state", onlineStateString(child.GetState()), []string{"online", "offline"})
		for _, metric := range child.GetMetrics() {
			appendReportedMetric(lines, childLabels, metric)
		}
		if child.GetMedia() != nil {
			appendMedia(lines, childLabels, child.GetMedia())
		}
		appendAccessories(lines, childLabels, child.GetAccessories())
	}
}

func appendMedia(lines *[]string, baseLabels map[string]string, media *mev1.MediaStatus) {
	if media == nil || media.GetTitle() == "" {
		return
	}
	labels := copyLabels(baseLabels)
	labels["title"] = media.GetTitle()
	if media.GetArtist() != "" {
		labels["artist"] = media.GetArtist()
	}
	appendSample(lines, "realtime_device_media_playing", labels, 1)
}

func appendAccessories(lines *[]string, baseLabels map[string]string, accessories []*mev1.Accessory) {
	for _, accessory := range accessories {
		if accessory.GetKind() == "" || accessory.GetDisplayName() == "" {
			continue
		}
		labels := copyLabels(baseLabels)
		labels["accessory_kind"] = accessory.GetKind()
		labels["accessory_name"] = accessory.GetDisplayName()
		if accessory.GetModel() != "" {
			labels["accessory_model"] = accessory.GetModel()
		}
		appendSample(lines, "realtime_device_accessory_connected", labels, 1)
		if accessory.BatteryPercent != nil {
			appendSample(lines, "realtime_device_accessory_battery_level_ratio", labels, float64(accessory.GetBatteryPercent())/100)
		}
	}
}

func appendReportedMetric(lines *[]string, baseLabels map[string]string, metric *mev1.MetricSample) {
	labels := copyLabels(baseLabels)
	for key, value := range metric.GetAttributes() {
		labels[key] = value
	}
	switch metric.GetName() {
	case metricSystemCPULogicalCount:
		appendSample(lines, "realtime_host_cpu_cores", labels, metric.GetValue())
	case metricSystemCPUUtilization:
		appendSample(lines, "realtime_host_cpu_usage_ratio", labels, metric.GetValue())
	case metricSystemMemoryUsage:
		appendSample(lines, "realtime_host_memory_usage_bytes", labels, metric.GetValue())
	case metricSystemMemoryLimit:
		appendSample(lines, "realtime_host_memory_limit_bytes", labels, metric.GetValue())
	case metricSystemFilesystemUsage:
		appendSample(lines, "realtime_host_filesystem_usage_bytes", labels, metric.GetValue())
	case metricSystemFilesystemLimit:
		appendSample(lines, "realtime_host_filesystem_limit_bytes", labels, metric.GetValue())
	case metricSystemFilesystemUsagePct:
		appendSample(lines, "realtime_host_filesystem_usage_ratio", labels, metric.GetValue())
	}
}

func appendGitHubMetrics(lines *[]string, status *mev1.GithubSyncDetail) {
	appendState(lines, "realtime_github_status_sync_state", nil, "state", githubStateString(status.GetState()), []string{"disabled", "pending", "ok", "error"})
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
	clone := map[string]string{}
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

func wristStateString(state mev1.WristState) string {
	switch state {
	case mev1.WristState_WRIST_STATE_ON_WRIST:
		return "on_wrist"
	case mev1.WristState_WRIST_STATE_OFF_WRIST:
		return "off_wrist"
	default:
		return "unknown"
	}
}
