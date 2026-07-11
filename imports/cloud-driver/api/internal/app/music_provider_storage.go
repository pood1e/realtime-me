package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"example.com/cloud-drive/api/internal/domain"
)

func (s *musicProviderSession) providerCredentials(ctx context.Context, provider domain.MusicProvider) (domain.ProviderConnection, []byte, error) {
	adapter, err := s.providerAdapter(provider)
	if err != nil || !adapter.Configured() {
		return domain.ProviderConnection{}, nil, fmt.Errorf("%w: music provider is not configured", domain.ErrConflict)
	}
	connection, err := s.providerStore.GetProviderConnection(ctx, provider)
	if err != nil {
		return domain.ProviderConnection{}, nil, err
	}
	if connection.Status == domain.ProviderReconnectRequired {
		return domain.ProviderConnection{}, nil, domain.ErrProviderReconnectRequired
	}
	credentials, err := s.credentials.Open(domain.MusicProviderCredentialPurpose(provider), connection.EncryptedCredentials)
	return connection, credentials, err
}

func (s *musicProviderSession) persistProviderAccount(ctx context.Context, provider domain.MusicProvider, account domain.ProviderAccount) (domain.ProviderConnection, error) {
	if strings.TrimSpace(account.AccountID) == "" || len(account.Credentials) == 0 {
		return domain.ProviderConnection{}, fmt.Errorf("%w: provider returned an incomplete account", domain.ErrUnavailable)
	}
	encrypted, err := s.credentials.Seal(domain.MusicProviderCredentialPurpose(provider), account.Credentials)
	if err != nil {
		return domain.ProviderConnection{}, err
	}
	now := s.clock.Now().UTC()
	return s.providerStore.UpsertProviderConnection(ctx, domain.ProviderConnection{
		Provider: provider, Status: domain.ProviderConnected, AccountID: strings.TrimSpace(account.AccountID),
		DisplayName: strings.TrimSpace(account.DisplayName), AvatarURL: strings.TrimSpace(account.AvatarURL),
		Membership: strings.TrimSpace(account.Membership), MembershipExpireTime: account.MembershipExpireTime,
		EncryptedCredentials: encrypted, CreateTime: now, UpdateTime: now,
	})
}

func (s *musicProviderSession) persistCredentialUpdate(ctx context.Context, connection domain.ProviderConnection, previous, updated []byte) error {
	if len(updated) == 0 || bytes.Equal(previous, updated) {
		return nil
	}
	encrypted, err := s.credentials.Seal(domain.MusicProviderCredentialPurpose(connection.Provider), updated)
	if err != nil {
		return err
	}
	connection.EncryptedCredentials = encrypted
	connection.Status = domain.ProviderConnected
	connection.UpdateTime = s.clock.Now().UTC()
	_, err = s.providerStore.UpsertProviderConnection(ctx, connection)
	return err
}

func (s *musicProviderSession) markProviderFailure(ctx context.Context, connection domain.ProviderConnection, providerErr error) {
	if !errors.Is(providerErr, domain.ErrProviderReconnectRequired) {
		return
	}
	connection.Status = domain.ProviderReconnectRequired
	connection.UpdateTime = s.clock.Now().UTC()
	_, _ = s.providerStore.UpsertProviderConnection(ctx, connection)
}

func (s *musicProviderSession) providerAdapter(provider domain.MusicProvider) (domain.MusicProviderAdapter, error) {
	adapter, found := s.providers.Get(provider)
	if !found {
		return nil, fmt.Errorf("%w: unknown music provider", domain.ErrInvalidArgument)
	}
	return adapter, nil
}

func providerAttemptTerminal(status domain.ProviderAttemptStatus) bool {
	return status == domain.ProviderAttemptConnected || status == domain.ProviderAttemptExpired ||
		status == domain.ProviderAttemptRefused || status == domain.ProviderAttemptFailed
}

func attemptCredentialPurpose(uid string) string { return "music-provider-attempt:" + uid }
