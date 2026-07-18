package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
	"github.com/pood1e/realtime-me/services/library/internal/provider/netease"
)

func searchNetEase(ctx context.Context, rawCredentials []byte, query string, pageSize int, pageToken string) ([]domain.PlayableTrack, string, []byte, error) {
	client, _, err := netEaseClient(rawCredentials)
	if err != nil {
		return nil, "", nil, err
	}
	offset, err := parseOffsetToken(pageToken)
	if err != nil {
		return nil, "", nil, err
	}
	page, err := client.Search(ctx, netease.SearchRequest{Keywords: query, Offset: offset, Limit: pageSize})
	if err != nil {
		return nil, "", nil, mapProviderError(err)
	}
	tracks := make([]domain.PlayableTrack, 0, len(page.Songs))
	for _, song := range page.Songs {
		tracks = append(tracks, netEasePlayableTrack(song))
	}
	nextPageToken := ""
	if page.HasMore {
		nextPageToken = strconv.Itoa(page.NextOffset)
	}
	updated, err := json.Marshal(client.Credentials())
	if err != nil {
		return nil, "", nil, fmt.Errorf("%w: encode NetEase credentials", domain.ErrUnavailable)
	}
	return tracks, nextPageToken, updated, nil
}

func netEasePlayableTrack(song netease.SearchSong) domain.PlayableTrack {
	artists := make([]string, 0, len(song.Artists))
	for _, artist := range song.Artists {
		artists = append(artists, artist.Name)
	}
	trackID := strconv.FormatInt(song.ID, 10)
	return domain.PlayableTrack{
		Provider: domain.MusicProviderNetEase, TrackID: trackID, Title: song.Name, Artists: artists,
		Album: song.Album.Name, Duration: time.Duration(song.DurationMilliseconds) * time.Millisecond,
		ArtworkURL: song.Album.CoverURL, ProviderURL: "https://music.163.com/song?id=" + trackID,
		Playable: true, LyricsAvailable: true,
	}
}

func parseOffsetToken(pageToken string) (int, error) {
	if pageToken == "" {
		return 0, nil
	}
	offset, err := strconv.Atoi(pageToken)
	if err != nil || offset < 0 || offset > 100_000 {
		return 0, fmt.Errorf("%w: invalid provider page token", domain.ErrInvalidArgument)
	}
	return offset, nil
}
