package qqmusic

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
)

func (client *Client) RefreshCredentials(ctx context.Context, credentials Credentials) (Credentials, error) {
	operation := "refresh credentials"
	if err := validateCredentials(operation, credentials, true); err != nil {
		return Credentials{}, err
	}

	params := map[string]any{
		"openid":        credentials.OpenID,
		"refresh_token": credentials.RefreshToken,
		"musickey":      credentials.MusicKey,
		"refresh_key":   credentials.RefreshKey,
		"loginMode":     2,
	}
	switch credentials.LoginType {
	case 1:
		params["str_musicid"] = stringMusicID(credentials)
		params["unionid"] = credentials.UnionID
	case 2:
		params["access_token"] = credentials.AccessToken
		params["expired_in"] = credentials.ExpiresAt
		params["musicid"] = credentials.MusicID
	default:
		params["access_token"] = credentials.AccessToken
		params["expired_in"] = credentials.ExpiresAt
		params["str_musicid"] = stringMusicID(credentials)
		params["musicid"] = credentials.MusicID
		params["unionid"] = credentials.UnionID
	}

	result, err := client.callMusicU(
		ctx,
		operation,
		credentials,
		map[string]any{"tmeLoginType": credentials.LoginType},
		"music.login.LoginServer",
		"Login",
		params,
		true,
	)
	if err != nil {
		return Credentials{}, err
	}
	if result.Code != 0 {
		return Credentials{}, mapLoginError(operation, result.Code)
	}
	return decodeCredentials(operation, result.Data)
}

func (client *Client) GetVIP(ctx context.Context, credentials Credentials) (VIPInfo, error) {
	operation := "get VIP status"
	if err := validateCredentials(operation, credentials, true); err != nil {
		return VIPInfo{}, err
	}
	result, err := client.callMusicU(
		ctx,
		operation,
		credentials,
		nil,
		"VipLogin.VipLoginInter",
		"vip_login_base",
		map[string]any{},
		false,
	)
	if err != nil {
		return VIPInfo{}, err
	}
	return decodeVIPInfo(operation, result.Data)
}

func decodeCredentials(operation string, raw json.RawMessage) (Credentials, error) {
	object, err := decodeObject(raw)
	if err != nil {
		return Credentials{}, providerError(operation, ErrorKindUpstream, "credential response is invalid", 0)
	}
	credentials := Credentials{
		OpenID:            stringValue(object, "openid", "openId"),
		RefreshToken:      stringValue(object, "refresh_token", "refreshToken"),
		AccessToken:       stringValue(object, "access_token", "accessToken"),
		ExpiresAt:         int64Value(object, "expired_at", "expired_in", "expiredAt"),
		MusicID:           int64Value(object, "musicid", "music_id", "musicId"),
		MusicKey:          stringValue(object, "musickey", "music_key", "musicKey"),
		UnionID:           stringValue(object, "unionid", "union_id", "unionId"),
		StringMusicID:     stringValue(object, "str_musicid", "strMusicid", "strMusicID"),
		RefreshKey:        stringValue(object, "refresh_key", "refreshKey"),
		MusicKeyCreatedAt: int64Value(object, "musickeyCreateTime", "music_key_created_at"),
		KeyExpiresIn:      int64Value(object, "keyExpiresIn", "key_expires_in"),
		EncryptedUIN:      stringValue(object, "encryptUin", "encrypted_uin"),
		LoginType:         int(int64Value(object, "loginType", "login_type")),
	}
	if credentials.StringMusicID == "" && credentials.MusicID > 0 {
		credentials.StringMusicID = strconv.FormatInt(credentials.MusicID, 10)
	}
	if credentials.LoginType == 0 && credentials.MusicKey != "" {
		credentials.LoginType = 2
		if strings.HasPrefix(credentials.MusicKey, "W_X") {
			credentials.LoginType = 1
		}
	}
	if !credentials.Valid() {
		return Credentials{}, providerError(operation, ErrorKindUpstream, "credential response is incomplete", 0)
	}
	return credentials, nil
}

func decodeVIPInfo(operation string, raw json.RawMessage) (VIPInfo, error) {
	object, err := decodeObject(raw)
	if err != nil {
		return VIPInfo{}, providerError(operation, ErrorKindUpstream, "VIP response is invalid", 0)
	}
	identity, _ := nestedObject(object, "identity")
	userInfo, _ := nestedObject(object, "userinfo", "userInfo")
	if len(identity) == 0 && len(userInfo) == 0 && !hasAnyField(object, "svip", "star", "ystar") {
		return VIPInfo{}, providerError(operation, ErrorKindUpstream, "VIP response is incomplete", 0)
	}
	return VIPInfo{
		VIP:             int64Value(identity, "vip") > 0,
		LuxuryVIP:       int64Value(identity, "HugeVip", "huge_vip") > 0,
		SuperVIP:        int64Value(object, "svip") > 0,
		Annual:          int64Value(identity, "yearflag", "HugeYearFlag", "huge_year_flag") > 0 || int64Value(object, "ystar") > 0,
		Level:           int(int64Value(identity, "level")),
		ExpiresAt:       int64Value(userInfo, "expire", "expires_at"),
		LuxuryExpiresAt: stringValue(identity, "HugeVipEnd", "huge_vip_end"),
		IconURL:         stringValue(identity, "icon"),
	}, nil
}

func mapLoginError(operation string, code int) error {
	switch code {
	case 1000, 104400, 104401:
		return providerError(operation, ErrorKindUnauthorized, "login credentials have expired", code)
	case 20271:
		return providerError(operation, ErrorKindUnauthorized, "verification code is invalid", code)
	case 20272, 20274:
		return providerError(operation, ErrorKindUnauthorized, "account binding is incomplete", code)
	case 20277, 20278, 20450:
		return providerError(operation, ErrorKindUnauthorized, "account login is restricted", code)
	case 20279:
		return providerError(operation, ErrorKindUnauthorized, "account device limit was reached", code)
	case 104604:
		return providerError(operation, ErrorKindUnavailable, "login was rate limited", code)
	default:
		return providerError(operation, ErrorKindUpstream, "login was rejected", code)
	}
}

func stringMusicID(credentials Credentials) string {
	if credentials.StringMusicID != "" {
		return credentials.StringMusicID
	}
	return strconv.FormatInt(credentials.MusicID, 10)
}
