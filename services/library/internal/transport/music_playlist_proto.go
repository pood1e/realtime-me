package transport

import (
	musicv1 "github.com/pood1e/realtime-me/services/library/gen/cloud/music/v1"
	"github.com/pood1e/realtime-me/services/library/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func playlistProto(playlist domain.Playlist) *musicv1.Playlist {
	return &musicv1.Playlist{
		Uid: playlist.UID, ProviderId: string(playlist.Provider), ExternalId: playlist.ExternalID,
		DisplayName: playlist.DisplayName, ArtworkUrl: playlist.ArtworkURL, ProviderUrl: playlist.ProviderURL,
		TrackCount: int32(playlist.TrackCount), DownloadableTrackCount: int32(playlist.DownloadableTrackCount),
		PendingTrackCount:   int32(playlist.PendingTrackCount),
		CompletedTrackCount: int32(playlist.CompletedTrackCount), FailedTrackCount: int32(playlist.FailedTrackCount),
		DownloadSupported: playlist.DownloadSupported, CreateTime: timestamppb.New(playlist.CreateTime),
		UpdateTime: timestamppb.New(playlist.UpdateTime),
	}
}

func playlistImportProto(operation domain.PlaylistImport) *musicv1.PlaylistImport {
	return &musicv1.PlaylistImport{
		Uid: operation.UID, ProviderId: string(operation.Provider), Status: playlistImportStatusProto(operation.Status),
		PlaylistUid: operation.PlaylistUID, FailureCode: operation.FailureCode,
		CreateTime: timestamppb.New(operation.CreateTime), UpdateTime: timestamppb.New(operation.UpdateTime),
	}
}

func playlistImportStatusProto(status domain.PlaylistImportStatus) musicv1.PlaylistImportStatus {
	switch status {
	case domain.PlaylistImportPending:
		return musicv1.PlaylistImportStatus_PLAYLIST_IMPORT_STATUS_PENDING
	case domain.PlaylistImportRunning:
		return musicv1.PlaylistImportStatus_PLAYLIST_IMPORT_STATUS_RUNNING
	case domain.PlaylistImportCompleted:
		return musicv1.PlaylistImportStatus_PLAYLIST_IMPORT_STATUS_COMPLETED
	case domain.PlaylistImportFailed:
		return musicv1.PlaylistImportStatus_PLAYLIST_IMPORT_STATUS_FAILED
	default:
		return musicv1.PlaylistImportStatus_PLAYLIST_IMPORT_STATUS_UNSPECIFIED
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
