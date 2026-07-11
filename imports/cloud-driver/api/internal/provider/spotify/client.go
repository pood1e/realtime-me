package spotify

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	accountsAuthorizeEndpoint = "https://accounts.spotify.com/authorize"
	accountsTokenEndpoint     = "https://accounts.spotify.com/api/token"
	webAPIEndpoint            = "https://api.spotify.com/v1"

	requestTimeout  = 10 * time.Second
	maximumBodySize = 4 << 20
)

var (
	ErrInvalidConfiguration = errors.New("invalid Spotify configuration")
	ErrInvalidCredentials   = errors.New("invalid Spotify credentials")
	ErrInvalidLoginAttempt  = errors.New("invalid Spotify login attempt")
)

type Client struct {
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
}

func New(config Config) (*Client, error) {
	clientID := strings.TrimSpace(config.ClientID)
	redirectURI := strings.TrimSpace(config.RedirectURI)
	if clientID == "" {
		return nil, fmt.Errorf("%w: client ID is required", ErrInvalidConfiguration)
	}
	if err := validateRedirectURI(redirectURI); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfiguration, err)
	}

	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("%w: standard HTTP transport is unavailable", ErrInvalidConfiguration)
	}
	transport := defaultTransport.Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}

	return &Client{
		clientID:     clientID,
		clientSecret: config.ClientSecret,
		redirectURI:  redirectURI,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   requestTimeout,
		},
	}, nil
}

func validateRedirectURI(rawURI string) error {
	if rawURI == "" {
		return errors.New("redirect URI is required")
	}

	parsed, err := url.ParseRequestURI(rawURI)
	if err != nil || parsed.Host == "" {
		return errors.New("redirect URI is invalid")
	}
	if parsed.Fragment != "" {
		return errors.New("redirect URI must not contain a fragment")
	}
	if parsed.Scheme == "https" {
		return nil
	}
	if parsed.Scheme == "http" && (parsed.Hostname() == "127.0.0.1" || parsed.Hostname() == "::1") {
		return nil
	}
	return errors.New("redirect URI must use HTTPS or an HTTP loopback IP address")
}

func (c *Client) newRequest(
	ctx context.Context,
	method string,
	endpoint string,
	body io.Reader,
) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create Spotify request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	return request, nil
}

func (c *Client) authorizedRequest(
	ctx context.Context,
	method string,
	endpoint string,
	credentials Credentials,
) (*http.Request, error) {
	if err := validateAccessCredentials(credentials, time.Now()); err != nil {
		return nil, err
	}

	request, err := c.newRequest(ctx, method, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+credentials.AccessToken)
	return request, nil
}

func validateAccessCredentials(credentials Credentials, now time.Time) error {
	if credentials.AccessToken == "" {
		return fmt.Errorf("%w: access token is required", ErrInvalidCredentials)
	}
	if credentials.TokenType != "" && !strings.EqualFold(credentials.TokenType, "Bearer") {
		return fmt.Errorf("%w: unsupported token type", ErrInvalidCredentials)
	}
	if credentials.ExpiresAt.IsZero() || !now.Before(credentials.ExpiresAt) {
		return fmt.Errorf("%w: access token has expired", ErrInvalidCredentials)
	}
	return nil
}

// CurrentUser retrieves the stable account_id used to link a Spotify account.
func (c *Client) CurrentUser(ctx context.Context, credentials Credentials) (User, error) {
	request, err := c.authorizedRequest(
		ctx,
		http.MethodGet,
		webAPIEndpoint+"/me",
		credentials,
	)
	if err != nil {
		return User{}, err
	}

	var user User
	if err := c.doJSON(request, &user, credentials.AccessToken); err != nil {
		return User{}, err
	}
	if user.AccountID == "" {
		return User{}, errors.New("Spotify profile response does not contain account_id")
	}
	return user, nil
}

func (c *Client) doJSON(request *http.Request, destination any, redactions ...string) error {
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send Spotify request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, maximumBodySize+1))
	if err != nil {
		return fmt.Errorf("read Spotify response: %w", err)
	}
	if len(body) > maximumBodySize {
		return errors.New("Spotify response exceeds size limit")
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return parseAPIError(response, body, redactions)
	}
	if destination == nil || len(bytes.TrimSpace(body)) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, destination); err != nil {
		return fmt.Errorf("decode Spotify response: %w", err)
	}
	return nil
}

func parseAPIError(response *http.Response, body []byte, redactions []string) error {
	var envelope struct {
		Error            json.RawMessage `json:"error"`
		ErrorDescription string          `json:"error_description"`
		Message          string          `json:"message"`
	}
	_ = json.Unmarshal(body, &envelope)

	code := ""
	message := envelope.ErrorDescription
	if len(envelope.Error) > 0 {
		if envelope.Error[0] == '"' {
			_ = json.Unmarshal(envelope.Error, &code)
		} else {
			var nested struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}
			if json.Unmarshal(envelope.Error, &nested) == nil {
				code = nested.Code
				message = nested.Message
			}
		}
	}
	if message == "" {
		message = envelope.Message
	}
	if code == "" {
		code = strings.ToLower(strings.ReplaceAll(http.StatusText(response.StatusCode), " ", "_"))
	}
	if message == "" {
		message = http.StatusText(response.StatusCode)
	}

	return &APIError{
		StatusCode: response.StatusCode,
		Code:       safeErrorText(code, redactions),
		Message:    safeErrorText(message, redactions),
		RetryAfter: parseRetryAfter(response.Header.Get("Retry-After"), time.Now()),
	}
}

func safeErrorText(value string, redactions []string) string {
	value = strings.TrimSpace(value)
	for _, secret := range redactions {
		if secret != "" {
			value = strings.ReplaceAll(value, secret, "[redacted]")
		}
	}

	value = strings.Map(func(character rune) rune {
		if unicode.IsControl(character) {
			return ' '
		}
		return character
	}, value)
	runes := []rune(value)
	if len(runes) > 512 {
		value = string(runes[:512])
	}
	return strings.TrimSpace(value)
}

func parseRetryAfter(value string, now time.Time) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if retryAt, err := http.ParseTime(value); err == nil && retryAt.After(now) {
		return retryAt.Sub(now)
	}
	return 0
}
