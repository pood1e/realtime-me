package gateway

import (
	"net"
	"strings"
	"time"
	"unicode/utf8"
)

func validateMobile(input *MobileIngest) bool {
	if input == nil || len(input.DeviceID) < 1 || len(input.DeviceID) > 80 {
		return false
	}
	if len(input.DeviceName) > 80 || len(input.DeviceModel) > 120 {
		return false
	}
	if input.UpdatedAt != "" && !isDateTime(input.UpdatedAt) {
		return false
	}
	if input.Phone != nil && !validatePhone(input.Phone) {
		return false
	}
	if input.Watch != nil && !validateWatch(input.Watch) {
		return false
	}
	return true
}

func normalizeMobile(input *MobileIngest) {
	if input.Phone != nil && input.Phone.ChargeState == "" {
		input.Phone.ChargeState = ChargeUnknown
	}
	if input.Watch != nil {
		if input.Watch.ChargeState == "" {
			input.Watch.ChargeState = ChargeUnknown
		}
		if input.Watch.WristState == "" {
			input.Watch.WristState = WristUnknown
		}
		if input.Watch.WristState == WristOffWrist {
			input.Watch.HeartRate = nil
		}
	}
}

func validatePhone(phone *PhoneStatus) bool {
	if phone.BatteryPercent != nil && !inRange(*phone.BatteryPercent, 0, 100) {
		return false
	}
	if phone.ChargeState != "" && !validChargeState(phone.ChargeState) {
		return false
	}
	return len(phone.Network) <= 40
}

func validateWatch(watch *WatchStatus) bool {
	if len(watch.DeviceName) > 80 || len(watch.DeviceModel) > 120 {
		return false
	}
	if watch.HeartRate != nil && !inRange(*watch.HeartRate, 1, 240) {
		return false
	}
	if watch.Steps != nil && !inRange(*watch.Steps, 0, 500000) {
		return false
	}
	if watch.BatteryPercent != nil && !inRange(*watch.BatteryPercent, 0, 100) {
		return false
	}
	if watch.ChargeState != "" && !validChargeState(watch.ChargeState) {
		return false
	}
	return watch.WristState == "" || watch.WristState == WristUnknown || watch.WristState == WristOnWrist || watch.WristState == WristOffWrist
}

func validateDevice(input *DeviceStatus) bool {
	if input == nil || len(input.DeviceID) < 1 || len(input.DeviceID) > 80 {
		return false
	}
	if len(input.DeviceName) > 80 || len(input.DeviceModel) > 120 || len(input.Kind) > 40 || len(input.Role) > 40 || len(input.State) > 40 {
		return false
	}
	if input.UpdatedAt != "" && !isDateTime(input.UpdatedAt) {
		return false
	}
	if input.Media != nil && !validateMedia(input.Media) {
		return false
	}
	for _, metric := range input.Metrics {
		if !validateMetric(metric) {
			return false
		}
	}
	for _, child := range input.Children {
		if !validateDevice(&child) {
			return false
		}
	}
	return true
}

func validateMedia(media *MediaStatus) bool {
	return media != nil &&
		utf8.RuneCountInString(media.Title) > 0 &&
		utf8.RuneCountInString(media.Title) <= 120 &&
		utf8.RuneCountInString(media.Artist) <= 120
}

func validateMetric(metric MetricSample) bool {
	if len(metric.Name) < 1 || len(metric.Name) > 120 || len(metric.Unit) > 24 {
		return false
	}
	for key, value := range metric.Attributes {
		if len(key) < 1 || len(key) > 80 || len(value) > 120 {
			return false
		}
	}
	return true
}

func validateAgent(input *AgentIngest) bool {
	if input == nil || len(input.AgentID) < 1 || len(input.AgentID) > 80 {
		return false
	}
	if len(input.DeviceID) > 80 || len(input.DeviceName) > 80 {
		return false
	}
	if input.UpdatedAt != "" && !isDateTime(input.UpdatedAt) {
		return false
	}
	if input.State == "" {
		input.State = "idle"
	}
	if input.State != "idle" && input.State != "running" && input.State != "failed" {
		return false
	}
	if utf8.RuneCountInString(input.Task) > 120 {
		return false
	}
	return input.BudgetRemainingPercent == nil || inRange(*input.BudgetRemainingPercent, 0, 100)
}

func validatePrometheusRegistration(input *PrometheusTargetRegistration) bool {
	if input == nil || len(input.Targets) < 1 || len(input.Targets) > 20 {
		return false
	}
	for _, target := range input.Targets {
		if !validatePrometheusTarget(target) {
			return false
		}
	}
	return true
}

func validatePrometheusTarget(target PrometheusScrapeTarget) bool {
	if !validPrometheusJob(target.Job) || len(target.Target) < 3 || len(target.Target) > 120 {
		return false
	}
	host, port, err := net.SplitHostPort(target.Target)
	if err != nil || strings.TrimSpace(host) == "" || strings.TrimSpace(port) == "" {
		return false
	}
	if strings.ContainsAny(target.Target, "/?#@ \t\r\n") {
		return false
	}
	if len(target.Instance) > 80 || len(target.DeviceID) > 80 || len(target.DeviceName) > 80 || len(target.DeviceModel) > 120 {
		return false
	}
	return len(target.DeviceKind) <= 40 && len(target.DeviceRole) <= 40
}

func validPrometheusJob(value string) bool {
	return value == "node-exporter" || value == "vm-node-exporter" || value == "device-exporter" || value == "agent-exporter"
}

func validChargeState(value ChargeState) bool {
	return value == ChargeUnknown || value == ChargeCharging || value == ChargeNotCharging
}

func inRange(value, minimum, maximum int) bool {
	return value >= minimum && value <= maximum
}

func isDateTime(value string) bool {
	_, err := time.Parse(time.RFC3339, value)
	return err == nil
}
