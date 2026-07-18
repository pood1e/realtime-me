package netease

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	createLoginOperation = "create login attempt"
	checkLoginOperation  = "check login attempt"
	accountOperation     = "get account"
	vipOperation         = "get VIP information"
	loginAttemptLifetime = 5 * time.Minute
)

type vipMembershipResponse struct {
	VIPCode    int   `json:"vipCode"`
	ExpireTime int64 `json:"expireTime"`
}

// CreateLoginAttempt creates a QR-code key and the URL to render.
func (c *Client) CreateLoginAttempt(ctx context.Context) (LoginAttempt, error) {
	var response struct {
		Code   int    `json:"code"`
		UniKey string `json:"unikey"`
	}
	if err := c.postWEAPI(ctx, createLoginOperation, "/weapi/login/qrcode/unikey", map[string]any{"type": 3}, &response); err != nil {
		return LoginAttempt{}, err
	}
	if err := validateSuccess(createLoginOperation, response.Code); err != nil {
		return LoginAttempt{}, err
	}
	key := strings.TrimSpace(response.UniKey)
	if key == "" {
		return LoginAttempt{}, malformedError(createLoginOperation)
	}
	loginURL := providerBaseURL + "/login?codekey=" + url.QueryEscape(key)
	return LoginAttempt{
		Key:         key,
		LoginURL:    loginURL,
		Credentials: c.Credentials(),
		ExpiresAt:   time.Now().Add(loginAttemptLifetime).UTC(),
	}, nil
}

// CheckLoginAttempt restores the attempt session and polls its QR-code state.
func (c *Client) CheckLoginAttempt(ctx context.Context, attempt LoginAttempt) (LoginStatus, error) {
	key := strings.TrimSpace(attempt.Key)
	if key == "" {
		return LoginStatus{}, invalidError(checkLoginOperation)
	}
	if attempt.ExpiresAt.IsZero() {
		return LoginStatus{}, invalidError(checkLoginOperation)
	}
	if !time.Now().Before(attempt.ExpiresAt) {
		return LoginStatus{State: LoginStateExpired}, nil
	}
	if err := c.SetCredentials(attempt.Credentials); err != nil {
		return LoginStatus{}, invalidError(checkLoginOperation)
	}
	var response struct {
		Code      int    `json:"code"`
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatarUrl"`
	}
	payload := map[string]any{"key": key, "type": 3}
	if err := c.postWEAPI(ctx, checkLoginOperation, "/weapi/login/qrcode/client/login", payload, &response); err != nil {
		return LoginStatus{}, err
	}

	status := LoginStatus{Nickname: response.Nickname, AvatarURL: response.AvatarURL}
	switch response.Code {
	case 800:
		status.State = LoginStateExpired
	case 801:
		status.State = LoginStateWaiting
	case 802:
		status.State = LoginStateScanned
	case 803:
		credentials := c.Credentials()
		if !credentials.hasCookie("MUSIC_U") {
			return LoginStatus{}, malformedError(checkLoginOperation)
		}
		status.State = LoginStateAuthorized
		status.Credentials = &credentials
	default:
		return LoginStatus{}, &ProviderError{Operation: checkLoginOperation, Kind: ErrorKindUpstream, UpstreamCode: response.Code}
	}
	return status, nil
}

// Account returns the currently authenticated account.
func (c *Client) Account(ctx context.Context) (Account, error) {
	var response struct {
		Code    int `json:"code"`
		Account *struct {
			ID       int64  `json:"id"`
			Username string `json:"userName"`
			Status   int    `json:"status"`
			VIPType  int    `json:"vipType"`
		} `json:"account"`
		Profile *struct {
			UserID    int64  `json:"userId"`
			Nickname  string `json:"nickname"`
			AvatarURL string `json:"avatarUrl"`
			VIPType   int    `json:"vipType"`
		} `json:"profile"`
	}
	if err := c.postWEAPI(ctx, accountOperation, "/weapi/nuser/account/get", map[string]any{}, &response); err != nil {
		return Account{}, err
	}
	if err := validateSuccess(accountOperation, response.Code); err != nil {
		return Account{}, err
	}
	if response.Account == nil || response.Profile == nil || response.Account.ID <= 0 || response.Profile.UserID != response.Account.ID || strings.TrimSpace(response.Profile.Nickname) == "" {
		return Account{}, malformedError(accountOperation)
	}
	return Account{
		ID:       response.Account.ID,
		Username: response.Account.Username,
		Status:   response.Account.Status,
		VIPType:  response.Account.VIPType,
		Profile: Profile{
			UserID:    response.Profile.UserID,
			Nickname:  response.Profile.Nickname,
			AvatarURL: response.Profile.AvatarURL,
			VIPType:   response.Profile.VIPType,
		},
	}, nil
}

// VIP returns the membership state for an authenticated user.
func (c *Client) VIP(ctx context.Context, userID int64) (VIPInfo, error) {
	if userID <= 0 {
		return VIPInfo{}, invalidError(vipOperation)
	}
	var response struct {
		Code int `json:"code"`
		Data *struct {
			UserID            int64                  `json:"userId"`
			RedVIPLevel       int                    `json:"redVipLevel"`
			RedVIPAnnualCount int                    `json:"redVipAnnualCount"`
			Now               int64                  `json:"now"`
			Associator        *vipMembershipResponse `json:"associator"`
			MusicPackage      *vipMembershipResponse `json:"musicPackage"`
		} `json:"data"`
	}
	payload := map[string]any{"userId": strconv.FormatInt(userID, 10)}
	if err := c.postWEAPI(ctx, vipOperation, "/weapi/music-vip-membership/client/vip/info", payload, &response); err != nil {
		return VIPInfo{}, err
	}
	if err := validateSuccess(vipOperation, response.Code); err != nil {
		return VIPInfo{}, err
	}
	if response.Data == nil || (response.Data.UserID != 0 && response.Data.UserID != userID) || response.Data.RedVIPLevel < 0 {
		return VIPInfo{}, malformedError(vipOperation)
	}
	if !validMembership(response.Data.Associator) || !validMembership(response.Data.MusicPackage) {
		return VIPInfo{}, malformedError(vipOperation)
	}
	info := VIPInfo{
		UserID:         userID,
		RedVIPLevel:    response.Data.RedVIPLevel,
		RedAnnualCount: response.Data.RedVIPAnnualCount,
		ServerTime:     millisecondsToTime(response.Data.Now),
		Associator:     normalizeMembership(response.Data.Associator),
		MusicPackage:   normalizeMembership(response.Data.MusicPackage),
	}
	return info, nil
}

func normalizeMembership(membership *vipMembershipResponse) *VIPMembership {
	if membership == nil || (membership.VIPCode == 0 && membership.ExpireTime == 0) {
		return nil
	}
	return &VIPMembership{Code: membership.VIPCode, ExpiresAt: millisecondsToTime(membership.ExpireTime)}
}

func validMembership(membership *vipMembershipResponse) bool {
	return membership == nil || (membership.VIPCode >= 0 && membership.ExpireTime >= 0)
}

func millisecondsToTime(value int64) time.Time {
	if value <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(value).UTC()
}

func (credentials Credentials) hasCookie(name string) bool {
	for _, cookie := range credentials.Cookies {
		if cookie.Name == name && cookie.Value != "" {
			return true
		}
	}
	return false
}
