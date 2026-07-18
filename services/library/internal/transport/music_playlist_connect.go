package transport

import (
	"context"

	"connectrpc.com/connect"
	musicv1 "github.com/pood1e/realtime-me/gen/go/realtime/me/library/music/v1"
	"github.com/pood1e/realtime-me/services/library/internal/app"
	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

type musicPlaylistServer struct{ service *app.MusicPlaylistService }

func (s *musicPlaylistServer) ImportPlaylist(ctx context.Context, request *connect.Request[musicv1.ImportPlaylistRequest]) (*connect.Response[musicv1.ImportPlaylistResponse], error) {
	operation, err := s.service.QueuePlaylistImport(ctx, domain.MusicProvider(request.Msg.GetProviderId()), request.Msg.GetSource())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.ImportPlaylistResponse{PlaylistImport: playlistImportProto(operation)}), nil
}

func (s *musicPlaylistServer) GetPlaylistImport(ctx context.Context, request *connect.Request[musicv1.GetPlaylistImportRequest]) (*connect.Response[musicv1.GetPlaylistImportResponse], error) {
	operation, err := s.service.GetPlaylistImport(ctx, request.Msg.GetPlaylistImportUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.GetPlaylistImportResponse{PlaylistImport: playlistImportProto(operation)}), nil
}

func (s *musicPlaylistServer) GetPlaylist(ctx context.Context, request *connect.Request[musicv1.GetPlaylistRequest]) (*connect.Response[musicv1.GetPlaylistResponse], error) {
	playlist, err := s.service.GetPlaylist(ctx, request.Msg.GetPlaylistUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.GetPlaylistResponse{Playlist: playlistProto(playlist)}), nil
}

func (s *musicPlaylistServer) ListPlaylists(ctx context.Context, request *connect.Request[musicv1.ListPlaylistsRequest]) (*connect.Response[musicv1.ListPlaylistsResponse], error) {
	page, err := s.service.ListPlaylists(ctx, int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	playlists := make([]*musicv1.Playlist, 0, len(page.Playlists))
	for _, playlist := range page.Playlists {
		playlists = append(playlists, playlistProto(playlist))
	}
	return connect.NewResponse(&musicv1.ListPlaylistsResponse{Playlists: playlists, NextPageToken: page.NextPageToken}), nil
}

func (s *musicPlaylistServer) ListPlaylistTracks(ctx context.Context, request *connect.Request[musicv1.ListPlaylistTracksRequest]) (*connect.Response[musicv1.ListPlaylistTracksResponse], error) {
	page, err := s.service.ListPlaylistTracks(ctx, request.Msg.GetPlaylistUid(), int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	tracks := make([]*musicv1.PlaylistTrack, 0, len(page.Tracks))
	for _, track := range page.Tracks {
		tracks = append(tracks, playlistTrackProto(track))
	}
	return connect.NewResponse(&musicv1.ListPlaylistTracksResponse{PlaylistTracks: tracks, NextPageToken: page.NextPageToken}), nil
}

func (s *musicPlaylistServer) DownloadPlaylist(ctx context.Context, request *connect.Request[musicv1.DownloadPlaylistRequest]) (*connect.Response[musicv1.DownloadPlaylistResponse], error) {
	playlist, err := s.service.DownloadPlaylist(ctx, request.Msg.GetPlaylistUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.DownloadPlaylistResponse{Playlist: playlistProto(playlist)}), nil
}

func (s *musicPlaylistServer) DeletePlaylist(ctx context.Context, request *connect.Request[musicv1.DeletePlaylistRequest]) (*connect.Response[musicv1.DeletePlaylistResponse], error) {
	if err := s.service.DeletePlaylist(ctx, request.Msg.GetPlaylistUid()); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.DeletePlaylistResponse{}), nil
}
