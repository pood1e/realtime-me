package transport

import (
	"context"

	"connectrpc.com/connect"
	musicv1 "example.com/cloud-drive/api/gen/cloud/music/v1"
	"example.com/cloud-drive/api/internal/app"
)

type musicServer struct{ service *app.MusicService }

func (s *musicServer) GetTrack(ctx context.Context, request *connect.Request[musicv1.GetTrackRequest]) (*connect.Response[musicv1.GetTrackResponse], error) {
	track, err := s.service.Get(ctx, request.Msg.GetTrackUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.GetTrackResponse{Track: trackProto(track)}), nil
}

func (s *musicServer) ListTracks(ctx context.Context, request *connect.Request[musicv1.ListTracksRequest]) (*connect.Response[musicv1.ListTracksResponse], error) {
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

func (s *musicServer) ImportTrack(ctx context.Context, request *connect.Request[musicv1.ImportTrackRequest]) (*connect.Response[musicv1.ImportTrackResponse], error) {
	track, err := s.service.Import(ctx, request.Msg.GetUploadUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.ImportTrackResponse{Track: trackProto(track)}), nil
}

func (s *musicServer) SetTrackFavorite(ctx context.Context, request *connect.Request[musicv1.SetTrackFavoriteRequest]) (*connect.Response[musicv1.SetTrackFavoriteResponse], error) {
	track, err := s.service.SetFavorite(ctx, request.Msg.GetTrackUid(), request.Msg.GetFavorite())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.SetTrackFavoriteResponse{Track: trackProto(track)}), nil
}

func (s *musicServer) DeleteTrack(ctx context.Context, request *connect.Request[musicv1.DeleteTrackRequest]) (*connect.Response[musicv1.DeleteTrackResponse], error) {
	track, err := s.service.Trash(ctx, request.Msg.GetTrackUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.DeleteTrackResponse{Track: trackProto(track)}), nil
}

func (s *musicServer) RestoreTrack(ctx context.Context, request *connect.Request[musicv1.RestoreTrackRequest]) (*connect.Response[musicv1.RestoreTrackResponse], error) {
	track, err := s.service.Restore(ctx, request.Msg.GetTrackUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.RestoreTrackResponse{Track: trackProto(track)}), nil
}

func (s *musicServer) PurgeTrack(ctx context.Context, request *connect.Request[musicv1.PurgeTrackRequest]) (*connect.Response[musicv1.PurgeTrackResponse], error) {
	if err := s.service.Purge(ctx, request.Msg.GetTrackUid()); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.PurgeTrackResponse{}), nil
}

func (s *musicServer) EmptyTrackTrash(ctx context.Context, _ *connect.Request[musicv1.EmptyTrackTrashRequest]) (*connect.Response[musicv1.EmptyTrackTrashResponse], error) {
	if err := s.service.EmptyTrash(ctx); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.EmptyTrackTrashResponse{}), nil
}

func (s *musicServer) RetryTrackProcessing(ctx context.Context, request *connect.Request[musicv1.RetryTrackProcessingRequest]) (*connect.Response[musicv1.RetryTrackProcessingResponse], error) {
	track, err := s.service.RetryProcessing(ctx, request.Msg.GetTrackUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.RetryTrackProcessingResponse{Track: trackProto(track)}), nil
}

func (s *musicServer) ListAlbums(ctx context.Context, request *connect.Request[musicv1.ListAlbumsRequest]) (*connect.Response[musicv1.ListAlbumsResponse], error) {
	albums, err := s.service.ListAlbums(ctx, request.Msg.GetQuery())
	if err != nil {
		return nil, connectError(err)
	}
	result := make([]*musicv1.Album, 0, len(albums))
	for _, album := range albums {
		result = append(result, albumProto(album))
	}
	return connect.NewResponse(&musicv1.ListAlbumsResponse{Albums: result}), nil
}

func (s *musicServer) ListArtists(ctx context.Context, request *connect.Request[musicv1.ListArtistsRequest]) (*connect.Response[musicv1.ListArtistsResponse], error) {
	artists, err := s.service.ListArtists(ctx, request.Msg.GetQuery())
	if err != nil {
		return nil, connectError(err)
	}
	result := make([]*musicv1.Artist, 0, len(artists))
	for _, artist := range artists {
		result = append(result, artistProto(artist))
	}
	return connect.NewResponse(&musicv1.ListArtistsResponse{Artists: result}), nil
}

func (s *musicServer) RecordPlayback(ctx context.Context, request *connect.Request[musicv1.RecordPlaybackRequest]) (*connect.Response[musicv1.RecordPlaybackResponse], error) {
	entry, err := s.service.RecordPlayback(ctx, request.Msg.GetTrackUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.RecordPlaybackResponse{PlaybackEntry: playbackProto(entry)}), nil
}

func (s *musicServer) ListPlaybackHistory(ctx context.Context, request *connect.Request[musicv1.ListPlaybackHistoryRequest]) (*connect.Response[musicv1.ListPlaybackHistoryResponse], error) {
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
