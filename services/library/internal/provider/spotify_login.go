package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
	"github.com/pood1e/realtime-me/services/library/internal/provider/spotify"
)

func (a *SpotifyAdapter) beginSpotifyLogin() (domain.ProviderLoginChallenge, error) {
	if a.client == nil {
		return domain.ProviderLoginChallenge{}, fmt.Errorf("%w: Spotify is not configured", domain.ErrConflict)
	}
	authorizationURL, attempt, err := a.client.AuthorizationURL()
	if err != nil {
		return domain.ProviderLoginChallenge{}, mapProviderError(err)
	}
	state, err := json.Marshal(attempt)
	if err != nil {
		return domain.ProviderLoginChallenge{}, fmt.Errorf("%w: encode Spotify login state", domain.ErrUnavailable)
	}
	return domain.ProviderLoginChallenge{
		AuthorizationURL: authorizationURL, OAuthState: attempt.State, State: state, ExpireTime: attempt.ExpiresAt,
	}, nil
}

func (a *SpotifyAdapter) completeSpotifyLogin(ctx context.Context, code string, state []byte) (domain.ProviderAccount, error) {
	var attempt spotify.LoginAttempt
	if err := json.Unmarshal(state, &attempt); err != nil {
		return domain.ProviderAccount{}, fmt.Errorf("%w: invalid Spotify login state", domain.ErrInvalidArgument)
	}
	credentials, err := a.client.ExchangeCode(ctx, code, attempt.State, attempt)
	if err != nil {
		return domain.ProviderAccount{}, mapProviderError(err)
	}
	user, err := a.client.CurrentUser(ctx, credentials)
	if err != nil {
		return domain.ProviderAccount{}, mapProviderError(err)
	}
	encoded, err := json.Marshal(credentials)
	if err != nil {
		return domain.ProviderAccount{}, fmt.Errorf("%w: encode Spotify credentials", domain.ErrUnavailable)
	}
	return domain.ProviderAccount{
		AccountID: user.AccountID, DisplayName: user.DisplayName, Membership: "播放资格待 SDK 验证", Credentials: encoded,
	}, nil
}
