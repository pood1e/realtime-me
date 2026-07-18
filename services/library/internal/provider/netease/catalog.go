package netease

import (
	"context"
	"strings"
)

const (
	searchOperation  = "search songs"
	defaultPageLimit = 30
	maximumPageLimit = 100
)

// Search returns one normalized, offset-based page of songs.
func (c *Client) Search(ctx context.Context, request SearchRequest) (SearchPage, error) {
	keywords := strings.TrimSpace(request.Keywords)
	limit := request.Limit
	if limit == 0 {
		limit = defaultPageLimit
	}
	if keywords == "" || request.Offset < 0 || limit < 1 || limit > maximumPageLimit {
		return SearchPage{}, invalidError(searchOperation)
	}

	type artist struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	type album struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		CoverURL string `json:"picUrl"`
	}
	type song struct {
		ID       int64    `json:"id"`
		Name     string   `json:"name"`
		Artists  []artist `json:"ar"`
		Album    album    `json:"al"`
		Duration int64    `json:"dt"`
		Fee      int      `json:"fee"`
	}
	var response struct {
		Code   int `json:"code"`
		Result *struct {
			Songs     []song `json:"songs"`
			SongCount int    `json:"songCount"`
		} `json:"result"`
	}
	payload := map[string]any{
		"s":      keywords,
		"type":   1,
		"limit":  limit,
		"offset": request.Offset,
		"total":  true,
	}
	if err := c.postWEAPI(ctx, searchOperation, "/weapi/cloudsearch/pc", payload, &response); err != nil {
		return SearchPage{}, err
	}
	if err := validateSuccess(searchOperation, response.Code); err != nil {
		return SearchPage{}, err
	}
	if response.Result == nil || response.Result.SongCount < 0 || len(response.Result.Songs) > limit || len(response.Result.Songs) > response.Result.SongCount {
		return SearchPage{}, malformedError(searchOperation)
	}
	if len(response.Result.Songs) == 0 && request.Offset < response.Result.SongCount {
		return SearchPage{}, malformedError(searchOperation)
	}

	page := SearchPage{
		Songs:  make([]SearchSong, 0, len(response.Result.Songs)),
		Total:  response.Result.SongCount,
		Offset: request.Offset,
		Limit:  limit,
	}
	for _, result := range response.Result.Songs {
		if result.ID <= 0 || strings.TrimSpace(result.Name) == "" || result.Duration < 0 || len(result.Artists) == 0 || result.Album.ID <= 0 || strings.TrimSpace(result.Album.Name) == "" {
			return SearchPage{}, malformedError(searchOperation)
		}
		song := SearchSong{
			ID:                   result.ID,
			Name:                 result.Name,
			Artists:              make([]SearchArtist, 0, len(result.Artists)),
			Album:                SearchAlbum{ID: result.Album.ID, Name: result.Album.Name, CoverURL: result.Album.CoverURL},
			DurationMilliseconds: result.Duration,
			Fee:                  result.Fee,
		}
		for _, resultArtist := range result.Artists {
			if resultArtist.ID <= 0 || strings.TrimSpace(resultArtist.Name) == "" {
				return SearchPage{}, malformedError(searchOperation)
			}
			song.Artists = append(song.Artists, SearchArtist{ID: resultArtist.ID, Name: resultArtist.Name})
		}
		page.Songs = append(page.Songs, song)
	}
	page.HasMore = request.Offset+len(page.Songs) < page.Total
	if page.HasMore {
		page.NextOffset = request.Offset + len(page.Songs)
	}
	return page, nil
}
