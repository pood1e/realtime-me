package qqmusic

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"
)

const fallbackPlaybackCDN = "https://isure.stream.qqmusic.qq.com/"

type audioFormat struct {
	prefix    string
	extension string
	mimeType  string
}

var browserAudioFormats = map[AudioQuality]audioFormat{
	AudioQualityAAC96:  {prefix: "C400", extension: ".m4a", mimeType: "audio/mp4"},
	AudioQualityAAC192: {prefix: "C600", extension: ".m4a", mimeType: "audio/mp4"},
	AudioQualityMP3128: {prefix: "M500", extension: ".mp3", mimeType: "audio/mpeg"},
	AudioQualityMP3320: {prefix: "M800", extension: ".mp3", mimeType: "audio/mpeg"},
	AudioQualityFLAC:   {prefix: "F000", extension: ".flac", mimeType: "audio/flac"},
}

func (client *Client) GetPlaybackURL(
	ctx context.Context,
	credentials Credentials,
	request PlaybackRequest,
) (PlaybackURL, error) {
	operation := "get playback URL"
	if err := validateCredentials(operation, credentials, false); err != nil {
		return PlaybackURL{}, err
	}
	trackMID := strings.TrimSpace(request.TrackMID)
	if !validIdentifier(trackMID) {
		return PlaybackURL{}, providerError(operation, ErrorKindInvalidInput, "track MID is invalid", 0)
	}
	mediaMID := strings.TrimSpace(request.MediaMID)
	if mediaMID == "" {
		mediaMID = trackMID
	}
	if !validIdentifier(mediaMID) {
		return PlaybackURL{}, providerError(operation, ErrorKindInvalidInput, "media MID is invalid", 0)
	}
	quality := request.Quality
	if quality == "" {
		quality = AudioQualityMP3128
	}
	format, ok := browserAudioFormats[quality]
	if !ok {
		return PlaybackURL{}, providerError(operation, ErrorKindInvalidInput, "audio quality is unsupported", 0)
	}

	guid, err := randomHex(16)
	if err != nil {
		return PlaybackURL{}, providerError(operation, ErrorKindUnavailable, "secure random generation failed", 0)
	}
	cdn, cdnExpiration, err := client.getPlaybackCDN(ctx, credentials, guid)
	if err != nil {
		return PlaybackURL{}, err
	}
	filename := format.prefix + mediaMID + format.extension
	accountID := "0"
	if credentials.MusicID > 0 {
		accountID = stringMusicID(credentials)
	}
	result, err := client.callMusicU(
		ctx,
		operation,
		credentials,
		nil,
		"music.vkey.GetVkey",
		"UrlGetVkey",
		map[string]any{
			"uin":      accountID,
			"filename": []string{filename},
			"guid":     guid,
			"songmid":  []string{trackMID},
			"songtype": []int{0},
			"ctx":      0,
		},
		false,
	)
	if err != nil {
		return PlaybackURL{}, err
	}
	purl, expiration, err := decodeVKey(operation, result.Data)
	if err != nil {
		return PlaybackURL{}, err
	}
	playbackURL, err := resolvePlaybackURL(cdn, purl)
	if err != nil {
		return PlaybackURL{}, providerError(operation, ErrorKindUpstream, "playback URL is invalid", 0)
	}
	if expiration <= 0 {
		expiration = cdnExpiration
	}

	response := PlaybackURL{URL: playbackURL, MIMEType: format.mimeType, Quality: quality}
	if expiration > 0 {
		response.ExpiresAt = time.Now().UTC().Add(time.Duration(expiration) * time.Second)
	}
	return response, nil
}

func (client *Client) getPlaybackCDN(
	ctx context.Context,
	credentials Credentials,
	guid string,
) (string, int64, error) {
	operation := "get playback CDN"
	result, err := client.callMusicU(
		ctx,
		operation,
		credentials,
		nil,
		"music.audioCdnDispatch.cdnDispatch",
		"GetCdnDispatch",
		map[string]any{
			"guid":           guid,
			"uid":            "0",
			"use_new_domain": 1,
			"use_ipv6":       1,
		},
		false,
	)
	if err != nil {
		return "", 0, err
	}
	object, err := decodeObject(result.Data)
	if err != nil {
		return "", 0, providerError(operation, ErrorKindUpstream, "CDN response is invalid", 0)
	}
	domains, err := stringArray(object, "sip")
	if err != nil {
		return "", 0, providerError(operation, ErrorKindUpstream, "CDN list is invalid", 0)
	}
	for _, domain := range domains {
		parsed, parseErr := url.Parse(domain)
		if parseErr == nil && parsed.Scheme == "https" && parsed.Host != "" && !strings.HasPrefix(parsed.Host, "ws.") {
			if !strings.HasSuffix(domain, "/") {
				domain += "/"
			}
			return domain, int64Value(object, "expiration", "expire"), nil
		}
	}
	return fallbackPlaybackCDN, int64Value(object, "expiration", "expire"), nil
}

func decodeVKey(operation string, raw json.RawMessage) (string, int64, error) {
	object, err := decodeObject(raw)
	if err != nil {
		return "", 0, providerError(operation, ErrorKindUpstream, "playback response is invalid", 0)
	}
	items, err := rawArray(object, "midurlinfo")
	if err != nil || len(items) != 1 {
		return "", 0, providerError(operation, ErrorKindUpstream, "playback authorization result is invalid", 0)
	}
	item, err := decodeObject(items[0])
	if err != nil {
		return "", 0, providerError(operation, ErrorKindUpstream, "playback authorization result is invalid", 0)
	}
	resultCode := int(int64Value(item, "result"))
	if resultCode != 0 {
		return "", 0, mapMusicUError(operation, resultCode)
	}
	purl := strings.TrimSpace(stringValue(item, "purl"))
	if purl == "" {
		return "", 0, providerError(operation, ErrorKindNotFound, "playback URL is unavailable for this account", 0)
	}
	return purl, int64Value(object, "expiration", "expire"), nil
}

func resolvePlaybackURL(cdn string, purl string) (string, error) {
	base, err := url.Parse(cdn)
	if err != nil || base.Scheme != "https" || base.Host == "" {
		return "", err
	}
	reference, err := url.Parse(purl)
	if err != nil || reference.User != nil {
		return "", err
	}
	resolved := base.ResolveReference(reference)
	if resolved.Scheme == "http" {
		resolved.Scheme = "https"
	}
	if resolved.Scheme != "https" || resolved.Host == "" {
		return "", url.InvalidHostError("invalid playback host")
	}
	return resolved.String(), nil
}

func stringArray(object map[string]json.RawMessage, key string) ([]string, error) {
	raw, exists := object[key]
	if !exists || string(raw) == "null" {
		return []string{}, nil
	}
	var values []string
	if err := decodeJSON(raw, &values); err != nil {
		return nil, err
	}
	return values, nil
}
