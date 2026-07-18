package app

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"example.com/cloud-drive/api/internal/domain"
)

const (
	providerAttemptTTL = 5 * time.Minute
	redirectAttemptTTL = 10 * time.Minute
)

// ListProviders returns every registered plugin plus the local catalog.
func (s *MusicProviderService) ListProviders() []domain.ProviderDescriptor {
	descriptors := []domain.ProviderDescriptor{{
		ID: domain.MusicProviderLocal, DisplayName: "本地音乐", Configured: true,
		Capabilities: []domain.MusicProviderCapability{
			domain.MusicProviderCatalogSearch,
			domain.MusicProviderPlayback,
		},
	}}
	for _, adapter := range s.providers.List() {
		descriptors = append(descriptors, domain.ProviderDescriptor{
			ID: adapter.Provider(), DisplayName: adapter.DisplayName(),
			Capabilities: musicProviderCapabilities(adapter), Configured: adapter.Configured(),
		})
	}
	return descriptors
}

func (s *MusicProviderService) ListProviderConnections(ctx context.Context) ([]domain.ProviderConnection, error) {
	stored, err := s.providerStore.ListProviderConnections(ctx)
	if err != nil {
		return nil, err
	}
	byProvider := make(map[domain.MusicProvider]domain.ProviderConnection, len(stored))
	for _, connection := range stored {
		connection.EncryptedCredentials = nil
		byProvider[connection.Provider] = connection
	}
	adapters := s.providers.List()
	connections := make([]domain.ProviderConnection, 0, len(adapters))
	for _, adapter := range adapters {
		provider := adapter.Provider()
		if connection, found := byProvider[provider]; found {
			connection.Capabilities = musicProviderCapabilities(adapter)
			connections = append(connections, connection)
			continue
		}
		status := domain.ProviderDisconnected
		if !adapter.Configured() {
			status = domain.ProviderNotConfigured
		}
		connections = append(connections, domain.ProviderConnection{
			Provider: provider, Status: status, Capabilities: musicProviderCapabilities(adapter),
		})
	}
	return connections, nil
}

// BeginProviderConnection creates a QR or OAuth login attempt.
func (s *MusicProviderService) BeginProviderConnection(ctx context.Context, provider domain.MusicProvider) (domain.ProviderConnectionAttempt, error) {
	adapter, err := s.providerAdapter(provider)
	if err != nil || !adapter.Configured() {
		return domain.ProviderConnectionAttempt{}, fmt.Errorf("%w: music provider is not configured", domain.ErrConflict)
	}
	connector, supported := adapter.(domain.MusicLoginStarter)
	if !supported {
		return domain.ProviderConnectionAttempt{}, fmt.Errorf("%w: music provider cannot connect accounts", domain.ErrConflict)
	}
	challenge, err := connector.BeginLogin(ctx)
	if err != nil {
		return domain.ProviderConnectionAttempt{}, err
	}
	now := s.clock.Now().UTC()
	expireTime := challenge.ExpireTime.UTC()
	if expireTime.IsZero() {
		expireTime = now.Add(providerAttemptTTL)
		if challenge.OAuthState != "" {
			expireTime = now.Add(redirectAttemptTTL)
		}
	}
	uid := uuid.NewString()
	encryptedState, err := s.credentials.Seal(attemptCredentialPurpose(uid), challenge.State)
	if err != nil {
		return domain.ProviderConnectionAttempt{}, err
	}
	attempt := domain.ProviderConnectionAttempt{
		UID: uid, Provider: provider, Status: domain.ProviderAttemptWaiting, QRImage: challenge.QRImage,
		QRContentType: challenge.QRContentType, QRPayload: challenge.QRPayload, AuthorizationURL: challenge.AuthorizationURL,
		EncryptedState: encryptedState, CreateTime: now, UpdateTime: now, ExpireTime: expireTime,
	}
	if challenge.AuthorizationURL != "" || challenge.OAuthState != "" {
		if challenge.AuthorizationURL == "" || challenge.OAuthState == "" {
			return domain.ProviderConnectionAttempt{}, fmt.Errorf("%w: provider returned an incomplete redirect challenge", domain.ErrUnavailable)
		}
		stateHash := sha256.Sum256([]byte(challenge.OAuthState))
		attempt.StateHash = stateHash[:]
	}
	return s.providerStore.CreateProviderConnectionAttempt(ctx, attempt)
}

// GetProviderConnectionAttempt polls one QR flow and commits successful credentials.
func (s *MusicProviderService) GetProviderConnectionAttempt(ctx context.Context, uid string) (domain.ProviderConnectionAttempt, error) {
	attempt, err := s.providerStore.GetProviderConnectionAttempt(ctx, strings.TrimSpace(uid))
	if err != nil {
		return domain.ProviderConnectionAttempt{}, err
	}
	if providerAttemptTerminal(attempt.Status) {
		return attempt, nil
	}
	adapter, adapterErr := s.providerAdapter(attempt.Provider)
	poller, canPoll := adapter.(domain.MusicQRLoginPoller)
	if adapterErr != nil || !canPoll {
		return attempt, nil
	}
	now := s.clock.Now().UTC()
	if !now.Before(attempt.ExpireTime) {
		attempt.Status = domain.ProviderAttemptExpired
		attempt.UpdateTime = now
		return s.providerStore.UpdateProviderConnectionAttempt(ctx, attempt)
	}
	state, err := s.credentials.Open(attemptCredentialPurpose(attempt.UID), attempt.EncryptedState)
	if err != nil {
		return domain.ProviderConnectionAttempt{}, err
	}
	poll, err := poller.PollLogin(ctx, state)
	if err != nil {
		return domain.ProviderConnectionAttempt{}, err
	}
	attempt.Status = poll.Status
	attempt.UpdateTime = now
	if len(poll.State) > 0 {
		attempt.EncryptedState, err = s.credentials.Seal(attemptCredentialPurpose(attempt.UID), poll.State)
		if err != nil {
			return domain.ProviderConnectionAttempt{}, err
		}
	}
	if poll.Account != nil {
		if _, err := s.persistProviderAccount(ctx, attempt.Provider, *poll.Account); err != nil {
			return domain.ProviderConnectionAttempt{}, err
		}
		attempt.Status = domain.ProviderAttemptConnected
		consumedTime := now
		attempt.ConsumedTime = &consumedTime
	}
	return s.providerStore.UpdateProviderConnectionAttempt(ctx, attempt)
}

// CompleteRedirectConnection validates one unauthenticated provider callback.
func (s *MusicProviderService) CompleteRedirectConnection(ctx context.Context, provider domain.MusicProvider, state, code string) error {
	state = strings.TrimSpace(state)
	code = strings.TrimSpace(code)
	if provider == "" || provider == domain.MusicProviderLocal || state == "" || code == "" {
		return fmt.Errorf("%w: missing provider callback parameters", domain.ErrInvalidArgument)
	}
	stateHash := sha256.Sum256([]byte(state))
	attempt, err := s.providerStore.GetProviderConnectionAttemptByStateHash(ctx, stateHash[:])
	if err != nil {
		return err
	}
	if attempt.Provider != provider {
		return fmt.Errorf("%w: provider callback does not match its login attempt", domain.ErrInvalidArgument)
	}
	adapter, err := s.providerAdapter(attempt.Provider)
	if err != nil {
		return err
	}
	completer, supported := adapter.(domain.MusicRedirectLoginCompleter)
	if !supported {
		return fmt.Errorf("%w: provider does not support redirect login", domain.ErrConflict)
	}
	now := s.clock.Now().UTC()
	if err := s.providerStore.ConsumeProviderConnectionAttempt(ctx, attempt.UID, now); err != nil {
		return err
	}
	loginState, err := s.credentials.Open(attemptCredentialPurpose(attempt.UID), attempt.EncryptedState)
	if err != nil {
		return err
	}
	account, err := completer.CompleteRedirectLogin(ctx, code, loginState)
	if err != nil {
		attempt.Status = domain.ProviderAttemptFailed
		attempt.UpdateTime = now
		attempt.ConsumedTime = &now
		_, _ = s.providerStore.UpdateProviderConnectionAttempt(ctx, attempt)
		return err
	}
	if _, err := s.persistProviderAccount(ctx, attempt.Provider, account); err != nil {
		return err
	}
	attempt.Status = domain.ProviderAttemptConnected
	attempt.UpdateTime = now
	attempt.ConsumedTime = &now
	_, err = s.providerStore.UpdateProviderConnectionAttempt(ctx, attempt)
	return err
}

// DisconnectProvider removes long-lived credentials for one source.
func (s *MusicProviderService) DisconnectProvider(ctx context.Context, provider domain.MusicProvider) error {
	if _, err := s.providerAdapter(provider); err != nil {
		return fmt.Errorf("%w: invalid external music provider", domain.ErrInvalidArgument)
	}
	return s.providerStore.DeleteProviderConnection(ctx, provider)
}

// SearchMusic fans out one query while isolating provider failures.
