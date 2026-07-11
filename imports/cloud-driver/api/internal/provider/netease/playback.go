package netease

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	songURLOperation         = "get song URL"
	lyricsOperation          = "get lyrics"
	highestCompatibleQuality = "hires"
	fallback320KQuality      = "exhigh"
)

type lyricBlockResponse struct {
	Lyric string `json:"lyric"`
}

// SongURL returns an authenticated, short-lived playback URL.
func (c *Client) SongURL(ctx context.Context, request SongURLRequest) (SongURL, error) {
	if request.SongID <= 0 {
		return SongURL{}, invalidError(songURLOperation)
	}
	quality := string(request.Quality)
	if quality == "" {
		quality = highestCompatibleQuality
	}
	if quality != highestCompatibleQuality && quality != fallback320KQuality && quality != string(AudioQualityStandard) {
		return SongURL{}, invalidError(songURLOperation)
	}
	resource, available, err := c.requestSongURL(ctx, request.SongID, quality)
	if err != nil {
		return SongURL{}, err
	}
	if available {
		return resource, nil
	}
	if request.FallbackTo320K && quality != fallback320KQuality {
		resource, available, err = c.requestSongURL(ctx, request.SongID, fallback320KQuality)
		if err != nil {
			return SongURL{}, err
		}
		if available {
			return resource, nil
		}
	}
	return SongURL{}, &ProviderError{Operation: songURLOperation, Kind: ErrorKindUnavailable}
}

func (c *Client) requestSongURL(ctx context.Context, songID int64, quality string) (SongURL, bool, error) {
	type freeTrial struct {
		Start int64 `json:"start"`
		End   int64 `json:"end"`
	}
	type songURL struct {
		ID        int64      `json:"id"`
		URL       *string    `json:"url"`
		Bitrate   int        `json:"br"`
		Size      int64      `json:"size"`
		MD5       string     `json:"md5"`
		Code      int        `json:"code"`
		ExpiresIn int        `json:"expi"`
		Format    string     `json:"type"`
		Level     string     `json:"level"`
		Fee       int        `json:"fee"`
		FreeTrial *freeTrial `json:"freeTrialInfo"`
	}
	var response struct {
		Code int       `json:"code"`
		Data []songURL `json:"data"`
	}
	payload := map[string]any{
		"ids":        "[" + strconv.FormatInt(songID, 10) + "]",
		"level":      quality,
		"encodeType": "flac",
	}
	if err := c.postWEAPI(ctx, songURLOperation, "/weapi/song/enhance/player/url/v1", payload, &response); err != nil {
		return SongURL{}, false, err
	}
	if err := validateSuccess(songURLOperation, response.Code); err != nil {
		return SongURL{}, false, err
	}
	if len(response.Data) != 1 || response.Data[0].ID != songID {
		return SongURL{}, false, malformedError(songURLOperation)
	}
	result := response.Data[0]
	if result.Code != http.StatusOK || result.URL == nil || strings.TrimSpace(*result.URL) == "" {
		return SongURL{}, false, nil
	}
	parsedURL, err := url.Parse(*result.URL)
	if err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || result.Bitrate <= 0 || result.Size <= 0 || result.ExpiresIn < 0 || strings.TrimSpace(result.Level) == "" || strings.TrimSpace(result.Format) == "" {
		return SongURL{}, false, malformedError(songURLOperation)
	}
	resource := SongURL{
		SongID:           songID,
		URL:              parsedURL.String(),
		Quality:          result.Level,
		Format:           result.Format,
		Bitrate:          result.Bitrate,
		SizeBytes:        result.Size,
		MD5:              result.MD5,
		ExpiresInSeconds: result.ExpiresIn,
		Fee:              result.Fee,
	}
	if result.FreeTrial != nil {
		if result.FreeTrial.Start < 0 || result.FreeTrial.End < result.FreeTrial.Start {
			return SongURL{}, false, malformedError(songURLOperation)
		}
		resource.FreeTrial = &FreeTrial{StartMilliseconds: result.FreeTrial.Start, EndMilliseconds: result.FreeTrial.End}
	}
	return resource, true, nil
}

// Lyrics returns original, translated, and romanized lyrics when available.
func (c *Client) Lyrics(ctx context.Context, songID int64) (Lyrics, error) {
	if songID <= 0 {
		return Lyrics{}, invalidError(lyricsOperation)
	}
	var response struct {
		Code        int                 `json:"code"`
		Original    *lyricBlockResponse `json:"lrc"`
		Translation *lyricBlockResponse `json:"tlyric"`
		Romanized   *lyricBlockResponse `json:"romalrc"`
		NoLyrics    bool                `json:"nolyric"`
		Uncollected bool                `json:"uncollected"`
	}
	payload := map[string]any{
		"id":      songID,
		"tv":      -1,
		"lv":      -1,
		"rv":      -1,
		"kv":      -1,
		"_nmclfl": 1,
	}
	if err := c.postWEAPI(ctx, lyricsOperation, "/weapi/song/lyric", payload, &response); err != nil {
		return Lyrics{}, err
	}
	if err := validateSuccess(lyricsOperation, response.Code); err != nil {
		return Lyrics{}, err
	}
	if !response.NoLyrics && !response.Uncollected && response.Original == nil {
		return Lyrics{}, malformedError(lyricsOperation)
	}
	if !response.NoLyrics && !response.Uncollected && strings.TrimSpace(response.Original.Lyric) == "" {
		return Lyrics{}, malformedError(lyricsOperation)
	}
	return Lyrics{
		SongID:       songID,
		Original:     lyricText(response.Original),
		Translation:  lyricText(response.Translation),
		Romanization: lyricText(response.Romanized),
		Instrumental: response.NoLyrics,
		NotCollected: response.Uncollected,
	}, nil
}

func lyricText(block *lyricBlockResponse) string {
	if block == nil {
		return ""
	}
	return block.Lyric
}
