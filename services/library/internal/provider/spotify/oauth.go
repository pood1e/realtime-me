package spotify

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const loginAttemptLifetime = 10 * time.Minute

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
}

// DefaultScopes returns the least set used by this package's browser playback
// flow. Catalog search and account_id do not require additional scopes.
func DefaultScopes() []string {
	return []string{ScopeStreaming, ScopeUserReadPrivate, ScopeUserModifyPlaybackState, ScopePlaylistReadPrivate}
}

// AuthorizationURL creates a one-time Authorization Code with PKCE request.
// The returned LoginAttempt must be persisted until the OAuth callback.
func (c *Client) AuthorizationURL(scopes ...string) (string, LoginAttempt, error) {
	if len(scopes) == 0 {
		scopes = DefaultScopes()
	}
	scopes = normalizedScopes(scopes)

	state, err := randomBase64URL(32)
	if err != nil {
		return "", LoginAttempt{}, fmt.Errorf("generate Spotify OAuth state: %w", err)
	}
	codeVerifier, err := randomBase64URL(64)
	if err != nil {
		return "", LoginAttempt{}, fmt.Errorf("generate Spotify PKCE verifier: %w", err)
	}
	challengeHash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(challengeHash[:])

	parameters := url.Values{
		"response_type":         {"code"},
		"client_id":             {c.clientID},
		"redirect_uri":          {c.redirectURI},
		"state":                 {state},
		"code_challenge_method": {"S256"},
		"code_challenge":        {codeChallenge},
	}
	if len(scopes) > 0 {
		parameters.Set("scope", strings.Join(scopes, " "))
	}

	attempt := LoginAttempt{
		State:        state,
		CodeVerifier: codeVerifier,
		Scopes:       scopes,
		ExpiresAt:    time.Now().Add(loginAttemptLifetime).UTC(),
	}
	return accountsAuthorizeEndpoint + "?" + parameters.Encode(), attempt, nil
}

// ExchangeCode validates the callback state and exchanges its authorization
// code for reusable Spotify credentials.
func (c *Client) ExchangeCode(
	ctx context.Context,
	code string,
	returnedState string,
	attempt LoginAttempt,
) (Credentials, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return Credentials{}, fmt.Errorf("%w: authorization code is required", ErrInvalidLoginAttempt)
	}
	if attempt.CodeVerifier == "" || !attempt.validState(returnedState, time.Now()) {
		return Credentials{}, fmt.Errorf("%w: state is invalid or expired", ErrInvalidLoginAttempt)
	}

	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {c.redirectURI},
		"client_id":     {c.clientID},
		"code_verifier": {attempt.CodeVerifier},
	}
	response, err := c.requestToken(ctx, form, code, attempt.CodeVerifier)
	if err != nil {
		return Credentials{}, err
	}
	return credentialsFromToken(response, "", attempt.Scopes, time.Now())
}

// Refresh exchanges a reusable refresh token for a new short-lived access
// token. Spotify may omit a replacement refresh token and scope list.
func (c *Client) Refresh(ctx context.Context, credentials Credentials) (Credentials, error) {
	if credentials.RefreshToken == "" {
		return Credentials{}, fmt.Errorf("%w: refresh token is required", ErrInvalidCredentials)
	}

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {credentials.RefreshToken},
		"client_id":     {c.clientID},
	}
	response, err := c.requestToken(ctx, form, credentials.RefreshToken)
	if err != nil {
		return Credentials{}, err
	}
	return credentialsFromToken(
		response,
		credentials.RefreshToken,
		credentials.Scopes,
		time.Now(),
	)
}

func (c *Client) requestToken(
	ctx context.Context,
	form url.Values,
	redactions ...string,
) (tokenResponse, error) {
	request, err := c.newRequest(
		ctx,
		http.MethodPost,
		accountsTokenEndpoint,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return tokenResponse{}, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.clientSecret != "" {
		request.SetBasicAuth(c.clientID, c.clientSecret)
		redactions = append(redactions, c.clientSecret)
	}

	var response tokenResponse
	if err := c.doJSON(request, &response, redactions...); err != nil {
		return tokenResponse{}, err
	}
	return response, nil
}

func credentialsFromToken(
	response tokenResponse,
	currentRefreshToken string,
	currentScopes []string,
	now time.Time,
) (Credentials, error) {
	if response.AccessToken == "" || response.ExpiresIn <= 0 {
		return Credentials{}, errors.New("Spotify token response is incomplete")
	}
	tokenType := response.TokenType
	if tokenType == "" {
		tokenType = "Bearer"
	}
	if !strings.EqualFold(tokenType, "Bearer") {
		return Credentials{}, errors.New("Spotify token response uses an unsupported token type")
	}

	refreshToken := response.RefreshToken
	if refreshToken == "" {
		refreshToken = currentRefreshToken
	}
	scopes := normalizedScopes(strings.Fields(response.Scope))
	if len(scopes) == 0 {
		scopes = normalizedScopes(currentScopes)
	}

	return Credentials{
		AccessToken:  response.AccessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		Scopes:       scopes,
		ExpiresAt:    now.Add(time.Duration(response.ExpiresIn) * time.Second).UTC(),
	}, nil
}

func randomBase64URL(byteCount int) (string, error) {
	value := make([]byte, byteCount)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}
