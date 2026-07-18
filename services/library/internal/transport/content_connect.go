package transport

import (
	"context"

	"connectrpc.com/connect"
	contentv1 "github.com/pood1e/realtime-me/gen/go/realtime/me/library/content/v1"
	"github.com/pood1e/realtime-me/services/library/internal/app"
)

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
