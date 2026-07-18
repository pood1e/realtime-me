package netease

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	providerBaseURL     = "https://music.163.com"
	providerUserAgent   = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/126.0.0.0 Safari/537.36"
	requestTimeout      = 10 * time.Second
	maxResponseBodySize = 8 << 20
)

// Option configures a Client.
type Option func(*clientConfig) error

type clientConfig struct {
	credentials Credentials
}

// WithCredentials restores a previously serialized authenticated session.
func WithCredentials(credentials Credentials) Option {
	return func(config *clientConfig) error {
		if err := validateCredentials(credentials); err != nil {
			return err
		}
		config.credentials = credentials
		return nil
	}
}

// Client is an isolated NetEase WEAPI session.
type Client struct {
	baseURL   *url.URL
	http      *http.Client
	jar       *credentialJar
	publicKey *rsa.PublicKey
}

// NewClient constructs a client with a private cookie jar, standard TLS
// verification, and a ten-second end-to-end request timeout.
func NewClient(options ...Option) (*Client, error) {
	config := clientConfig{}
	for _, option := range options {
		if option == nil {
			return nil, invalidError("create client")
		}
		if err := option(&config); err != nil {
			return nil, err
		}
	}

	baseURL, err := url.Parse(providerBaseURL)
	if err != nil {
		return nil, malformedError("create client")
	}
	jar, err := newCredentialJar()
	if err != nil {
		return nil, malformedError("create client")
	}
	publicKey, err := parseWEAPIPublicKey()
	if err != nil {
		return nil, malformedError("create client")
	}

	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, malformedError("create client")
	}
	transport := defaultTransport.Clone()
	transport.DialContext = (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	transport.TLSHandshakeTimeout = 5 * time.Second
	transport.ResponseHeaderTimeout = 8 * time.Second

	client := &Client{
		baseURL: baseURL,
		jar:     jar,
		http: &http.Client{
			Transport: transport,
			Jar:       jar,
			Timeout:   requestTimeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		publicKey: publicKey,
	}
	if len(config.credentials.Cookies) > 0 {
		if err := client.SetCredentials(config.credentials); err != nil {
			return nil, err
		}
	}
	return client, nil
}

// SetCredentials atomically replaces the client's session cookies.
func (c *Client) SetCredentials(credentials Credentials) error {
	if c == nil || c.jar == nil || c.baseURL == nil {
		return invalidError("restore credentials")
	}
	if err := validateCredentials(credentials); err != nil {
		return err
	}
	if err := c.jar.Replace(c.baseURL, credentials); err != nil {
		return malformedError("restore credentials")
	}
	return nil
}

// Credentials snapshots the current session in a serializable form.
func (c *Client) Credentials() Credentials {
	if c == nil || c.jar == nil || c.baseURL == nil {
		return Credentials{}
	}
	return c.jar.Snapshot(c.baseURL)
}

func (c *Client) postWEAPI(ctx context.Context, operation, path string, payload map[string]any, response any) error {
	if c == nil || c.http == nil || c.baseURL == nil || c.publicKey == nil {
		return invalidError(operation)
	}
	if ctx == nil || !strings.HasPrefix(path, "/weapi/") || response == nil {
		return invalidError(operation)
	}

	requestPayload := make(map[string]any, len(payload)+1)
	for key, value := range payload {
		requestPayload[key] = value
	}
	requestPayload["csrf_token"] = c.csrfToken()
	plaintext, err := json.Marshal(requestPayload)
	if err != nil {
		return invalidError(operation)
	}
	encrypted, err := encryptWEAPI(plaintext, c.publicKey, rand.Reader)
	if err != nil {
		return malformedError(operation)
	}

	form := url.Values{}
	form.Set("params", encrypted.Params)
	form.Set("encSecKey", encrypted.EncSecKey)
	requestURL := c.baseURL.ResolveReference(&url.URL{Path: path})
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return invalidError(operation)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Origin", providerBaseURL)
	request.Header.Set("Referer", providerBaseURL+"/")
	request.Header.Set("User-Agent", providerUserAgent)

	httpResponse, err := c.http.Do(request)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return context.Canceled
		}
		var networkError net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &networkError) && networkError.Timeout()) {
			return &ProviderError{Operation: operation, Kind: ErrorKindTimeout}
		}
		return &ProviderError{Operation: operation, Kind: ErrorKindTransport}
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		kind := ErrorKindHTTP
		if httpResponse.StatusCode == http.StatusNotFound {
			kind = ErrorKindNotFound
		} else if httpResponse.StatusCode == http.StatusTooManyRequests {
			kind = ErrorKindRateLimited
		}
		return &ProviderError{Operation: operation, Kind: kind, HTTPStatus: httpResponse.StatusCode}
	}

	body, err := io.ReadAll(io.LimitReader(httpResponse.Body, maxResponseBodySize+1))
	if err != nil {
		return malformedError(operation)
	}
	if len(body) == 0 || len(body) > maxResponseBodySize || json.Unmarshal(body, response) != nil {
		return malformedError(operation)
	}
	return nil
}

func (c *Client) csrfToken() string {
	for _, cookie := range c.Credentials().Cookies {
		if cookie.Name == "__csrf" {
			return cookie.Value
		}
	}
	return ""
}

func validateCredentials(credentials Credentials) error {
	seen := make(map[string]struct{}, len(credentials.Cookies))
	for _, credential := range credentials.Cookies {
		cookie := &http.Cookie{Name: credential.Name, Value: credential.Value}
		if credential.Name == "" || cookie.Valid() != nil {
			return invalidError("restore credentials")
		}
		if _, exists := seen[credential.Name]; exists {
			return invalidError("restore credentials")
		}
		seen[credential.Name] = struct{}{}
	}
	return nil
}

func validateSuccess(operation string, code int) error {
	if code == http.StatusOK {
		return nil
	}
	kind := ErrorKindUpstream
	if code == http.StatusMovedPermanently || code == http.StatusFound {
		kind = ErrorKindUnauthorized
	}
	return &ProviderError{Operation: operation, Kind: kind, UpstreamCode: code}
}

func invalidError(operation string) *ProviderError {
	return &ProviderError{Operation: operation, Kind: ErrorKindInvalid}
}

func malformedError(operation string) *ProviderError {
	return &ProviderError{Operation: operation, Kind: ErrorKindMalformed}
}

type credentialJar struct {
	mu  sync.RWMutex
	jar http.CookieJar
}

func newCredentialJar() (*credentialJar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &credentialJar{jar: jar}, nil
}

func (j *credentialJar) SetCookies(target *url.URL, cookies []*http.Cookie) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.jar.SetCookies(target, cookies)
}

func (j *credentialJar) Cookies(target *url.URL) []*http.Cookie {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.jar.Cookies(target)
}

func (j *credentialJar) Replace(target *url.URL, credentials Credentials) error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	cookies := make([]*http.Cookie, 0, len(credentials.Cookies))
	for _, credential := range credentials.Cookies {
		cookies = append(cookies, &http.Cookie{Name: credential.Name, Value: credential.Value, Path: "/", Secure: true})
	}
	jar.SetCookies(target, cookies)

	j.mu.Lock()
	j.jar = jar
	j.mu.Unlock()
	return nil
}

func (j *credentialJar) Snapshot(target *url.URL) Credentials {
	cookies := j.Cookies(target)
	credentials := Credentials{Cookies: make([]CredentialCookie, 0, len(cookies))}
	for _, cookie := range cookies {
		credentials.Cookies = append(credentials.Cookies, CredentialCookie{Name: cookie.Name, Value: cookie.Value})
	}
	sort.Slice(credentials.Cookies, func(left, right int) bool {
		return credentials.Cookies[left].Name < credentials.Cookies[right].Name
	})
	return credentials
}
