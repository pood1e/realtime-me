package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
	"github.com/pood1e/realtime-me/services/library/internal/provider/qqmusic"
)

func beginQQLogin(ctx context.Context) (domain.ProviderLoginChallenge, error) {
	client, err := qqmusic.NewClient()
	if err != nil {
		return domain.ProviderLoginChallenge{}, mapProviderError(err)
	}
	attempt, err := client.StartLogin(ctx)
	if err != nil {
		return domain.ProviderLoginChallenge{}, mapProviderError(err)
	}
	state, err := json.Marshal(attempt)
	if err != nil {
		return domain.ProviderLoginChallenge{}, fmt.Errorf("%w: encode QQ Music login state", domain.ErrUnavailable)
	}
	return domain.ProviderLoginChallenge{
		QRImage: attempt.QRCode, QRContentType: attempt.MIMEType, State: state, ExpireTime: attempt.ExpiresAt,
	}, nil
}

func pollQQLogin(ctx context.Context, state []byte) (domain.ProviderLoginPoll, error) {
	var attempt qqmusic.LoginAttempt
	if err := json.Unmarshal(state, &attempt); err != nil {
		return domain.ProviderLoginPoll{}, fmt.Errorf("%w: invalid QQ Music login state", domain.ErrInvalidArgument)
	}
	client, err := qqmusic.NewClient()
	if err != nil {
		return domain.ProviderLoginPoll{}, mapProviderError(err)
	}
	result, err := client.PollLogin(ctx, &attempt)
	if err != nil {
		return domain.ProviderLoginPoll{}, mapProviderError(err)
	}
	poll := domain.ProviderLoginPoll{State: state}
	switch result.Status {
	case qqmusic.LoginStatusWaiting:
		poll.Status = domain.ProviderAttemptWaiting
	case qqmusic.LoginStatusScanned:
		poll.Status = domain.ProviderAttemptScanned
	case qqmusic.LoginStatusExpired:
		poll.Status = domain.ProviderAttemptExpired
	case qqmusic.LoginStatusRejected:
		poll.Status = domain.ProviderAttemptRefused
	case qqmusic.LoginStatusSucceeded:
		if result.Credentials == nil {
			return domain.ProviderLoginPoll{}, fmt.Errorf("%w: QQ Music returned incomplete credentials", domain.ErrUnavailable)
		}
		account, err := qqAccount(ctx, client, *result.Credentials)
		if err != nil {
			return domain.ProviderLoginPoll{}, err
		}
		poll.Status = domain.ProviderAttemptConnected
		poll.Account = &account
	default:
		return domain.ProviderLoginPoll{}, fmt.Errorf("%w: unknown QQ Music login status", domain.ErrUnavailable)
	}
	return poll, nil
}

func qqAccount(ctx context.Context, client *qqmusic.Client, credentials qqmusic.Credentials) (domain.ProviderAccount, error) {
	vip, err := client.GetVIP(ctx, credentials)
	if err != nil {
		return domain.ProviderAccount{}, mapProviderError(err)
	}
	encoded, err := json.Marshal(credentials)
	if err != nil {
		return domain.ProviderAccount{}, fmt.Errorf("%w: encode QQ Music credentials", domain.ErrUnavailable)
	}
	accountID := credentials.StringMusicID
	if accountID == "" {
		accountID = strconv.FormatInt(credentials.MusicID, 10)
	}
	return domain.ProviderAccount{
		AccountID: accountID, DisplayName: "QQ 音乐账号 " + accountID, Membership: qqMembership(vip),
		MembershipExpireTime: qqMembershipExpiry(vip), Credentials: encoded,
	}, nil
}

func qqMembership(info qqmusic.VIPInfo) string {
	switch {
	case info.SuperVIP:
		return "超级会员"
	case info.LuxuryVIP:
		return "豪华绿钻"
	case info.VIP:
		return "绿钻会员"
	default:
		return "普通账号"
	}
}

func qqMembershipExpiry(info qqmusic.VIPInfo) *time.Time {
	if info.ExpiresAt <= 0 {
		return nil
	}
	value := time.Unix(info.ExpiresAt, 0).UTC()
	if info.ExpiresAt > 1_000_000_000_000 {
		value = time.UnixMilli(info.ExpiresAt).UTC()
	}
	return &value
}
