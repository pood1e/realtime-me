package transport

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	authv1 "github.com/pood1e/realtime-me/gen/go/realtime/me/library/auth/v1"
	contentv1 "github.com/pood1e/realtime-me/gen/go/realtime/me/library/content/v1"
	"github.com/pood1e/realtime-me/services/library/internal/app"
	"github.com/pood1e/realtime-me/services/library/internal/auth"
	"github.com/pood1e/realtime-me/services/library/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type returnURLValidator interface {
	ReturnURL(string) (string, error)
}

type authServer struct {
	sessions  *auth.Manager
	validator returnURLValidator
}

func (s *authServer) Login(_ context.Context, request *connect.Request[authv1.LoginRequest]) (*connect.Response[authv1.LoginResponse], error) {
	returnURL, err := s.validator.ReturnURL(request.Msg.GetReturnUrl())
	if err != nil {
		return nil, connectError(fmt.Errorf("%w: invalid return URL", domain.ErrInvalidArgument))
	}
	clientAddress := ""
	if values := request.Header().Values("CF-Connecting-IP"); len(values) == 1 {
		clientAddress = values[0]
	}
	cookie, err := s.sessions.Login(request.Msg.GetPassword(), clientAddress)
	request.Msg.Password = ""
	if err != nil {
		return nil, sessionError(err)
	}
	response := connect.NewResponse(&authv1.LoginResponse{ReturnUrl: returnURL})
	response.Header().Add("Set-Cookie", cookie.String())
	return response, nil
}

func (s *authServer) Logout(ctx context.Context, _ *connect.Request[authv1.LogoutRequest]) (*connect.Response[authv1.LogoutResponse], error) {
	if _, authenticated := auth.SessionFromContext(ctx); !authenticated {
		return nil, sessionError(auth.ErrUnauthenticated)
	}
	response := connect.NewResponse(&authv1.LogoutResponse{})
	response.Header().Add("Set-Cookie", s.sessions.LogoutCookie().String())
	return response, nil
}

func (s *authServer) GetSession(ctx context.Context, _ *connect.Request[authv1.GetSessionRequest]) (*connect.Response[authv1.GetSessionResponse], error) {
	session, authenticated := auth.SessionFromContext(ctx)
	if !authenticated {
		return nil, sessionError(auth.ErrUnauthenticated)
	}
	return connect.NewResponse(&authv1.GetSessionResponse{ExpireTime: timestamppb.New(session.ExpireTime)}), nil
}

type contentServer struct{ service *app.ContentService }

func (s *contentServer) StartUpload(ctx context.Context, request *connect.Request[contentv1.StartUploadRequest]) (*connect.Response[contentv1.StartUploadResponse], error) {
	upload, err := s.service.StartUpload(ctx, request.Msg.GetFileName(), request.Msg.GetContentType(), request.Msg.GetTotalSizeBytes())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&contentv1.StartUploadResponse{Upload: uploadProto(upload), ChunkUrl: "/v1/uploads/" + upload.UID + "/chunks"}), nil
}

func (s *contentServer) GetUpload(ctx context.Context, request *connect.Request[contentv1.GetUploadRequest]) (*connect.Response[contentv1.GetUploadResponse], error) {
	upload, err := s.service.GetUpload(ctx, request.Msg.GetUploadUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&contentv1.GetUploadResponse{Upload: uploadProto(upload)}), nil
}

func (s *contentServer) FinalizeUpload(ctx context.Context, request *connect.Request[contentv1.FinalizeUploadRequest]) (*connect.Response[contentv1.FinalizeUploadResponse], error) {
	upload, err := s.service.FinalizeUpload(ctx, request.Msg.GetUploadUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&contentv1.FinalizeUploadResponse{Upload: uploadProto(upload)}), nil
}

func (s *contentServer) AbandonUpload(ctx context.Context, request *connect.Request[contentv1.AbandonUploadRequest]) (*connect.Response[contentv1.AbandonUploadResponse], error) {
	if err := s.service.AbandonUpload(ctx, request.Msg.GetUploadUid()); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&contentv1.AbandonUploadResponse{}), nil
}
