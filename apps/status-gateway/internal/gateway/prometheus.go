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

func (client *PrometheusClient) ServerSummary(ctx context.Context) ServerSummary {
	up := client.queryScalar(ctx, `max(up{job="node-exporter"})`)
	cpu := client.queryScalar(ctx, `100 * (1 - avg(rate(node_cpu_seconds_total{job="node-exporter",mode="idle"}[2m])))`)
	memory := client.queryScalar(ctx, `100 * (1 - (node_memory_MemAvailable_bytes{job="node-exporter"} / node_memory_MemTotal_bytes{job="node-exporter"}))`)
	disk := client.queryScalar(ctx, `100 * (1 - (node_filesystem_avail_bytes{job="node-exporter",mountpoint="/",fstype!~"tmpfs|overlay|squashfs"} / node_filesystem_size_bytes{job="node-exporter",mountpoint="/",fstype!~"tmpfs|overlay|squashfs"}))`)

	return ServerSummary{
		Online:        up != nil && *up > 0,
		CPUPercent:    clampPercent(cpu),
		MemoryPercent: clampPercent(memory),
		DiskPercent:   clampPercent(disk),
		UpdatedAt:     time.Now().UTC().Format(time.RFC3339),
	}
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

func clampPercent(value *float64) *float64 {
	if value == nil {
		return nil
	}
	clamped := math.Round(math.Max(0, math.Min(100, *value))*10) / 10
	return &clamped
}
