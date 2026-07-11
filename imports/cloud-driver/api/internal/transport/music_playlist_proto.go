package transport

import (
	musicv1 "example.com/cloud-drive/api/gen/cloud/music/v1"
	"example.com/cloud-drive/api/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func playlistProto(playlist domain.Playlist) *musicv1.Playlist {
	return &musicv1.Playlist{
		Uid: playlist.UID, Provider: musicProviderProto(playlist.Provider), ExternalId: playlist.ExternalID,
		DisplayName: playlist.DisplayName, ArtworkUrl: playlist.ArtworkURL, ProviderUrl: playlist.ProviderURL,
		TrackCount: int32(playlist.TrackCount), DownloadableTrackCount: int32(playlist.DownloadableTrackCount),
		PendingTrackCount:   int32(playlist.PendingTrackCount),
		CompletedTrackCount: int32(playlist.CompletedTrackCount), FailedTrackCount: int32(playlist.FailedTrackCount),
		DownloadSupported: playlist.DownloadSupported, CreateTime: timestamppb.New(playlist.CreateTime),
		UpdateTime: timestamppb.New(playlist.UpdateTime),
	}
}

func playlistTrackProto(item domain.PlaylistTrack) *musicv1.PlaylistTrack {
	return &musicv1.PlaylistTrack{
		Uid: item.UID, PlaylistUid: item.PlaylistUID, Position: int32(item.Position), Track: playableTrackProto(item.Track),
		DownloadStatus: playlistTrackDownloadStatusProto(item.DownloadStatus), LocalTrackUid: item.LocalTrackUID,
	}
}

func playlistTrackDownloadStatusProto(status domain.PlaylistTrackDownloadStatus) musicv1.PlaylistTrackDownloadStatus {
	switch status {
	case domain.PlaylistTrackDownloadNotStarted:
		return musicv1.PlaylistTrackDownloadStatus_PLAYLIST_TRACK_DOWNLOAD_STATUS_NOT_STARTED
	case domain.PlaylistTrackDownloadPending:
		return musicv1.PlaylistTrackDownloadStatus_PLAYLIST_TRACK_DOWNLOAD_STATUS_PENDING
	case domain.PlaylistTrackDownloadRunning:
		return musicv1.PlaylistTrackDownloadStatus_PLAYLIST_TRACK_DOWNLOAD_STATUS_RUNNING
	case domain.PlaylistTrackDownloadCompleted:
		return musicv1.PlaylistTrackDownloadStatus_PLAYLIST_TRACK_DOWNLOAD_STATUS_COMPLETED
	case domain.PlaylistTrackDownloadFailed:
		return musicv1.PlaylistTrackDownloadStatus_PLAYLIST_TRACK_DOWNLOAD_STATUS_FAILED
	default:
		return musicv1.PlaylistTrackDownloadStatus_PLAYLIST_TRACK_DOWNLOAD_STATUS_UNSPECIFIED
	}
}
