package spotify

import (
	"context"
	"fmt"
	"time"
)

const playbackRefreshWindow = 2 * time.Minute

// WebPlaybackToken returns a short-lived access token suitable for Spotify's
// browser SDK and the credentials that callers should persist. Credentials are
// refreshed first when their access token is close to expiry.
func (c *Client) WebPlaybackToken(
	ctx context.Context,
	credentials Credentials,
) (PlaybackToken, Credentials, error) {
	if !credentials.HasScope(ScopeStreaming) {
		return PlaybackToken{}, Credentials{}, fmt.Errorf(
			"%w: %s scope is required",
			ErrInvalidCredentials,
			ScopeStreaming,
		)
	}

	now := time.Now()
	if credentials.ExpiresAt.IsZero() || !now.Add(playbackRefreshWindow).Before(credentials.ExpiresAt) {
		if credentials.RefreshToken == "" {
			return PlaybackToken{}, Credentials{}, fmt.Errorf(
				"%w: access token is expired or close to expiry and cannot be refreshed",
				ErrInvalidCredentials,
			)
		}
		refreshed, err := c.Refresh(ctx, credentials)
		if err != nil {
			return PlaybackToken{}, Credentials{}, err
		}
		credentials = refreshed
	}
	if err := validateAccessCredentials(credentials, now); err != nil {
		return PlaybackToken{}, Credentials{}, err
	}
	if !credentials.HasScope(ScopeStreaming) {
		return PlaybackToken{}, Credentials{}, fmt.Errorf(
			"%w: refreshed credentials do not include %s scope",
			ErrInvalidCredentials,
			ScopeStreaming,
		)
	}

	return PlaybackToken{
		AccessToken: credentials.AccessToken,
		TokenType:   "Bearer",
		ExpiresAt:   credentials.ExpiresAt,
	}, credentials, nil
}
