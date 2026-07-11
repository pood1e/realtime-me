package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"example.com/cloud-drive/api/internal/domain"
)

func (s *MusicService) ResolvePlayback(ctx context.Context, provider domain.MusicProvider, trackID string, quality domain.PlaybackQuality) (domain.PlaybackDescriptor, error) {
	trackID = strings.TrimSpace(trackID)
	if trackID == "" {
		return domain.PlaybackDescriptor{}, fmt.Errorf("%w: track ID is required", domain.ErrInvalidArgument)
	}
	if provider == domain.MusicProviderLocal {
		track, err := s.store.GetTrack(ctx, trackID, false)
		if err != nil {
			return domain.PlaybackDescriptor{}, err
		}
		return localPlaybackDescriptor(track, quality), nil
	}
	localTrack, err := s.store.GetTrackBySource(ctx, provider, trackID)
	if err == nil {
		return localPlaybackDescriptor(localTrack, quality), nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.PlaybackDescriptor{}, err
	}
	connection, credentials, err := s.providerCredentials(ctx, provider)
	if err != nil {
		return domain.PlaybackDescriptor{}, err
	}
	if quality == "" {
		quality = domain.PlaybackQualityBest
	}
	adapter, err := s.providerAdapter(provider)
	if err != nil {
		return domain.PlaybackDescriptor{}, err
	}
	resolver, supported := adapter.(domain.MusicPlaybackResolver)
	if !supported {
		return domain.PlaybackDescriptor{}, fmt.Errorf("%w: provider does not support playback", domain.ErrConflict)
	}
	playback, updated, err := resolver.ResolvePlayback(ctx, credentials, trackID, quality)
	if err != nil {
		s.markProviderFailure(ctx, connection, err)
		return domain.PlaybackDescriptor{}, err
	}
	if err := s.persistCredentialUpdate(ctx, connection, credentials, updated); err != nil {
		return domain.PlaybackDescriptor{}, err
	}
	return playback, nil
}

func localPlaybackDescriptor(track domain.Track, quality domain.PlaybackQuality) domain.PlaybackDescriptor {
	return domain.PlaybackDescriptor{
		Provider: domain.MusicProviderLocal, DirectURL: "/v1/tracks/" + track.UID + "/content",
		ContentType: track.ContentType, Quality: quality,
	}
}

// GetProviderLyrics returns lyrics from the selected source only.
func (s *MusicService) GetProviderLyrics(ctx context.Context, provider domain.MusicProvider, trackID string) (domain.Lyric, error) {
	if provider == domain.MusicProviderLocal || strings.TrimSpace(trackID) == "" {
		return domain.Lyric{}, fmt.Errorf("%w: lyrics are unavailable for this source", domain.ErrNotFound)
	}
	connection, credentials, err := s.providerCredentials(ctx, provider)
	if err != nil {
		return domain.Lyric{}, err
	}
	adapter, err := s.providerAdapter(provider)
	if err != nil {
		return domain.Lyric{}, err
	}
	lyricsProvider, supported := adapter.(domain.MusicLyricsProvider)
	if !supported {
		return domain.Lyric{}, fmt.Errorf("%w: provider does not expose lyrics", domain.ErrNotFound)
	}
	lyric, updated, err := lyricsProvider.Lyrics(ctx, credentials, strings.TrimSpace(trackID))
	if err != nil {
		s.markProviderFailure(ctx, connection, err)
		return domain.Lyric{}, err
	}
	if err := s.persistCredentialUpdate(ctx, connection, credentials, updated); err != nil {
		return domain.Lyric{}, err
	}
	return lyric, nil
}

// GetSpotifyPlaybackToken returns a short-lived official SDK credential.
func (s *MusicService) GetSpotifyPlaybackToken(ctx context.Context) (domain.ProviderPlaybackToken, error) {
	connection, credentials, err := s.providerCredentials(ctx, domain.MusicProviderSpotify)
	if err != nil {
		return domain.ProviderPlaybackToken{}, err
	}
	adapter, err := s.providerAdapter(domain.MusicProviderSpotify)
	if err != nil {
		return domain.ProviderPlaybackToken{}, err
	}
	tokenProvider, supported := adapter.(domain.MusicPlaybackTokenProvider)
	if !supported {
		return domain.ProviderPlaybackToken{}, fmt.Errorf("%w: provider does not expose browser playback tokens", domain.ErrConflict)
	}
	token, updated, err := tokenProvider.PlaybackToken(ctx, credentials)
	if err != nil {
		s.markProviderFailure(ctx, connection, err)
		return domain.ProviderPlaybackToken{}, err
	}
	if err := s.persistCredentialUpdate(ctx, connection, credentials, updated); err != nil {
		return domain.ProviderPlaybackToken{}, err
	}
	return token, nil
}
