package qqmusic

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout       = 10 * time.Second
	maxJSONResponseBytes = 8 << 20
	musicUEndpoint       = "https://u.y.qq.com/cgi-bin/musicu.fcg"
	webUserAgent         = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

type Client struct {
	httpClient *http.Client
	userAgent  string
}

type clientConfig struct {
	timeout   time.Duration
	userAgent string
}

type Option func(*clientConfig) error

func WithTimeout(timeout time.Duration) Option {
	return func(config *clientConfig) error {
		if timeout <= 0 {
			return providerError("configure client", ErrorKindInvalidInput, "timeout must be positive", 0)
		}
		config.timeout = timeout
		return nil
	}
}

func NewClient(options ...Option) (*Client, error) {
	config := clientConfig{timeout: defaultTimeout, userAgent: webUserAgent}
	for _, option := range options {
		if option == nil {
			return nil, providerError("configure client", ErrorKindInvalidInput, "option is nil", 0)
		}
		if err := option(&config); err != nil {
			return nil, err
		}
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, providerError("create client", ErrorKindUnavailable, "cookie jar initialization failed", 0)
	}

	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, providerError("create client", ErrorKindUnavailable, "standard HTTP transport is unavailable", 0)
	}
	transport := defaultTransport.Clone()
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if transport.TLSClientConfig != nil {
		tlsConfig = transport.TLSClientConfig.Clone()
		tlsConfig.MinVersion = tls.VersionTLS12
	}
	transport.TLSClientConfig = tlsConfig
	transport.ForceAttemptHTTP2 = true

	return &Client{
		httpClient: &http.Client{
			Transport: transport,
			Jar:       jar,
			Timeout:   config.timeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		userAgent: config.userAgent,
	}, nil
}

type musicUCall struct {
	Module string `json:"module"`
	Method string `json:"method"`
	Param  any    `json:"param"`
}

type musicURequest struct {
	Common map[string]any `json:"comm"`
	Call   musicUCall     `json:"req_0"`
}

type musicUResponse struct {
	Code int             `json:"code"`
	Call json.RawMessage `json:"req_0"`
}

type musicUResult struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

func (client *Client) callMusicU(
	ctx context.Context,
	operation string,
	credentials Credentials,
	commonOverrides map[string]any,
	module string,
	method string,
	params any,
	allowBusinessError bool,
) (musicUResult, error) {
	payload := musicURequest{
		Common: buildWebCommon(credentials, commonOverrides),
		Call:   musicUCall{Module: module, Method: method, Param: params},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return musicUResult{}, providerError(operation, ErrorKindInvalidInput, "request encoding failed", 0)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, musicUEndpoint, bytes.NewReader(body))
	if err != nil {
		return musicUResult{}, providerError(operation, ErrorKindInvalidInput, "request creation failed", 0)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://y.qq.com")
	request.Header.Set("Referer", "https://y.qq.com/")
	client.applyCommonHeaders(request)
	applyCredentialCookies(request, credentials)

	response, err := client.execute(operation, request, http.StatusOK)
	if err != nil {
		return musicUResult{}, err
	}
	defer response.Body.Close()

	responseBody, err := readResponseBody(response.Body, maxJSONResponseBytes)
	if err != nil {
		return musicUResult{}, providerError(operation, ErrorKindUpstream, "response body is invalid", 0)
	}

	var envelope musicUResponse
	if err := decodeJSON(responseBody, &envelope); err != nil {
		return musicUResult{}, providerError(operation, ErrorKindUpstream, "response is not valid JSON", 0)
	}
	if envelope.Code != 0 {
		return musicUResult{}, providerError(operation, ErrorKindUpstream, "request was rejected", envelope.Code)
	}
	if len(envelope.Call) == 0 || bytes.Equal(envelope.Call, []byte("null")) {
		return musicUResult{}, providerError(operation, ErrorKindUpstream, "response is missing the requested operation", 0)
	}

	var result musicUResult
	if err := decodeJSON(envelope.Call, &result); err != nil {
		return musicUResult{}, providerError(operation, ErrorKindUpstream, "operation response is invalid", 0)
	}
	if result.Code != 0 && !allowBusinessError {
		return musicUResult{}, mapMusicUError(operation, result.Code)
	}
	return result, nil
}

func (client *Client) execute(operation string, request *http.Request, allowedStatuses ...int) (*http.Response, error) {
	response, err := client.httpClient.Do(request)
	if err != nil {
		if request.Context().Err() != nil {
			return nil, request.Context().Err()
		}
		var netError net.Error
		if errors.As(err, &netError) && netError.Timeout() {
			return nil, providerError(operation, ErrorKindNetwork, "request timed out", 0)
		}
		return nil, providerError(operation, ErrorKindNetwork, "request failed", 0)
	}

	for _, status := range allowedStatuses {
		if response.StatusCode == status {
			return response, nil
		}
	}
	response.Body.Close()
	return nil, providerError(operation, ErrorKindUpstream, "unexpected HTTP status", response.StatusCode)
}

func (client *Client) applyCommonHeaders(request *http.Request) {
	request.Header.Set("Accept", "application/json, text/plain, */*")
	request.Header.Set("User-Agent", client.userAgent)
}

func buildWebCommon(credentials Credentials, overrides map[string]any) map[string]any {
	token := 5381
	if credentials.MusicKey != "" {
		token = hash33(credentials.MusicKey, 5381)
	}
	common := map[string]any{
		"ct":                24,
		"cv":                4747474,
		"platform":          "yqq.json",
		"chid":              "0",
		"uin":               credentials.MusicID,
		"g_tk":              token,
		"g_tk_new_20200303": token,
		"format":            "json",
		"inCharset":         "utf-8",
		"outCharset":        "utf-8",
		"notice":            0,
		"needNewCode":       1,
	}
	for key, value := range overrides {
		common[key] = value
	}
	return common
}

func applyCredentialCookies(request *http.Request, credentials Credentials) {
	if credentials.MusicID > 0 {
		musicID := credentials.StringMusicID
		if musicID == "" {
			musicID = strconv.FormatInt(credentials.MusicID, 10)
		}
		request.AddCookie(&http.Cookie{Name: "uin", Value: musicID})
		request.AddCookie(&http.Cookie{Name: "qqmusic_uin", Value: musicID})
	}
	if credentials.MusicKey != "" {
		request.AddCookie(&http.Cookie{Name: "qm_keyst", Value: credentials.MusicKey})
		request.AddCookie(&http.Cookie{Name: "qqmusic_key", Value: credentials.MusicKey})
	}
}

func validateCredentials(operation string, credentials Credentials, required bool) error {
	hasID := credentials.MusicID > 0
	hasKey := strings.TrimSpace(credentials.MusicKey) != ""
	if !hasID && !hasKey && !required {
		return nil
	}
	if !hasID || !hasKey {
		return providerError(operation, ErrorKindUnauthorized, "valid login credentials are required", 0)
	}
	return nil
}

func mapMusicUError(operation string, code int) error {
	switch code {
	case 1000, 104400, 104401:
		return providerError(operation, ErrorKindUnauthorized, "login credentials have expired", code)
	case 104003:
		return providerError(operation, ErrorKindForbidden, "account does not have playback permission", code)
	case 104004:
		return providerError(operation, ErrorKindUnavailable, "playback authorization failed", code)
	case 104013:
		return providerError(operation, ErrorKindUnavailable, "playback is restricted on this device", code)
	case 2000:
		return providerError(operation, ErrorKindUpstream, "request signature is required", code)
	case 2001, 104604:
		return providerError(operation, ErrorKindRateLimited, "request was rate limited", code)
	default:
		return providerError(operation, ErrorKindUpstream, "operation was rejected", code)
	}
}

func readResponseBody(reader io.Reader, maximum int64) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(reader, maximum+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maximum {
		return nil, fmt.Errorf("response exceeds limit")
	}
	return body, nil
}

func decodeJSON(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("unexpected trailing JSON")
	}
	return nil
}

func hash33(value string, initial int) int {
	hash := initial
	for _, character := range value {
		hash = ((hash << 5) + hash + int(character)) & 0x7fffffff
	}
	return hash
}

func randomHex(byteCount int) (string, error) {
	value := make([]byte, byteCount)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func randomUUID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}

func setQuery(rawURL string, values url.Values) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsed.RawQuery = values.Encode()
	return parsed.String(), nil
}
