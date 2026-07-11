package transport

import (
	"time"

	musicv1 "example.com/cloud-drive/api/gen/cloud/music/v1"
	"example.com/cloud-drive/api/internal/domain"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func musicProviderProto(provider domain.MusicProvider) musicv1.MusicProvider {
	switch provider {
	case domain.MusicProviderLocal:
		return musicv1.MusicProvider_MUSIC_PROVIDER_LOCAL
	case domain.MusicProviderQQ:
		return musicv1.MusicProvider_MUSIC_PROVIDER_QQ_MUSIC
	case domain.MusicProviderNetEase:
		return musicv1.MusicProvider_MUSIC_PROVIDER_NETEASE_CLOUD_MUSIC
	case domain.MusicProviderSpotify:
		return musicv1.MusicProvider_MUSIC_PROVIDER_SPOTIFY
	default:
		return musicv1.MusicProvider_MUSIC_PROVIDER_UNSPECIFIED
	}
}

func musicProviderDomain(provider musicv1.MusicProvider) domain.MusicProvider {
	switch provider {
	case musicv1.MusicProvider_MUSIC_PROVIDER_LOCAL:
		return domain.MusicProviderLocal
	case musicv1.MusicProvider_MUSIC_PROVIDER_QQ_MUSIC:
		return domain.MusicProviderQQ
	case musicv1.MusicProvider_MUSIC_PROVIDER_NETEASE_CLOUD_MUSIC:
		return domain.MusicProviderNetEase
	case musicv1.MusicProvider_MUSIC_PROVIDER_SPOTIFY:
		return domain.MusicProviderSpotify
	default:
		return ""
	}
}

func providerConnectionProto(connection domain.ProviderConnection) *musicv1.ProviderConnection {
	capabilities := make([]musicv1.MusicProviderCapability, 0, len(connection.Capabilities))
	for _, capability := range connection.Capabilities {
		capabilities = append(capabilities, musicProviderCapabilityProto(capability))
	}
	result := &musicv1.ProviderConnection{
		Provider: musicProviderProto(connection.Provider), Status: providerConnectionStatusProto(connection.Status),
		AccountId: connection.AccountID, DisplayName: connection.DisplayName, AvatarUrl: connection.AvatarURL,
		Membership: connection.Membership, Capabilities: capabilities,
	}
	if connection.MembershipExpireTime != nil {
		result.MembershipExpireTime = timestamppb.New(*connection.MembershipExpireTime)
	}
	if !connection.UpdateTime.IsZero() {
		result.UpdateTime = timestamppb.New(connection.UpdateTime)
	}
	return result
}

func musicProviderCapabilityProto(capability domain.MusicProviderCapability) musicv1.MusicProviderCapability {
	switch capability {
	case domain.MusicProviderAccountConnection:
		return musicv1.MusicProviderCapability_MUSIC_PROVIDER_CAPABILITY_ACCOUNT_CONNECTION
	case domain.MusicProviderCatalogSearch:
		return musicv1.MusicProviderCapability_MUSIC_PROVIDER_CAPABILITY_CATALOG_SEARCH
	case domain.MusicProviderPlayback:
		return musicv1.MusicProviderCapability_MUSIC_PROVIDER_CAPABILITY_PLAYBACK
	case domain.MusicProviderLyrics:
		return musicv1.MusicProviderCapability_MUSIC_PROVIDER_CAPABILITY_LYRICS
	case domain.MusicProviderBrowserToken:
		return musicv1.MusicProviderCapability_MUSIC_PROVIDER_CAPABILITY_BROWSER_TOKEN
	default:
		return musicv1.MusicProviderCapability_MUSIC_PROVIDER_CAPABILITY_UNSPECIFIED
	}
}

func providerConnectionStatusProto(status domain.ProviderConnectionStatus) musicv1.ProviderConnectionStatus {
	switch status {
	case domain.ProviderDisconnected:
		return musicv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_DISCONNECTED
	case domain.ProviderConnected:
		return musicv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_CONNECTED
	case domain.ProviderReconnectRequired:
		return musicv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_RECONNECT_REQUIRED
	case domain.ProviderUnavailable:
		return musicv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_UNAVAILABLE
	case domain.ProviderNotConfigured:
		return musicv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_NOT_CONFIGURED
	default:
		return musicv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_UNSPECIFIED
	}
}

func providerAttemptProto(attempt domain.ProviderConnectionAttempt) *musicv1.ProviderConnectionAttempt {
	result := &musicv1.ProviderConnectionAttempt{
		Uid: attempt.UID, Provider: musicProviderProto(attempt.Provider), Status: providerAttemptStatusProto(attempt.Status),
		ExpireTime: timestamppb.New(attempt.ExpireTime),
	}
	if len(attempt.QRImage) > 0 || attempt.QRPayload != "" {
		result.Challenge = &musicv1.ProviderConnectionAttempt_Qr{Qr: &musicv1.ProviderQrChallenge{
			Image: attempt.QRImage, ContentType: attempt.QRContentType, Payload: attempt.QRPayload,
		}}
	} else if attempt.AuthorizationURL != "" {
		result.Challenge = &musicv1.ProviderConnectionAttempt_Redirect{Redirect: &musicv1.ProviderRedirectChallenge{
			AuthorizationUrl: attempt.AuthorizationURL,
		}}
	}
	return result
}

func providerAttemptStatusProto(status domain.ProviderAttemptStatus) musicv1.ProviderConnectionAttemptStatus {
	switch status {
	case domain.ProviderAttemptWaiting:
		return musicv1.ProviderConnectionAttemptStatus_PROVIDER_CONNECTION_ATTEMPT_STATUS_WAITING
	case domain.ProviderAttemptScanned:
		return musicv1.ProviderConnectionAttemptStatus_PROVIDER_CONNECTION_ATTEMPT_STATUS_SCANNED
	case domain.ProviderAttemptConnected:
		return musicv1.ProviderConnectionAttemptStatus_PROVIDER_CONNECTION_ATTEMPT_STATUS_CONNECTED
	case domain.ProviderAttemptExpired:
		return musicv1.ProviderConnectionAttemptStatus_PROVIDER_CONNECTION_ATTEMPT_STATUS_EXPIRED
	case domain.ProviderAttemptRefused:
		return musicv1.ProviderConnectionAttemptStatus_PROVIDER_CONNECTION_ATTEMPT_STATUS_REFUSED
	case domain.ProviderAttemptFailed:
		return musicv1.ProviderConnectionAttemptStatus_PROVIDER_CONNECTION_ATTEMPT_STATUS_FAILED
	default:
		return musicv1.ProviderConnectionAttemptStatus_PROVIDER_CONNECTION_ATTEMPT_STATUS_UNSPECIFIED
	}
}

func providerSearchGroupProto(group domain.ProviderSearchGroup) *musicv1.ProviderSearchGroup {
	tracks := make([]*musicv1.PlayableTrack, 0, len(group.Tracks))
	for _, track := range group.Tracks {
		tracks = append(tracks, playableTrackProto(track))
	}
	return &musicv1.ProviderSearchGroup{
		Provider: musicProviderProto(group.Provider), Status: providerSearchStatusProto(group.Status),
		Tracks: tracks, NextPageToken: group.NextPageToken,
	}
}

func providerSearchStatusProto(status domain.ProviderSearchStatus) musicv1.ProviderSearchStatus {
	switch status {
	case domain.ProviderSearchReady:
		return musicv1.ProviderSearchStatus_PROVIDER_SEARCH_STATUS_READY
	case domain.ProviderSearchNotConnected:
		return musicv1.ProviderSearchStatus_PROVIDER_SEARCH_STATUS_NOT_CONNECTED
	case domain.ProviderSearchUnavailable:
		return musicv1.ProviderSearchStatus_PROVIDER_SEARCH_STATUS_UNAVAILABLE
	case domain.ProviderSearchReconnectRequired:
		return musicv1.ProviderSearchStatus_PROVIDER_SEARCH_STATUS_RECONNECT_REQUIRED
	default:
		return musicv1.ProviderSearchStatus_PROVIDER_SEARCH_STATUS_UNSPECIFIED
	}
}

func playableTrackProto(track domain.PlayableTrack) *musicv1.PlayableTrack {
	return &musicv1.PlayableTrack{
		Provider: musicProviderProto(track.Provider), TrackId: track.TrackID, Title: track.Title, Artists: track.Artists,
		Album: track.Album, Duration: durationpb.New(track.Duration), ArtworkUrl: track.ArtworkURL,
		ProviderUrl: track.ProviderURL, Playable: track.Playable, LyricsAvailable: track.LyricsAvailable,
	}
}

func playableTrackDomain(track *musicv1.PlayableTrack) domain.PlayableTrack {
	if track == nil {
		return domain.PlayableTrack{}
	}
	var duration time.Duration
	if track.Duration != nil && track.Duration.IsValid() {
		duration = track.Duration.AsDuration()
	}
	return domain.PlayableTrack{
		Provider: musicProviderDomain(track.GetProvider()), TrackID: track.GetTrackId(), Title: track.GetTitle(),
		Artists: track.GetArtists(), Album: track.GetAlbum(), Duration: duration, ArtworkURL: track.GetArtworkUrl(),
		ProviderURL: track.GetProviderUrl(), Playable: track.GetPlayable(),
		LyricsAvailable: track.GetLyricsAvailable(),
	}
}

func playbackQualityDomain(quality musicv1.PlaybackQuality) domain.PlaybackQuality {
	switch quality {
	case musicv1.PlaybackQuality_PLAYBACK_QUALITY_HIGH:
		return domain.PlaybackQualityHigh
	case musicv1.PlaybackQuality_PLAYBACK_QUALITY_STANDARD:
		return domain.PlaybackQualityStandard
	default:
		return domain.PlaybackQualityBest
	}
}

func playbackQualityProto(quality domain.PlaybackQuality) musicv1.PlaybackQuality {
	switch quality {
	case domain.PlaybackQualityHigh:
		return musicv1.PlaybackQuality_PLAYBACK_QUALITY_HIGH
	case domain.PlaybackQualityStandard:
		return musicv1.PlaybackQuality_PLAYBACK_QUALITY_STANDARD
	default:
		return musicv1.PlaybackQuality_PLAYBACK_QUALITY_BEST_COMPATIBLE
	}
}

func playbackDescriptorProto(playback domain.PlaybackDescriptor) *musicv1.PlaybackDescriptor {
	result := &musicv1.PlaybackDescriptor{Provider: musicProviderProto(playback.Provider)}
	if !playback.ExpireTime.IsZero() {
		result.ExpireTime = timestamppb.New(playback.ExpireTime)
	}
	if playback.SpotifyURI != "" {
		result.Playback = &musicv1.PlaybackDescriptor_Spotify{Spotify: &musicv1.SpotifyPlayback{Uri: playback.SpotifyURI}}
	} else {
		result.Playback = &musicv1.PlaybackDescriptor_DirectAudio{DirectAudio: &musicv1.DirectAudioPlayback{
			Url: playback.DirectURL, ContentType: playback.ContentType, Quality: playbackQualityProto(playback.Quality),
		}}
	}
	return result
}

func lyricProto(lyric domain.Lyric) *musicv1.Lyric {
	return &musicv1.Lyric{PlainText: lyric.PlainText, SyncedText: lyric.SyncedText, TranslatedText: lyric.TranslatedText}
}
