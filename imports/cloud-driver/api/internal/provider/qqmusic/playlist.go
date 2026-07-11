package qqmusic

import (
	"context"
	"encoding/json"
	"strings"
)

const (
	playlistPageSize     = 500
	maximumPlaylistSongs = 5000
)

// Playlist is one QQ Music playlist snapshot.
type Playlist struct {
	ID       int64
	Name     string
	CoverURL string
	Tracks   []Track
}

// GetPlaylist retrieves a complete, bounded playlist in provider order.
func (client *Client) GetPlaylist(ctx context.Context, credentials Credentials, playlistID int64) (Playlist, error) {
	operation := "get playlist"
	if playlistID <= 0 {
		return Playlist{}, providerError(operation, ErrorKindInvalidInput, "playlist ID is invalid", 0)
	}
	if err := validateCredentials(operation, credentials, true); err != nil {
		return Playlist{}, err
	}
	playlist := Playlist{ID: playlistID, Tracks: make([]Track, 0, playlistPageSize)}
	seen := make(map[string]struct{})
	total := 0
	for start := 0; start < maximumPlaylistSongs; start += playlistPageSize {
		onlySongs := 0
		if start > 0 {
			onlySongs = 1
		}
		result, err := client.callMusicU(ctx, operation, credentials, nil,
			"music.srfDissInfo.aiDissInfo", "uniform_get_Dissinfo", map[string]any{
				"disstid": playlistID, "enc_host_uin": "", "tag": 1, "userinfo": 1,
				"song_begin": start, "song_num": playlistPageSize, "onlysonglist": onlySongs,
			}, false)
		if err != nil {
			return Playlist{}, err
		}
		root, err := decodeObject(result.Data)
		if err != nil {
			return Playlist{}, providerError(operation, ErrorKindUpstream, "playlist response is invalid", 0)
		}
		if start == 0 {
			directory, ok := nestedObject(root, "dirinfo")
			if !ok {
				return Playlist{}, providerError(operation, ErrorKindUpstream, "playlist metadata is missing", 0)
			}
			playlist.Name = cleanText(stringValue(directory, "title", "dissname", "name"))
			playlist.CoverURL = strings.TrimSpace(stringValue(directory, "picurl", "logo", "cover_url"))
			total = int(int64Value(directory, "songnum", "song_count"))
			if playlist.Name == "" || total < 0 || total > maximumPlaylistSongs {
				return Playlist{}, providerError(operation, ErrorKindUnavailable, "playlist is empty or too large", 0)
			}
		}
		items, err := rawArray(root, "songlist")
		if err != nil {
			return Playlist{}, providerError(operation, ErrorKindUpstream, "playlist tracks are invalid", 0)
		}
		for _, item := range items {
			track, err := decodePlaylistTrack(item)
			if err != nil {
				return Playlist{}, providerError(operation, ErrorKindUpstream, "playlist track is invalid", 0)
			}
			if _, duplicate := seen[track.MID]; duplicate {
				continue
			}
			seen[track.MID] = struct{}{}
			playlist.Tracks = append(playlist.Tracks, track)
		}
		if len(items) < playlistPageSize || (total > 0 && len(playlist.Tracks) >= total) {
			break
		}
	}
	if len(playlist.Tracks) == 0 && total > 0 {
		return Playlist{}, providerError(operation, ErrorKindUpstream, "playlist tracks are missing", 0)
	}
	return playlist, nil
}

func decodePlaylistTrack(raw json.RawMessage) (Track, error) {
	object, err := decodeObject(raw)
	if err != nil {
		return Track{}, err
	}
	if songInfo, exists := object["songInfo"]; exists && len(songInfo) > 0 && string(songInfo) != "null" {
		return decodeTrack(songInfo)
	}
	return decodeTrack(raw)
}
