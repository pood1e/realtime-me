package app

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"example.com/cloud-drive/api/internal/domain"
)

func (s *MusicService) validatePlaybackTrack(ctx context.Context, track domain.PlayableTrack) (domain.PlayableTrack, error) {
	if track.Provider == domain.MusicProviderLocal {
		local, err := s.store.GetTrack(ctx, strings.TrimSpace(track.TrackID), false)
		if err != nil {
			return domain.PlayableTrack{}, err
		}
		return playableTrackFromLocal(local), nil
	}
	track.TrackID = strings.TrimSpace(track.TrackID)
	track.Title = strings.TrimSpace(track.Title)
	if _, err := s.providerAdapter(track.Provider); err != nil || track.TrackID == "" || track.Title == "" || len(track.TrackID) > 512 || len(track.Title) > 512 {
		return domain.PlayableTrack{}, fmt.Errorf("%w: invalid playback track", domain.ErrInvalidArgument)
	}
	track.Artists = normalizeTrackArtists(track.Artists)
	track.Album = truncateText(strings.TrimSpace(track.Album), 512)
	track.ArtworkURL = validatedProviderURL(track.ArtworkURL)
	track.ProviderURL = validatedProviderURL(track.ProviderURL)
	return track, nil
}

func playableTrackFromLocal(track domain.Track) domain.PlayableTrack {
	artworkURL := ""
	if track.ArtworkStorageKey != "" {
		artworkURL = "/v1/tracks/" + track.UID + "/artwork"
	}
	return domain.PlayableTrack{
		Provider: domain.MusicProviderLocal, TrackID: track.UID, Title: track.Title, Artists: track.Artists,
		Album: track.Album, Duration: track.Duration, ArtworkURL: artworkURL,
		Playable: track.ProcessingStatus == domain.ProcessingStatusReady,
	}
}

func normalizeTrackArtists(artists []string) []string {
	result := make([]string, 0, min(len(artists), 32))
	for _, artist := range artists {
		artist = truncateText(strings.TrimSpace(artist), 256)
		if artist != "" {
			result = append(result, artist)
		}
		if len(result) == 32 {
			break
		}
	}
	return result
}

func validatedProviderURL(value string) string {
	value = strings.TrimSpace(value)
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
		return ""
	}
	return value
}

func truncateText(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
