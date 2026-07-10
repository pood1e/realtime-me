package transport

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	cloud_drivev1 "example.com/cloud-drive/api/gen/cloud/drive/v1"
	"example.com/cloud-drive/api/internal/app"
	"example.com/cloud-drive/api/internal/auth"
	"example.com/cloud-drive/api/internal/domain"
)

// ConnectServer adapts generated ConnectRPC contracts to the application service.
type ConnectServer struct {
	service  *app.Service
	sessions *auth.Manager
}

// NewConnectServer constructs generated-service implementations.
func NewConnectServer(service *app.Service, sessions *auth.Manager) *ConnectServer {
	return &ConnectServer{service: service, sessions: sessions}
}

// GetDriveItem implements DriveService.GetDriveItem.
func (s *ConnectServer) GetDriveItem(ctx context.Context, request *connect.Request[cloud_drivev1.GetDriveItemRequest]) (*connect.Response[cloud_drivev1.GetDriveItemResponse], error) {
	item, err := s.service.GetItem(ctx, request.Msg.GetItemUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&cloud_drivev1.GetDriveItemResponse{Item: itemProto(item)}), nil
}

// ListDriveItems implements DriveService.ListDriveItems.
func (s *ConnectServer) ListDriveItems(ctx context.Context, request *connect.Request[cloud_drivev1.ListDriveItemsRequest]) (*connect.Response[cloud_drivev1.ListDriveItemsResponse], error) {
	page, err := s.service.ListItems(ctx, stringPointer(request.Msg.GetParentUid()), request.Msg.GetIncludeTrashed(), int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	items, nextToken := pageProto(page)
	return connect.NewResponse(&cloud_drivev1.ListDriveItemsResponse{Items: items, NextPageToken: nextToken}), nil
}

// ListTrashedItems implements DriveService.ListTrashedItems.
func (s *ConnectServer) ListTrashedItems(ctx context.Context, request *connect.Request[cloud_drivev1.ListTrashedItemsRequest]) (*connect.Response[cloud_drivev1.ListTrashedItemsResponse], error) {
	page, err := s.service.ListTrashedItems(ctx, int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	items, nextToken := pageProto(page)
	return connect.NewResponse(&cloud_drivev1.ListTrashedItemsResponse{Items: items, NextPageToken: nextToken}), nil
}

// SearchDriveItems implements DriveService.SearchDriveItems.
func (s *ConnectServer) SearchDriveItems(ctx context.Context, request *connect.Request[cloud_drivev1.SearchDriveItemsRequest]) (*connect.Response[cloud_drivev1.SearchDriveItemsResponse], error) {
	page, err := s.service.SearchItems(ctx, request.Msg.GetQuery(), int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	items, nextToken := pageProto(page)
	return connect.NewResponse(&cloud_drivev1.SearchDriveItemsResponse{Items: items, NextPageToken: nextToken}), nil
}

// CreateDirectory implements DriveService.CreateDirectory.
func (s *ConnectServer) CreateDirectory(ctx context.Context, request *connect.Request[cloud_drivev1.CreateDirectoryRequest]) (*connect.Response[cloud_drivev1.CreateDirectoryResponse], error) {
	item, err := s.service.CreateDirectory(ctx, stringPointer(request.Msg.GetParentUid()), request.Msg.GetName())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&cloud_drivev1.CreateDirectoryResponse{Item: itemProto(item)}), nil
}

// RenameDriveItem implements DriveService.RenameDriveItem.
func (s *ConnectServer) RenameDriveItem(ctx context.Context, request *connect.Request[cloud_drivev1.RenameDriveItemRequest]) (*connect.Response[cloud_drivev1.RenameDriveItemResponse], error) {
	item, err := s.service.RenameItem(ctx, request.Msg.GetItemUid(), request.Msg.GetName())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&cloud_drivev1.RenameDriveItemResponse{Item: itemProto(item)}), nil
}

// MoveDriveItem implements DriveService.MoveDriveItem.
func (s *ConnectServer) MoveDriveItem(ctx context.Context, request *connect.Request[cloud_drivev1.MoveDriveItemRequest]) (*connect.Response[cloud_drivev1.MoveDriveItemResponse], error) {
	item, err := s.service.MoveItem(ctx, request.Msg.GetItemUid(), stringPointer(request.Msg.GetParentUid()))
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&cloud_drivev1.MoveDriveItemResponse{Item: itemProto(item)}), nil
}

// DeleteDriveItem implements DriveService.DeleteDriveItem.
func (s *ConnectServer) DeleteDriveItem(ctx context.Context, request *connect.Request[cloud_drivev1.DeleteDriveItemRequest]) (*connect.Response[cloud_drivev1.DeleteDriveItemResponse], error) {
	item, err := s.service.TrashItem(ctx, request.Msg.GetItemUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&cloud_drivev1.DeleteDriveItemResponse{Item: itemProto(item)}), nil
}

// RestoreDriveItem implements DriveService.RestoreDriveItem.
func (s *ConnectServer) RestoreDriveItem(ctx context.Context, request *connect.Request[cloud_drivev1.RestoreDriveItemRequest]) (*connect.Response[cloud_drivev1.RestoreDriveItemResponse], error) {
	item, err := s.service.RestoreItem(ctx, request.Msg.GetItemUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&cloud_drivev1.RestoreDriveItemResponse{Item: itemProto(item)}), nil
}

// PurgeDriveItem implements DriveService.PurgeDriveItem.
func (s *ConnectServer) PurgeDriveItem(ctx context.Context, request *connect.Request[cloud_drivev1.PurgeDriveItemRequest]) (*connect.Response[cloud_drivev1.PurgeDriveItemResponse], error) {
	if err := s.service.PurgeItem(ctx, request.Msg.GetItemUid()); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&cloud_drivev1.PurgeDriveItemResponse{}), nil
}

// EmptyTrash implements DriveService.EmptyTrash.
func (s *ConnectServer) EmptyTrash(ctx context.Context, _ *connect.Request[cloud_drivev1.EmptyTrashRequest]) (*connect.Response[cloud_drivev1.EmptyTrashResponse], error) {
	if err := s.service.EmptyTrash(ctx); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&cloud_drivev1.EmptyTrashResponse{}), nil
}

// GetDownload implements DriveService.GetDownload.
func (s *ConnectServer) GetDownload(ctx context.Context, request *connect.Request[cloud_drivev1.GetDownloadRequest]) (*connect.Response[cloud_drivev1.GetDownloadResponse], error) {
	item, err := s.service.GetItem(ctx, request.Msg.GetItemUid())
	if err != nil {
		return nil, connectError(err)
	}
	if item.Kind != domain.ItemKindFile {
		return nil, connectError(fmt.Errorf("%w: item is not a file", domain.ErrConflict))
	}
	return connect.NewResponse(&cloud_drivev1.GetDownloadResponse{Item: itemProto(item), DownloadUrl: "/v1/items/" + item.UID + "/content"}), nil
}

// StartUpload implements UploadService.StartUpload.
func (s *ConnectServer) StartUpload(ctx context.Context, request *connect.Request[cloud_drivev1.StartUploadRequest]) (*connect.Response[cloud_drivev1.StartUploadResponse], error) {
	upload, err := s.service.StartUpload(ctx, stringPointer(request.Msg.GetParentUid()), request.Msg.GetFileName(), request.Msg.GetContentType(), request.Msg.GetTotalSizeBytes())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&cloud_drivev1.StartUploadResponse{Upload: uploadProto(upload), ChunkUrl: "/v1/uploads/" + upload.UID + "/chunks"}), nil
}

// GetUpload implements UploadService.GetUpload.
func (s *ConnectServer) GetUpload(ctx context.Context, request *connect.Request[cloud_drivev1.GetUploadRequest]) (*connect.Response[cloud_drivev1.GetUploadResponse], error) {
	upload, err := s.service.GetUpload(ctx, request.Msg.GetUploadUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&cloud_drivev1.GetUploadResponse{Upload: uploadProto(upload)}), nil
}

// WriteUploadChunk implements UploadService.WriteUploadChunk.
func (s *ConnectServer) WriteUploadChunk(ctx context.Context, request *connect.Request[cloud_drivev1.WriteUploadChunkRequest]) (*connect.Response[cloud_drivev1.WriteUploadChunkResponse], error) {
	upload, err := s.service.WriteUploadChunk(ctx, request.Msg.GetUploadUid(), request.Msg.GetStartOffset(), request.Msg.GetTotalSizeBytes(), request.Msg.GetData())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&cloud_drivev1.WriteUploadChunkResponse{Upload: uploadProto(upload)}), nil
}

// CompleteUpload implements UploadService.CompleteUpload.
func (s *ConnectServer) CompleteUpload(ctx context.Context, request *connect.Request[cloud_drivev1.CompleteUploadRequest]) (*connect.Response[cloud_drivev1.CompleteUploadResponse], error) {
	item, err := s.service.CompleteUpload(ctx, request.Msg.GetUploadUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&cloud_drivev1.CompleteUploadResponse{Item: itemProto(item)}), nil
}

// CreateShareLink implements ShareService.CreateShareLink.
func (s *ConnectServer) CreateShareLink(ctx context.Context, request *connect.Request[cloud_drivev1.CreateShareLinkRequest]) (*connect.Response[cloud_drivev1.CreateShareLinkResponse], error) {
	expireTime, err := timestampFromProto(request.Msg.GetExpireTime())
	if err != nil {
		return nil, connectError(fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
	}
	share, shareURL, err := s.service.CreateShare(ctx, request.Msg.GetTargetUid(), expireTime)
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&cloud_drivev1.CreateShareLinkResponse{ShareLink: shareProto(share), ShareUrl: shareURL}), nil
}

// ListShareLinks implements ShareService.ListShareLinks.
func (s *ConnectServer) ListShareLinks(ctx context.Context, request *connect.Request[cloud_drivev1.ListShareLinksRequest]) (*connect.Response[cloud_drivev1.ListShareLinksResponse], error) {
	page, err := s.service.ListShareLinks(ctx, request.Msg.GetTargetUid(), int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	links := make([]*cloud_drivev1.ShareLink, 0, len(page.ShareLinks))
	for _, share := range page.ShareLinks {
		links = append(links, shareProto(share))
	}
	return connect.NewResponse(&cloud_drivev1.ListShareLinksResponse{ShareLinks: links, NextPageToken: page.NextPageToken}), nil
}

// RevokeShareLink implements ShareService.RevokeShareLink.
func (s *ConnectServer) RevokeShareLink(ctx context.Context, request *connect.Request[cloud_drivev1.RevokeShareLinkRequest]) (*connect.Response[cloud_drivev1.RevokeShareLinkResponse], error) {
	share, err := s.service.RevokeShare(ctx, request.Msg.GetShareUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&cloud_drivev1.RevokeShareLinkResponse{ShareLink: shareProto(share)}), nil
}

// ResolveShare implements ShareService.ResolveShare.
func (s *ConnectServer) ResolveShare(ctx context.Context, request *connect.Request[cloud_drivev1.ResolveShareRequest]) (*connect.Response[cloud_drivev1.ResolveShareResponse], error) {
	share, item, err := s.service.ResolveShare(ctx, request.Msg.GetShareToken())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&cloud_drivev1.ResolveShareResponse{ShareLink: shareProto(share), Target: itemProto(item)}), nil
}

// ListSharedItems implements ShareService.ListSharedItems.
func (s *ConnectServer) ListSharedItems(ctx context.Context, request *connect.Request[cloud_drivev1.ListSharedItemsRequest]) (*connect.Response[cloud_drivev1.ListSharedItemsResponse], error) {
	page, err := s.service.ListSharedItems(ctx, request.Msg.GetShareToken(), stringPointer(request.Msg.GetParentUid()), int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	items, nextToken := pageProto(page)
	return connect.NewResponse(&cloud_drivev1.ListSharedItemsResponse{Items: items, NextPageToken: nextToken}), nil
}

// GetSharedDownload implements ShareService.GetSharedDownload.
func (s *ConnectServer) GetSharedDownload(ctx context.Context, request *connect.Request[cloud_drivev1.GetSharedDownloadRequest]) (*connect.Response[cloud_drivev1.GetSharedDownloadResponse], error) {
	_, item, err := s.service.SharedItem(ctx, request.Msg.GetShareToken(), request.Msg.GetItemUid())
	if err != nil {
		return nil, connectError(err)
	}
	if item.Kind != domain.ItemKindFile {
		return nil, connectError(fmt.Errorf("%w: item is not a file", domain.ErrConflict))
	}
	downloadURL := "/v1/shares/" + request.Msg.GetShareToken() + "/items/" + item.UID + "/content"
	return connect.NewResponse(&cloud_drivev1.GetSharedDownloadResponse{Item: itemProto(item), DownloadUrl: downloadURL}), nil
}

// Login implements SessionService.Login.
func (s *ConnectServer) Login(_ context.Context, request *connect.Request[cloud_drivev1.LoginRequest]) (*connect.Response[cloud_drivev1.LoginResponse], error) {
	clientAddress := ""
	if values := request.Header().Values("CF-Connecting-IP"); len(values) == 1 {
		clientAddress = values[0]
	}
	cookie, err := s.sessions.Login(request.Msg.GetPassword(), clientAddress)
	request.Msg.Password = ""
	if err != nil {
		return nil, sessionError(err)
	}
	response := connect.NewResponse(&cloud_drivev1.LoginResponse{})
	response.Header().Add("Set-Cookie", cookie.String())
	return response, nil
}

// Logout implements SessionService.Logout.
func (s *ConnectServer) Logout(ctx context.Context, _ *connect.Request[cloud_drivev1.LogoutRequest]) (*connect.Response[cloud_drivev1.LogoutResponse], error) {
	if _, authenticated := auth.SessionFromContext(ctx); !authenticated {
		return nil, sessionError(auth.ErrUnauthenticated)
	}
	response := connect.NewResponse(&cloud_drivev1.LogoutResponse{})
	response.Header().Add("Set-Cookie", s.sessions.LogoutCookie().String())
	return response, nil
}

// GetSession implements SessionService.GetSession.
func (s *ConnectServer) GetSession(ctx context.Context, _ *connect.Request[cloud_drivev1.GetSessionRequest]) (*connect.Response[cloud_drivev1.GetSessionResponse], error) {
	if _, authenticated := auth.SessionFromContext(ctx); !authenticated {
		return nil, sessionError(auth.ErrUnauthenticated)
	}
	return connect.NewResponse(&cloud_drivev1.GetSessionResponse{}), nil
}

// Check implements HealthService.Check.
func (s *ConnectServer) Check(ctx context.Context, _ *connect.Request[cloud_drivev1.CheckRequest]) (*connect.Response[cloud_drivev1.CheckResponse], error) {
	if err := s.service.Ping(ctx); err != nil {
		return connect.NewResponse(&cloud_drivev1.CheckResponse{Healthy: false}), nil
	}
	return connect.NewResponse(&cloud_drivev1.CheckResponse{Healthy: true}), nil
}

func sessionError(err error) error {
	if errors.Is(err, auth.ErrUnauthenticated) {
		return connect.NewError(connect.CodeUnauthenticated, auth.ErrUnauthenticated)
	}
	return connect.NewError(connect.CodeInternal, errors.New("internal server error"))
}

func connectError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidArgument):
		return connect.NewError(connect.CodeInvalidArgument, errors.New("invalid request"))
	case errors.Is(err, domain.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, errors.New("resource not found"))
	case errors.Is(err, domain.ErrForbidden):
		return connect.NewError(connect.CodePermissionDenied, errors.New("access denied"))
	case errors.Is(err, domain.ErrResourceExhausted):
		return connect.NewError(connect.CodeResourceExhausted, errors.New("storage capacity is unavailable"))
	case errors.Is(err, domain.ErrConflict):
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("request cannot be applied"))
	case errors.Is(err, domain.ErrUnavailable):
		return connect.NewError(connect.CodeUnavailable, errors.New("service temporarily unavailable"))
	default:
		return connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}
}
