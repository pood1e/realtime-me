package transport

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	musicv1 "example.com/cloud-drive/api/gen/cloud/music/v1"
	"example.com/cloud-drive/api/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *musicServer) ListProviderConnections(ctx context.Context, _ *connect.Request[musicv1.ListProviderConnectionsRequest]) (*connect.Response[musicv1.ListProviderConnectionsResponse], error) {
	connections, err := s.service.ListProviderConnections(ctx)
	if err != nil {
		return nil, connectError(err)
	}
	result := make([]*musicv1.ProviderConnection, 0, len(connections))
	for _, connection := range connections {
		result = append(result, providerConnectionProto(connection))
	}
	return connect.NewResponse(&musicv1.ListProviderConnectionsResponse{Connections: result}), nil
}

func (s *musicServer) BeginProviderConnection(ctx context.Context, request *connect.Request[musicv1.BeginProviderConnectionRequest]) (*connect.Response[musicv1.BeginProviderConnectionResponse], error) {
	attempt, err := s.service.BeginProviderConnection(ctx, musicProviderDomain(request.Msg.GetProvider()))
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.BeginProviderConnectionResponse{Attempt: providerAttemptProto(attempt)}), nil
}

func (s *musicServer) GetProviderConnectionAttempt(ctx context.Context, request *connect.Request[musicv1.GetProviderConnectionAttemptRequest]) (*connect.Response[musicv1.GetProviderConnectionAttemptResponse], error) {
	attempt, err := s.service.GetProviderConnectionAttempt(ctx, request.Msg.GetAttemptUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.GetProviderConnectionAttemptResponse{Attempt: providerAttemptProto(attempt)}), nil
}

func (s *musicServer) DisconnectProvider(ctx context.Context, request *connect.Request[musicv1.DisconnectProviderRequest]) (*connect.Response[musicv1.DisconnectProviderResponse], error) {
	if err := s.service.DisconnectProvider(ctx, musicProviderDomain(request.Msg.GetProvider())); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.DisconnectProviderResponse{}), nil
}

func (s *musicServer) SearchMusic(ctx context.Context, request *connect.Request[musicv1.SearchMusicRequest]) (*connect.Response[musicv1.SearchMusicResponse], error) {
	cursors := make(map[domain.MusicProvider]string, len(request.Msg.GetCursors()))
	for _, cursor := range request.Msg.GetCursors() {
		provider := musicProviderDomain(cursor.GetProvider())
		if provider == "" {
			return nil, connectError(fmt.Errorf("%w: invalid music provider", domain.ErrInvalidArgument))
		}
		if _, duplicate := cursors[provider]; duplicate {
			return nil, connectError(fmt.Errorf("%w: duplicate music provider cursor", domain.ErrInvalidArgument))
		}
		cursors[provider] = cursor.GetPageToken()
	}
	groups, err := s.service.SearchMusic(ctx, request.Msg.GetQuery(), cursors)
	if err != nil {
		return nil, connectError(err)
	}
	result := make([]*musicv1.ProviderSearchGroup, 0, len(groups))
	for _, group := range groups {
		result = append(result, providerSearchGroupProto(group))
	}
	return connect.NewResponse(&musicv1.SearchMusicResponse{Groups: result}), nil
}

func (s *musicServer) ResolvePlayback(ctx context.Context, request *connect.Request[musicv1.ResolvePlaybackRequest]) (*connect.Response[musicv1.ResolvePlaybackResponse], error) {
	playback, err := s.service.ResolvePlayback(ctx, musicProviderDomain(request.Msg.GetProvider()), request.Msg.GetTrackId(), playbackQualityDomain(request.Msg.GetQuality()))
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.ResolvePlaybackResponse{Playback: playbackDescriptorProto(playback)}), nil
}

func (s *musicServer) GetProviderLyrics(ctx context.Context, request *connect.Request[musicv1.GetProviderLyricsRequest]) (*connect.Response[musicv1.GetProviderLyricsResponse], error) {
	lyric, err := s.service.GetProviderLyrics(ctx, musicProviderDomain(request.Msg.GetProvider()), request.Msg.GetTrackId())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.GetProviderLyricsResponse{Lyric: lyricProto(lyric)}), nil
}

func (s *musicServer) GetSpotifyPlaybackToken(ctx context.Context, _ *connect.Request[musicv1.GetSpotifyPlaybackTokenRequest]) (*connect.Response[musicv1.GetSpotifyPlaybackTokenResponse], error) {
	token, err := s.service.GetSpotifyPlaybackToken(ctx)
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.GetSpotifyPlaybackTokenResponse{
		AccessToken: token.AccessToken, ExpireTime: timestamppb.New(token.ExpireTime),
	}), nil
}
