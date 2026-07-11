package qqmusic

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"html"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultSearchPageSize = 20
	maximumSearchPageSize = 50
	maximumPageTokenBytes = 32 << 10
	lyricsEndpoint        = "https://c.y.qq.com/lyric/fcgi-bin/fcg_query_lyric_new.fcg"
)

type searchCursor struct {
	SearchID string          `json:"search_id"`
	Page     int             `json:"page"`
	Start    json.RawMessage `json:"start,omitempty"`
}

func (client *Client) Search(ctx context.Context, request SearchRequest) (SearchResponse, error) {
	operation := "search catalog"
	query := strings.TrimSpace(request.Query)
	if query == "" || len([]rune(query)) > 200 {
		return SearchResponse{}, providerError(operation, ErrorKindInvalidInput, "query must contain 1 to 200 characters", 0)
	}
	pageSize := request.PageSize
	if pageSize == 0 {
		pageSize = defaultSearchPageSize
	}
	if pageSize < 1 || pageSize > maximumSearchPageSize {
		return SearchResponse{}, providerError(operation, ErrorKindInvalidInput, "page size must be between 1 and 50", 0)
	}

	cursor, err := initialSearchCursor(request.PageToken)
	if err != nil {
		return SearchResponse{}, providerError(operation, ErrorKindInvalidInput, "page token is invalid", 0)
	}
	if cursor.SearchID == "" {
		cursor.SearchID, err = newSearchID()
		if err != nil {
			return SearchResponse{}, providerError(operation, ErrorKindUnavailable, "secure random generation failed", 0)
		}
	}

	params := map[string]any{
		"searchid":    cursor.SearchID,
		"search_type": 100,
		"page_num":    pageSize,
		"query":       query,
		"page_id":     cursor.Page,
		"highlight":   0,
		"grp":         1,
	}
	if len(cursor.Start) > 0 && string(cursor.Start) != "null" {
		var pageStart any
		if err := decodeJSON(cursor.Start, &pageStart); err != nil {
			return SearchResponse{}, providerError(operation, ErrorKindInvalidInput, "page token is invalid", 0)
		}
		params["page_start"] = pageStart
	}

	result, err := client.callMusicU(
		ctx,
		operation,
		Credentials{},
		nil,
		"music.adaptor.SearchAdaptor",
		"do_search_v2",
		params,
		false,
	)
	if err != nil {
		return SearchResponse{}, err
	}
	return decodeSearchResponse(operation, result.Data, cursor.SearchID)
}

func (client *Client) GetLyrics(ctx context.Context, request LyricsRequest) (Lyrics, error) {
	operation := "get lyrics"
	trackMID := strings.TrimSpace(request.TrackMID)
	if !validIdentifier(trackMID) {
		return Lyrics{}, providerError(operation, ErrorKindInvalidInput, "track MID is invalid", 0)
	}
	requestURL, err := setQuery(lyricsEndpoint, url.Values{
		"songmid":     {trackMID},
		"format":      {"json"},
		"nobase64":    {"1"},
		"g_tk":        {"5381"},
		"loginUin":    {"0"},
		"hostUin":     {"0"},
		"inCharset":   {"utf8"},
		"outCharset":  {"utf-8"},
		"notice":      {"0"},
		"platform":    {"yqq.json"},
		"needNewCode": {"0"},
	})
	if err != nil {
		return Lyrics{}, providerError(operation, ErrorKindInvalidInput, "request URL is invalid", 0)
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return Lyrics{}, providerError(operation, ErrorKindInvalidInput, "request creation failed", 0)
	}
	httpRequest.Header.Set("Referer", "https://y.qq.com/n/ryqq/songDetail/"+trackMID)
	client.applyCommonHeaders(httpRequest)
	response, err := client.execute(operation, httpRequest, http.StatusOK)
	if err != nil {
		return Lyrics{}, err
	}
	defer response.Body.Close()
	body, err := readResponseBody(response.Body, 4<<20)
	if err != nil {
		return Lyrics{}, providerError(operation, ErrorKindUpstream, "lyrics response is invalid", 0)
	}
	object, err := decodeObject(body)
	if err != nil {
		return Lyrics{}, providerError(operation, ErrorKindUpstream, "lyrics response is not valid JSON", 0)
	}
	if code := int64Value(object, "code", "retcode"); code != 0 {
		return Lyrics{}, providerError(operation, ErrorKindUpstream, "lyrics request was rejected", int(code))
	}
	original := decodeMaybeBase64(stringValue(object, "lyric"))
	if strings.TrimSpace(original) == "" {
		return Lyrics{}, providerError(operation, ErrorKindNotFound, "lyrics are unavailable", 0)
	}
	return Lyrics{
		TrackMID:    trackMID,
		Original:    original,
		Translation: decodeMaybeBase64(stringValue(object, "trans", "translation")),
		Romanized:   decodeMaybeBase64(stringValue(object, "roma", "romanized")),
	}, nil
}

func decodeSearchResponse(operation string, raw json.RawMessage, fallbackSearchID string) (SearchResponse, error) {
	root, err := decodeObject(raw)
	if err != nil {
		return SearchResponse{}, providerError(operation, ErrorKindUpstream, "search response is invalid", 0)
	}
	meta, ok := nestedObject(root, "meta")
	if !ok {
		return SearchResponse{}, providerError(operation, ErrorKindUpstream, "search metadata is missing", 0)
	}
	body, ok := nestedObject(root, "body")
	if !ok {
		return SearchResponse{}, providerError(operation, ErrorKindUpstream, "search result body is missing", 0)
	}
	itemSong, ok := nestedObject(body, "item_song")
	if !ok {
		return SearchResponse{}, providerError(operation, ErrorKindUpstream, "track search results are missing", 0)
	}

	items, err := rawArray(itemSong, "items")
	if err != nil {
		return SearchResponse{}, providerError(operation, ErrorKindUpstream, "track search results are invalid", 0)
	}
	tracks := make([]Track, 0, len(items))
	for _, item := range items {
		track, parseErr := decodeTrack(item)
		if parseErr != nil {
			return SearchResponse{}, providerError(operation, ErrorKindUpstream, "track search result is invalid", 0)
		}
		tracks = append(tracks, track)
	}

	searchID := stringValue(meta, "sid", "searchid")
	if searchID == "" {
		searchID = fallbackSearchID
	}
	nextPage := int(int64Value(meta, "nextpage"))
	nextToken := ""
	if nextPage > 0 {
		cursor := searchCursor{SearchID: searchID, Page: nextPage}
		if start, exists := meta["nextpage_start"]; exists && len(start) > 0 && string(start) != "null" {
			cursor.Start = start
		}
		nextToken, err = encodePageToken(cursor)
		if err != nil {
			return SearchResponse{}, providerError(operation, ErrorKindUpstream, "search cursor is invalid", 0)
		}
	}

	total := int64Value(itemSong, "total_num", "estimate_sum")
	if total == 0 {
		total = int64Value(meta, "sum", "total_num", "estimate_sum")
	}
	return SearchResponse{Tracks: tracks, Total: total, NextPageToken: nextToken}, nil
}

func decodeTrack(raw json.RawMessage) (Track, error) {
	object, err := decodeObject(raw)
	if err != nil {
		return Track{}, err
	}
	mid := stringValue(object, "mid", "songmid", "songMid")
	title := cleanText(stringValue(object, "title", "name", "title_main"))
	if !validIdentifier(mid) || title == "" {
		return Track{}, providerError("decode track", ErrorKindUpstream, "required track fields are missing", 0)
	}

	albumObject, _ := nestedObject(object, "album")
	fileObject, _ := nestedObject(object, "file")
	payObject, _ := nestedObject(object, "pay")
	albumMID := stringValue(albumObject, "mid", "albumMid", "albumMID")
	status := int64Value(object, "status")
	return Track{
		ID:       int64Value(object, "id", "songid", "songId"),
		MID:      mid,
		MediaMID: stringValue(fileObject, "media_mid", "mediaMid"),
		Title:    title,
		Subtitle: cleanText(stringValue(object, "subtitle", "title_extra")),
		Artists:  decodeArtists(object),
		Album: Album{
			ID:       int64Value(albumObject, "id", "albumID"),
			MID:      albumMID,
			Title:    cleanText(stringValue(albumObject, "title", "name", "albumName")),
			CoverURL: albumCoverURL(albumMID),
		},
		Duration:    int(int64Value(object, "interval")),
		PayToPlay:   int64Value(payObject, "pay_play") > 0,
		VIPRequired: int64Value(payObject, "pay_month") > 0,
		Available:   status == 0,
	}, nil
}

func decodeArtists(track map[string]json.RawMessage) []Artist {
	items, err := rawArray(track, "singer", "artists")
	if err != nil {
		return []Artist{}
	}
	artists := make([]Artist, 0, len(items))
	for _, item := range items {
		object, objectErr := decodeObject(item)
		if objectErr != nil {
			continue
		}
		name := cleanText(stringValue(object, "name", "singerName"))
		if name == "" {
			continue
		}
		artists = append(artists, Artist{
			ID:   int64Value(object, "id", "singerID", "singerId"),
			MID:  stringValue(object, "mid", "singerMid", "singerMID"),
			Name: name,
		})
	}
	return artists
}

func rawArray(object map[string]json.RawMessage, keys ...string) ([]json.RawMessage, error) {
	for _, key := range keys {
		raw, exists := object[key]
		if !exists || string(raw) == "null" {
			continue
		}
		var items []json.RawMessage
		if err := decodeJSON(raw, &items); err != nil {
			return nil, err
		}
		return items, nil
	}
	return []json.RawMessage{}, nil
}

func initialSearchCursor(pageToken string) (searchCursor, error) {
	if pageToken == "" {
		return searchCursor{Page: 1}, nil
	}
	if len(pageToken) > maximumPageTokenBytes {
		return searchCursor{}, strconv.ErrSyntax
	}
	decoded, err := base64.RawURLEncoding.DecodeString(pageToken)
	if err != nil || len(decoded) > maximumPageTokenBytes {
		return searchCursor{}, strconv.ErrSyntax
	}
	var cursor searchCursor
	if err := decodeJSON(decoded, &cursor); err != nil {
		return searchCursor{}, err
	}
	if cursor.Page < 1 || cursor.SearchID == "" || len(cursor.SearchID) > 128 {
		return searchCursor{}, strconv.ErrSyntax
	}
	return cursor, nil
}

func encodePageToken(cursor searchCursor) (string, error) {
	encoded, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

func newSearchID() (string, error) {
	randRange := func(maximum int64) (int64, error) {
		value, err := rand.Int(rand.Reader, big.NewInt(maximum))
		if err != nil {
			return 0, err
		}
		return value.Int64(), nil
	}
	first, err := randRange(20)
	if err != nil {
		return "", err
	}
	second, err := randRange(4_194_305)
	if err != nil {
		return "", err
	}
	millisOfDay := timeNowMillis() % 86_400_000
	identifier := uint64(first+1)*18_014_398_509_481_984 + uint64(second)*4_294_967_296 + uint64(millisOfDay)
	return strconv.FormatUint(identifier, 10), nil
}

func timeNowMillis() int64 {
	return time.Now().UnixMilli()
}

func validIdentifier(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}

func albumCoverURL(mid string) string {
	if !validIdentifier(mid) {
		return ""
	}
	return "https://y.gtimg.cn/music/photo_new/T002R500x500M000" + mid + ".jpg"
}

func cleanText(value string) string {
	value = strings.ReplaceAll(value, "<em>", "")
	value = strings.ReplaceAll(value, "</em>", "")
	return strings.TrimSpace(html.UnescapeString(value))
}

func decodeMaybeBase64(value string) string {
	if strings.TrimSpace(value) == "" || strings.ContainsAny(value, "[]<>{}\n\r") {
		return value
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil || !strings.Contains(string(decoded), "[") {
		return value
	}
	return string(decoded)
}
