package qqmusic

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	qqLoginAppID     = "716027609"
	qqMusicOpenAppID = "100497308"
	qqLoginDAID      = "383"
	qrCodeLifetime   = 3 * time.Minute
)

var pngSignature = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}

func (client *Client) StartLogin(ctx context.Context) (*LoginAttempt, error) {
	operation := "start login"
	cacheBuster, err := randomFraction()
	if err != nil {
		return nil, providerError(operation, ErrorKindUnavailable, "secure random generation failed", 0)
	}
	requestURL, err := setQuery("https://ssl.ptlogin2.qq.com/ptqrshow", url.Values{
		"appid":      {qqLoginAppID},
		"e":          {"2"},
		"l":          {"M"},
		"s":          {"3"},
		"d":          {"72"},
		"v":          {"4"},
		"t":          {cacheBuster},
		"daid":       {qqLoginDAID},
		"pt_3rd_aid": {qqMusicOpenAppID},
	})
	if err != nil {
		return nil, providerError(operation, ErrorKindInvalidInput, "request URL is invalid", 0)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, providerError(operation, ErrorKindInvalidInput, "request creation failed", 0)
	}
	request.Header.Set("Referer", "https://xui.ptlogin2.qq.com/")
	client.applyCommonHeaders(request)

	response, err := client.execute(operation, request, http.StatusOK)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	image, err := readResponseBody(response.Body, 2<<20)
	if err != nil || !bytes.HasPrefix(image, pngSignature) {
		return nil, providerError(operation, ErrorKindUpstream, "QR image is invalid", 0)
	}
	qrsig := responseCookie(response, "qrsig")
	if qrsig == "" || len(qrsig) > 1024 {
		return nil, providerError(operation, ErrorKindUpstream, "login state is missing", 0)
	}

	now := time.Now().UTC()
	return &LoginAttempt{
		QRCode:    image,
		MIMEType:  "image/png",
		QRSig:     qrsig,
		CreatedAt: now,
		ExpiresAt: now.Add(qrCodeLifetime),
	}, nil
}

func (client *Client) PollLogin(ctx context.Context, attempt *LoginAttempt) (LoginResult, error) {
	operation := "poll login"
	if err := validateLoginAttempt(attempt); err != nil {
		return LoginResult{}, err
	}
	if time.Now().After(attempt.ExpiresAt) {
		return LoginResult{Status: LoginStatusExpired}, nil
	}

	loginURL, _ := url.Parse("https://ssl.ptlogin2.qq.com/")
	client.httpClient.Jar.SetCookies(loginURL, []*http.Cookie{{
		Name:     "qrsig",
		Value:    attempt.QRSig,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
	}})

	requestURL, err := setQuery("https://ssl.ptlogin2.qq.com/ptqrlogin", url.Values{
		"u1":         {"https://graph.qq.com/oauth2.0/login_jump"},
		"ptqrtoken":  {strconv.Itoa(hash33(attempt.QRSig, 0))},
		"ptredirect": {"0"},
		"h":          {"1"},
		"t":          {"1"},
		"g":          {"1"},
		"from_ui":    {"1"},
		"ptlang":     {"2052"},
		"action":     {fmt.Sprintf("0-0-%d", time.Now().UnixMilli())},
		"js_ver":     {"20102616"},
		"js_type":    {"1"},
		"pt_uistyle": {"40"},
		"aid":        {qqLoginAppID},
		"daid":       {qqLoginDAID},
		"pt_3rd_aid": {qqMusicOpenAppID},
		"has_onekey": {"1"},
	})
	if err != nil {
		return LoginResult{}, providerError(operation, ErrorKindInvalidInput, "request URL is invalid", 0)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return LoginResult{}, providerError(operation, ErrorKindInvalidInput, "request creation failed", 0)
	}
	request.Header.Set("Referer", "https://xui.ptlogin2.qq.com/")
	client.applyCommonHeaders(request)

	response, err := client.execute(operation, request, http.StatusOK)
	if err != nil {
		return LoginResult{}, err
	}
	defer response.Body.Close()
	body, err := readResponseBody(response.Body, 64<<10)
	if err != nil {
		return LoginResult{}, providerError(operation, ErrorKindUpstream, "login response is invalid", 0)
	}

	arguments, err := parsePTUICallback(string(body))
	if err != nil || len(arguments) == 0 {
		return LoginResult{}, providerError(operation, ErrorKindUpstream, "login response cannot be parsed", 0)
	}
	statusCode, err := strconv.Atoi(arguments[0])
	if err != nil {
		return LoginResult{}, providerError(operation, ErrorKindUpstream, "login response has an invalid status", 0)
	}

	switch statusCode {
	case 66:
		return LoginResult{Status: LoginStatusWaiting}, nil
	case 67:
		return LoginResult{Status: LoginStatusScanned}, nil
	case 65:
		return LoginResult{Status: LoginStatusExpired}, nil
	case 68:
		return LoginResult{Status: LoginStatusRejected}, nil
	case 0:
		if len(arguments) < 3 {
			return LoginResult{}, providerError(operation, ErrorKindUpstream, "login response is incomplete", 0)
		}
		credentials, authErr := client.authorizeQQLogin(ctx, arguments[2])
		if authErr != nil {
			return LoginResult{}, authErr
		}
		return LoginResult{Status: LoginStatusSucceeded, Credentials: &credentials}, nil
	default:
		return LoginResult{}, providerError(operation, ErrorKindUpstream, "login operation was rejected", statusCode)
	}
}

func (client *Client) authorizeQQLogin(ctx context.Context, callbackURL string) (Credentials, error) {
	operation := "authorize login"
	parsedCallback, err := url.Parse(callbackURL)
	if err != nil {
		return Credentials{}, providerError(operation, ErrorKindUpstream, "login callback is invalid", 0)
	}
	uin := parsedCallback.Query().Get("uin")
	sigx := parsedCallback.Query().Get("ptsigx")
	if uin == "" || sigx == "" || len(uin) > 64 || len(sigx) > 2048 {
		return Credentials{}, providerError(operation, ErrorKindUpstream, "login callback is incomplete", 0)
	}

	checkURL, err := setQuery("https://ssl.ptlogin2.graph.qq.com/check_sig", url.Values{
		"uin":            {uin},
		"pttype":         {"1"},
		"service":        {"ptqrlogin"},
		"nodirect":       {"0"},
		"ptsigx":         {sigx},
		"s_url":          {"https://graph.qq.com/oauth2.0/login_jump"},
		"ptlang":         {"2052"},
		"ptredirect":     {"100"},
		"aid":            {qqLoginAppID},
		"daid":           {qqLoginDAID},
		"j_later":        {"0"},
		"low_login_hour": {"0"},
		"regmaster":      {"0"},
		"pt_login_type":  {"3"},
		"pt_aid":         {"0"},
		"pt_aaid":        {"16"},
		"pt_light":       {"0"},
		"pt_3rd_aid":     {qqMusicOpenAppID},
	})
	if err != nil {
		return Credentials{}, providerError(operation, ErrorKindInvalidInput, "authorization URL is invalid", 0)
	}
	checkRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, checkURL, nil)
	if err != nil {
		return Credentials{}, providerError(operation, ErrorKindInvalidInput, "authorization request creation failed", 0)
	}
	checkRequest.Header.Set("Referer", "https://xui.ptlogin2.qq.com/")
	client.applyCommonHeaders(checkRequest)
	checkResponse, err := client.execute(operation, checkRequest, http.StatusOK, http.StatusFound)
	if err != nil {
		return Credentials{}, err
	}
	io.Copy(io.Discard, io.LimitReader(checkResponse.Body, 64<<10))
	checkResponse.Body.Close()

	pSKey := responseCookie(checkResponse, "p_skey")
	if pSKey == "" {
		graphURL, _ := url.Parse("https://graph.qq.com/")
		pSKey = cookieValue(client.httpClient.Jar.Cookies(graphURL), "p_skey")
	}
	if pSKey == "" || len(pSKey) > 4096 {
		return Credentials{}, providerError(operation, ErrorKindUpstream, "authorization state is missing", 0)
	}

	ui, err := randomUUID()
	if err != nil {
		return Credentials{}, providerError(operation, ErrorKindUnavailable, "secure random generation failed", 0)
	}
	form := url.Values{
		"response_type": {"code"},
		"client_id":     {qqMusicOpenAppID},
		"redirect_uri":  {"https://y.qq.com/portal/wx_redirect.html?login_type=1&surl=https://y.qq.com/"},
		"scope":         {"get_user_info,get_app_friends"},
		"state":         {"state"},
		"switch":        {""},
		"from_ptlogin":  {"1"},
		"src":           {"1"},
		"update_auth":   {"1"},
		"openapi":       {"1010_1030"},
		"g_tk":          {strconv.Itoa(hash33(pSKey, 5381))},
		"auth_time":     {strconv.FormatInt(time.Now().UnixMilli(), 10)},
		"ui":            {ui},
	}
	authorizeRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://graph.qq.com/oauth2.0/authorize",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return Credentials{}, providerError(operation, ErrorKindInvalidInput, "authorization request creation failed", 0)
	}
	authorizeRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	authorizeRequest.Header.Set("Origin", "https://graph.qq.com")
	authorizeRequest.Header.Set("Referer", "https://graph.qq.com/oauth2.0/login_jump")
	client.applyCommonHeaders(authorizeRequest)
	authorizeResponse, err := client.execute(
		operation,
		authorizeRequest,
		http.StatusFound,
		http.StatusSeeOther,
	)
	if err != nil {
		return Credentials{}, err
	}
	io.Copy(io.Discard, io.LimitReader(authorizeResponse.Body, 64<<10))
	authorizeResponse.Body.Close()

	location := authorizeResponse.Header.Get("Location")
	redirect, err := url.Parse(location)
	if err != nil || redirect.Query().Get("code") == "" {
		return Credentials{}, providerError(operation, ErrorKindUpstream, "authorization code is missing", 0)
	}
	code := redirect.Query().Get("code")
	if len(code) > 4096 {
		return Credentials{}, providerError(operation, ErrorKindUpstream, "authorization code is invalid", 0)
	}

	result, err := client.callMusicU(
		ctx,
		operation,
		Credentials{},
		map[string]any{"tmeLoginType": 2},
		"QQConnectLogin.LoginServer",
		"QQLogin",
		map[string]any{"code": code},
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

func validateLoginAttempt(attempt *LoginAttempt) error {
	if attempt == nil || strings.TrimSpace(attempt.QRSig) == "" {
		return providerError("poll login", ErrorKindInvalidInput, "login attempt is invalid", 0)
	}
	if len(attempt.QRSig) > 1024 || attempt.ExpiresAt.IsZero() {
		return providerError("poll login", ErrorKindInvalidInput, "login attempt is invalid", 0)
	}
	return nil
}

func parsePTUICallback(raw string) ([]string, error) {
	value := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(raw), ";"))
	if !strings.HasPrefix(value, "ptuiCB(") || !strings.HasSuffix(value, ")") {
		return nil, fmt.Errorf("unexpected callback")
	}
	value = strings.TrimSuffix(strings.TrimPrefix(value, "ptuiCB("), ")")
	arguments := make([]string, 0, 6)
	for index := 0; index < len(value); {
		for index < len(value) && (value[index] == ' ' || value[index] == ',') {
			index++
		}
		if index >= len(value) {
			break
		}
		if value[index] != '\'' {
			return nil, fmt.Errorf("unquoted callback argument")
		}
		index++
		var argument strings.Builder
		closed := false
		for index < len(value) {
			character := value[index]
			index++
			if character == '\'' {
				closed = true
				break
			}
			if character == '\\' {
				if index >= len(value) {
					return nil, fmt.Errorf("unterminated callback escape")
				}
				escape := value[index]
				index++
				switch escape {
				case 'x':
					if index+2 > len(value) {
						return nil, fmt.Errorf("invalid callback escape")
					}
					decoded, err := strconv.ParseUint(value[index:index+2], 16, 8)
					if err != nil {
						return nil, fmt.Errorf("invalid callback escape")
					}
					argument.WriteByte(byte(decoded))
					index += 2
				case 'u':
					if index+4 > len(value) {
						return nil, fmt.Errorf("invalid callback escape")
					}
					decoded, err := strconv.ParseUint(value[index:index+4], 16, 16)
					if err != nil {
						return nil, fmt.Errorf("invalid callback escape")
					}
					argument.WriteRune(rune(decoded))
					index += 4
				case 'n':
					argument.WriteByte('\n')
				case 'r':
					argument.WriteByte('\r')
				case 't':
					argument.WriteByte('\t')
				default:
					argument.WriteByte(escape)
				}
				continue
			}
			argument.WriteByte(character)
		}
		if !closed {
			return nil, fmt.Errorf("unterminated callback argument")
		}
		arguments = append(arguments, argument.String())
	}
	return arguments, nil
}

func responseCookie(response *http.Response, name string) string {
	return cookieValue(response.Cookies(), name)
}

func cookieValue(cookies []*http.Cookie, name string) string {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

func randomFraction() (string, error) {
	maximum := new(big.Int).Lsh(big.NewInt(1), 62)
	value, err := rand.Int(rand.Reader, maximum)
	if err != nil {
		return "", err
	}
	return "0." + fmt.Sprintf("%018d", value), nil
}
