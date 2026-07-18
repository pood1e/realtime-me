package transport

import (
	"context"

	"connectrpc.com/connect"
	musicv1 "github.com/pood1e/realtime-me/services/library/gen/cloud/music/v1"
	"github.com/pood1e/realtime-me/services/library/internal/app"
)

type musicLibraryServer struct{ service *app.MusicLibraryService }

func (s *musicLibraryServer) GetTrack(ctx context.Context, request *connect.Request[musicv1.GetTrackRequest]) (*connect.Response[musicv1.GetTrackResponse], error) {
	track, err := s.service.Get(ctx, request.Msg.GetTrackUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.GetTrackResponse{Track: trackProto(track)}), nil
}

func (s *musicLibraryServer) ListTracks(ctx context.Context, request *connect.Request[musicv1.ListTracksRequest]) (*connect.Response[musicv1.ListTracksResponse], error) {
	page, err := s.service.List(ctx, request.Msg.GetQuery(), request.Msg.GetAlbum(), request.Msg.GetArtist(), request.Msg.GetFavorites(), request.Msg.GetTrashed(), int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	tracks := make([]*musicv1.Track, 0, len(page.Tracks))
	for _, track := range page.Tracks {
		tracks = append(tracks, trackProto(track))
	}
	return connect.NewResponse(&musicv1.ListTracksResponse{Tracks: tracks, NextPageToken: page.NextPageToken}), nil
}

func (s *musicLibraryServer) ImportTrack(ctx context.Context, request *connect.Request[musicv1.ImportTrackRequest]) (*connect.Response[musicv1.ImportTrackResponse], error) {
	track, err := s.service.Import(ctx, request.Msg.GetUploadUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.ImportTrackResponse{Track: trackProto(track)}), nil
}

func (s *musicLibraryServer) SetTrackFavorite(ctx context.Context, request *connect.Request[musicv1.SetTrackFavoriteRequest]) (*connect.Response[musicv1.SetTrackFavoriteResponse], error) {
	track, err := s.service.SetFavorite(ctx, request.Msg.GetTrackUid(), request.Msg.GetFavorite())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.SetTrackFavoriteResponse{Track: trackProto(track)}), nil
}

func (s *musicLibraryServer) DeleteTrack(ctx context.Context, request *connect.Request[musicv1.DeleteTrackRequest]) (*connect.Response[musicv1.DeleteTrackResponse], error) {
	track, err := s.service.Trash(ctx, request.Msg.GetTrackUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.DeleteTrackResponse{Track: trackProto(track)}), nil
}

func (s *musicLibraryServer) RestoreTrack(ctx context.Context, request *connect.Request[musicv1.RestoreTrackRequest]) (*connect.Response[musicv1.RestoreTrackResponse], error) {
	track, err := s.service.Restore(ctx, request.Msg.GetTrackUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.RestoreTrackResponse{Track: trackProto(track)}), nil
}

func (s *musicLibraryServer) PurgeTrack(ctx context.Context, request *connect.Request[musicv1.PurgeTrackRequest]) (*connect.Response[musicv1.PurgeTrackResponse], error) {
	if err := s.service.Purge(ctx, request.Msg.GetTrackUid()); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.PurgeTrackResponse{}), nil
}

func (s *musicLibraryServer) EmptyTrackTrash(ctx context.Context, _ *connect.Request[musicv1.EmptyTrackTrashRequest]) (*connect.Response[musicv1.EmptyTrackTrashResponse], error) {
	if err := s.service.EmptyTrash(ctx); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.EmptyTrackTrashResponse{}), nil
}

func (s *musicLibraryServer) RetryTrackProcessing(ctx context.Context, request *connect.Request[musicv1.RetryTrackProcessingRequest]) (*connect.Response[musicv1.RetryTrackProcessingResponse], error) {
	track, err := s.service.RetryProcessing(ctx, request.Msg.GetTrackUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.RetryTrackProcessingResponse{Track: trackProto(track)}), nil
}

func (s *musicLibraryServer) RecordPlayback(ctx context.Context, request *connect.Request[musicv1.RecordPlaybackRequest]) (*connect.Response[musicv1.RecordPlaybackResponse], error) {
	entry, err := s.service.RecordPlayback(ctx, playableTrackDomain(request.Msg.GetTrack()))
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.RecordPlaybackResponse{PlaybackEntry: playbackProto(entry)}), nil
}

func (s *musicLibraryServer) ListPlaybackHistory(ctx context.Context, request *connect.Request[musicv1.ListPlaybackHistoryRequest]) (*connect.Response[musicv1.ListPlaybackHistoryResponse], error) {
	page, err := s.service.ListPlaybackHistory(ctx, int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	entries := make([]*musicv1.PlaybackEntry, 0, len(page.Entries))
	for _, entry := range page.Entries {
		entries = append(entries, playbackProto(entry))
	}
	return connect.NewResponse(&musicv1.ListPlaybackHistoryResponse{PlaybackEntries: entries, NextPageToken: page.NextPageToken}), nil
}
