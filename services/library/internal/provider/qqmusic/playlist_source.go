package qqmusic

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const shortPlaylistPath = "/base/fcgi-bin/u"

// ResolvePlaylistID parses a QQ Music playlist ID or resolves an official share short link.
func (client *Client) ResolvePlaylistID(ctx context.Context, source string) (int64, error) {
	source = strings.TrimSpace(source)
	if id, valid := positivePlaylistID(source); valid {
		return id, nil
	}
	parsed, err := url.ParseRequestURI(source)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.Port() != "" {
		return 0, providerError("resolve playlist source", ErrorKindInvalidInput, "playlist source is invalid", 0)
	}
	host := strings.ToLower(parsed.Hostname())
	if officialPlaylistHost(host) {
		if id, valid := playlistIDFromURL(parsed); valid {
			return id, nil
		}
	}
	if (host == "c6.y.qq.com" || host == "c.y.qq.com") && parsed.Path == shortPlaylistPath {
		return client.resolveShortPlaylistID(ctx, parsed)
	}
	return 0, providerError("resolve playlist source", ErrorKindInvalidInput, "playlist ID is missing", 0)
}

func (client *Client) resolveShortPlaylistID(ctx context.Context, source *url.URL) (int64, error) {
	operation := "resolve playlist share link"
	token := source.Query().Get("__")
	if !validShareToken(token) {
		return 0, providerError(operation, ErrorKindInvalidInput, "share token is invalid", 0)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source.String(), nil)
	if err != nil {
		return 0, providerError(operation, ErrorKindInvalidInput, "share request is invalid", 0)
	}
	client.applyCommonHeaders(request)
	response, err := client.httpClient.Do(request)
	if err != nil {
		return 0, providerError(operation, ErrorKindNetwork, "share request failed", 0)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusMultipleChoices || response.StatusCode > http.StatusPermanentRedirect {
		return 0, providerError(operation, ErrorKindUpstream, "share link did not redirect", response.StatusCode)
	}
	location, err := response.Location()
	if err != nil {
		return 0, providerError(operation, ErrorKindUpstream, "share redirect is invalid", 0)
	}
	if location.Scheme != "https" || !officialPlaylistHost(strings.ToLower(location.Hostname())) || location.User != nil {
		return 0, providerError(operation, ErrorKindUpstream, "share redirect is untrusted", 0)
	}
	if id, valid := playlistIDFromURL(location); valid {
		return id, nil
	}
	return 0, providerError(operation, ErrorKindUpstream, "share redirect is missing the playlist ID", 0)
}

func playlistIDFromURL(source *url.URL) (int64, bool) {
	for _, key := range []string{"id", "disstid"} {
		if id, valid := positivePlaylistID(source.Query().Get(key)); valid {
			return id, true
		}
	}
	segments := strings.Split(strings.Trim(source.Path, "/"), "/")
	for index, segment := range segments {
		if (segment == "playlist" || segment == "playsquare") && index+1 < len(segments) {
			if id, valid := positivePlaylistID(segments[index+1]); valid {
				return id, true
			}
		}
	}
	return 0, false
}

func positivePlaylistID(value string) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return id, err == nil && id > 0
}

func officialPlaylistHost(host string) bool {
	return host == "y.qq.com" || host == "i.y.qq.com" || host == "c.y.qq.com"
}

func validShareToken(value string) bool {
	if len(value) < 1 || len(value) > 128 {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '-' || character == '_' {
			continue
		}
		return false
	}
	return true
}
