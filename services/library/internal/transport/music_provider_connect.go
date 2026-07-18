package transport

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	musicv1 "github.com/pood1e/realtime-me/services/library/gen/cloud/music/v1"
	"github.com/pood1e/realtime-me/services/library/internal/app"
	"github.com/pood1e/realtime-me/services/library/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type musicProviderServer struct{ service *app.MusicProviderService }

func (s *musicProviderServer) ListProviderConnections(ctx context.Context, _ *connect.Request[musicv1.ListProviderConnectionsRequest]) (*connect.Response[musicv1.ListProviderConnectionsResponse], error) {
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

func (s *musicProviderServer) ListProviders(_ context.Context, _ *connect.Request[musicv1.ListProvidersRequest]) (*connect.Response[musicv1.ListProvidersResponse], error) {
	descriptors := s.service.ListProviders()
	result := make([]*musicv1.ProviderDescriptor, 0, len(descriptors))
	for _, descriptor := range descriptors {
		result = append(result, providerDescriptorProto(descriptor))
	}
	return connect.NewResponse(&musicv1.ListProvidersResponse{Providers: result}), nil
}

func (s *musicProviderServer) BeginProviderConnection(ctx context.Context, request *connect.Request[musicv1.BeginProviderConnectionRequest]) (*connect.Response[musicv1.BeginProviderConnectionResponse], error) {
	attempt, err := s.service.BeginProviderConnection(ctx, domain.MusicProvider(request.Msg.GetProviderId()))
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.BeginProviderConnectionResponse{Attempt: providerAttemptProto(attempt)}), nil
}

func (s *musicProviderServer) GetProviderConnectionAttempt(ctx context.Context, request *connect.Request[musicv1.GetProviderConnectionAttemptRequest]) (*connect.Response[musicv1.GetProviderConnectionAttemptResponse], error) {
	attempt, err := s.service.GetProviderConnectionAttempt(ctx, request.Msg.GetAttemptUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.GetProviderConnectionAttemptResponse{Attempt: providerAttemptProto(attempt)}), nil
}

func (s *musicProviderServer) DisconnectProvider(ctx context.Context, request *connect.Request[musicv1.DisconnectProviderRequest]) (*connect.Response[musicv1.DisconnectProviderResponse], error) {
	if err := s.service.DisconnectProvider(ctx, domain.MusicProvider(request.Msg.GetProviderId())); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.DisconnectProviderResponse{}), nil
}

func (s *musicProviderServer) SearchMusic(ctx context.Context, request *connect.Request[musicv1.SearchMusicRequest]) (*connect.Response[musicv1.SearchMusicResponse], error) {
	cursors := make(map[domain.MusicProvider]string, len(request.Msg.GetCursors()))
	for _, cursor := range request.Msg.GetCursors() {
		provider := domain.MusicProvider(cursor.GetProviderId())
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

func (s *musicProviderServer) ResolvePlayback(ctx context.Context, request *connect.Request[musicv1.ResolvePlaybackRequest]) (*connect.Response[musicv1.ResolvePlaybackResponse], error) {
	playback, err := s.service.ResolvePlayback(ctx, domain.MusicProvider(request.Msg.GetProviderId()), request.Msg.GetTrackId(), playbackQualityDomain(request.Msg.GetQuality()))
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.ResolvePlaybackResponse{Playback: playbackDescriptorProto(playback)}), nil
}

func (s *musicProviderServer) GetProviderLyrics(ctx context.Context, request *connect.Request[musicv1.GetProviderLyricsRequest]) (*connect.Response[musicv1.GetProviderLyricsResponse], error) {
	lyric, err := s.service.GetProviderLyrics(ctx, domain.MusicProvider(request.Msg.GetProviderId()), request.Msg.GetTrackId())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.GetProviderLyricsResponse{Lyric: lyricProto(lyric)}), nil
}

func (s *musicProviderServer) GetProviderPlaybackToken(ctx context.Context, request *connect.Request[musicv1.GetProviderPlaybackTokenRequest]) (*connect.Response[musicv1.GetProviderPlaybackTokenResponse], error) {
	token, err := s.service.GetProviderPlaybackToken(ctx, domain.MusicProvider(request.Msg.GetProviderId()))
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&musicv1.GetProviderPlaybackTokenResponse{
		AccessToken: token.AccessToken, ExpireTime: timestamppb.New(token.ExpireTime),
	}), nil
}
