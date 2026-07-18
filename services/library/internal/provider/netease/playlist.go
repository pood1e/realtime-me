package netease

import (
	"context"
	"encoding/json"
	"strings"
)

const (
	playlistDetailOperation = "get playlist"
	maximumPlaylistTracks   = 5000
	trackDetailBatchSize    = 500
)

type playlistArtist struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type playlistAlbum struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	CoverURL string `json:"picUrl"`
}

type playlistSong struct {
	ID       int64            `json:"id"`
	Name     string           `json:"name"`
	Artists  []playlistArtist `json:"ar"`
	Album    playlistAlbum    `json:"al"`
	Duration int64            `json:"dt"`
	Fee      int              `json:"fee"`
}

// Playlist is one NetEase playlist snapshot.
type Playlist struct {
	ID       int64
	Name     string
	CoverURL string
	Tracks   []SearchSong
}

// GetPlaylist retrieves playlist metadata and ordered track details.
func (c *Client) GetPlaylist(ctx context.Context, playlistID int64) (Playlist, error) {
	if playlistID <= 0 {
		return Playlist{}, invalidError(playlistDetailOperation)
	}
	var response struct {
		Code     int `json:"code"`
		Playlist *struct {
			ID         int64          `json:"id"`
			Name       string         `json:"name"`
			CoverURL   string         `json:"coverImgUrl"`
			TrackCount int            `json:"trackCount"`
			Tracks     []playlistSong `json:"tracks"`
			TrackIDs   []struct {
				ID int64 `json:"id"`
			} `json:"trackIds"`
		} `json:"playlist"`
	}
	payload := map[string]any{"id": playlistID, "n": maximumPlaylistTracks, "s": 8}
	if err := c.postWEAPI(ctx, playlistDetailOperation, "/weapi/v6/playlist/detail", payload, &response); err != nil {
		return Playlist{}, err
	}
	if err := validateSuccess(playlistDetailOperation, response.Code); err != nil {
		return Playlist{}, err
	}
	metadata := response.Playlist
	if metadata == nil || metadata.ID != playlistID || strings.TrimSpace(metadata.Name) == "" ||
		metadata.TrackCount < 0 || metadata.TrackCount > maximumPlaylistTracks || len(metadata.TrackIDs) > maximumPlaylistTracks {
		return Playlist{}, malformedError(playlistDetailOperation)
	}
	orderedIDs := make([]int64, 0, len(metadata.TrackIDs))
	for _, reference := range metadata.TrackIDs {
		if reference.ID <= 0 {
			return Playlist{}, malformedError(playlistDetailOperation)
		}
		orderedIDs = append(orderedIDs, reference.ID)
	}
	if len(orderedIDs) == 0 {
		for _, song := range metadata.Tracks {
			orderedIDs = append(orderedIDs, song.ID)
		}
	}
	byID := make(map[int64]playlistSong, len(orderedIDs))
	for _, song := range metadata.Tracks {
		byID[song.ID] = song
	}
	missing := make([]int64, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		if _, found := byID[id]; !found {
			missing = append(missing, id)
		}
	}
	for start := 0; start < len(missing); start += trackDetailBatchSize {
		end := min(start+trackDetailBatchSize, len(missing))
		songs, err := c.getSongDetails(ctx, missing[start:end])
		if err != nil {
			return Playlist{}, err
		}
		for _, song := range songs {
			byID[song.ID] = song
		}
	}
	playlist := Playlist{ID: metadata.ID, Name: strings.TrimSpace(metadata.Name), CoverURL: metadata.CoverURL}
	playlist.Tracks = make([]SearchSong, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		song, found := byID[id]
		if !found {
			return Playlist{}, malformedError(playlistDetailOperation)
		}
		normalized, err := normalizePlaylistSong(song)
		if err != nil {
			return Playlist{}, err
		}
		playlist.Tracks = append(playlist.Tracks, normalized)
	}
	return playlist, nil
}

func (c *Client) getSongDetails(ctx context.Context, ids []int64) ([]playlistSong, error) {
	references := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		references = append(references, map[string]any{"id": id, "v": 0})
	}
	encodedReferences, err := json.Marshal(references)
	if err != nil {
		return nil, invalidError(playlistDetailOperation)
	}
	encodedIDs, err := json.Marshal(ids)
	if err != nil {
		return nil, invalidError(playlistDetailOperation)
	}
	var response struct {
		Code  int            `json:"code"`
		Songs []playlistSong `json:"songs"`
	}
	if err := c.postWEAPI(ctx, playlistDetailOperation, "/weapi/v3/song/detail", map[string]any{
		"c": string(encodedReferences), "ids": string(encodedIDs),
	}, &response); err != nil {
		return nil, err
	}
	if err := validateSuccess(playlistDetailOperation, response.Code); err != nil {
		return nil, err
	}
	return response.Songs, nil
}

func normalizePlaylistSong(song playlistSong) (SearchSong, error) {
	if song.ID <= 0 || strings.TrimSpace(song.Name) == "" || song.Duration < 0 || len(song.Artists) == 0 ||
		song.Album.ID <= 0 || strings.TrimSpace(song.Album.Name) == "" {
		return SearchSong{}, malformedError(playlistDetailOperation)
	}
	result := SearchSong{
		ID: song.ID, Name: strings.TrimSpace(song.Name), Album: SearchAlbum{
			ID: song.Album.ID, Name: strings.TrimSpace(song.Album.Name), CoverURL: song.Album.CoverURL,
		}, DurationMilliseconds: song.Duration, Fee: song.Fee,
	}
	for _, artist := range song.Artists {
		if artist.ID <= 0 || strings.TrimSpace(artist.Name) == "" {
			return SearchSong{}, malformedError(playlistDetailOperation)
		}
		result.Artists = append(result.Artists, SearchArtist{ID: artist.ID, Name: strings.TrimSpace(artist.Name)})
	}
	return result, nil
}
