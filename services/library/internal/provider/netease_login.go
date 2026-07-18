package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
	"github.com/pood1e/realtime-me/services/library/internal/provider/netease"
)

func beginNetEaseLogin(ctx context.Context) (domain.ProviderLoginChallenge, error) {
	client, err := netease.NewClient()
	if err != nil {
		return domain.ProviderLoginChallenge{}, mapProviderError(err)
	}
	attempt, err := client.CreateLoginAttempt(ctx)
	if err != nil {
		return domain.ProviderLoginChallenge{}, mapProviderError(err)
	}
	state, err := json.Marshal(attempt)
	if err != nil {
		return domain.ProviderLoginChallenge{}, fmt.Errorf("%w: encode NetEase login state", domain.ErrUnavailable)
	}
	return domain.ProviderLoginChallenge{QRPayload: attempt.LoginURL, State: state, ExpireTime: attempt.ExpiresAt}, nil
}

func pollNetEaseLogin(ctx context.Context, state []byte) (domain.ProviderLoginPoll, error) {
	var attempt netease.LoginAttempt
	if err := json.Unmarshal(state, &attempt); err != nil {
		return domain.ProviderLoginPoll{}, fmt.Errorf("%w: invalid NetEase login state", domain.ErrInvalidArgument)
	}
	client, err := netease.NewClient(netease.WithCredentials(attempt.Credentials))
	if err != nil {
		return domain.ProviderLoginPoll{}, mapProviderError(err)
	}
	status, err := client.CheckLoginAttempt(ctx, attempt)
	if err != nil {
		return domain.ProviderLoginPoll{}, mapProviderError(err)
	}
	attempt.Credentials = client.Credentials()
	updatedState, err := json.Marshal(attempt)
	if err != nil {
		return domain.ProviderLoginPoll{}, fmt.Errorf("%w: encode NetEase login state", domain.ErrUnavailable)
	}
	poll := domain.ProviderLoginPoll{State: updatedState}
	switch status.State {
	case netease.LoginStateWaiting:
		poll.Status = domain.ProviderAttemptWaiting
	case netease.LoginStateScanned:
		poll.Status = domain.ProviderAttemptScanned
	case netease.LoginStateExpired:
		poll.Status = domain.ProviderAttemptExpired
	case netease.LoginStateAuthorized:
		if status.Credentials == nil {
			return domain.ProviderLoginPoll{}, fmt.Errorf("%w: NetEase returned incomplete credentials", domain.ErrUnavailable)
		}
		account, err := netEaseAccount(ctx, *status.Credentials)
		if err != nil {
			return domain.ProviderLoginPoll{}, err
		}
		poll.Status = domain.ProviderAttemptConnected
		poll.Account = &account
	default:
		return domain.ProviderLoginPoll{}, fmt.Errorf("%w: unknown NetEase login status", domain.ErrUnavailable)
	}
	return poll, nil
}

func netEaseAccount(ctx context.Context, credentials netease.Credentials) (domain.ProviderAccount, error) {
	client, err := netease.NewClient(netease.WithCredentials(credentials))
	if err != nil {
		return domain.ProviderAccount{}, mapProviderError(err)
	}
	account, err := client.Account(ctx)
	if err != nil {
		return domain.ProviderAccount{}, mapProviderError(err)
	}
	vip, err := client.VIP(ctx, account.ID)
	if err != nil {
		return domain.ProviderAccount{}, mapProviderError(err)
	}
	encoded, err := json.Marshal(client.Credentials())
	if err != nil {
		return domain.ProviderAccount{}, fmt.Errorf("%w: encode NetEase credentials", domain.ErrUnavailable)
	}
	return domain.ProviderAccount{
		AccountID: strconv.FormatInt(account.ID, 10), DisplayName: account.Profile.Nickname,
		AvatarURL: account.Profile.AvatarURL, Membership: netEaseMembership(vip),
		MembershipExpireTime: netEaseMembershipExpiry(vip), Credentials: encoded,
	}, nil
}

func netEaseMembership(info netease.VIPInfo) string {
	if info.RedVIPLevel > 0 {
		return fmt.Sprintf("黑胶 VIP %d", info.RedVIPLevel)
	}
	return "普通账号"
}

func netEaseMembershipExpiry(info netease.VIPInfo) *time.Time {
	var latest time.Time
	for _, membership := range []*netease.VIPMembership{info.Associator, info.MusicPackage} {
		if membership != nil && membership.ExpiresAt.After(latest) {
			latest = membership.ExpiresAt
		}
	}
	if latest.IsZero() {
		return nil
	}
	return &latest
}
