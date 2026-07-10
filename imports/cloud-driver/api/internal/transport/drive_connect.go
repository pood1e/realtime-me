package transport

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	drivev1 "example.com/cloud-drive/api/gen/cloud/drive/v1"
	"example.com/cloud-drive/api/internal/app"
	"example.com/cloud-drive/api/internal/domain"
)

type driveServer struct{ service *app.DriveService }

func (s *driveServer) GetDriveItem(ctx context.Context, request *connect.Request[drivev1.GetDriveItemRequest]) (*connect.Response[drivev1.GetDriveItemResponse], error) {
	item, err := s.service.GetItem(ctx, request.Msg.GetItemUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&drivev1.GetDriveItemResponse{Item: itemProto(item)}), nil
}

func (s *driveServer) ListDriveItems(ctx context.Context, request *connect.Request[drivev1.ListDriveItemsRequest]) (*connect.Response[drivev1.ListDriveItemsResponse], error) {
	page, err := s.service.ListItems(ctx, stringPointer(request.Msg.GetParentUid()), request.Msg.GetIncludeTrashed(), int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	items, next := pageProto(page)
	return connect.NewResponse(&drivev1.ListDriveItemsResponse{Items: items, NextPageToken: next}), nil
}

func (s *driveServer) ListTrashedItems(ctx context.Context, request *connect.Request[drivev1.ListTrashedItemsRequest]) (*connect.Response[drivev1.ListTrashedItemsResponse], error) {
	page, err := s.service.ListTrashedItems(ctx, int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	items, next := pageProto(page)
	return connect.NewResponse(&drivev1.ListTrashedItemsResponse{Items: items, NextPageToken: next}), nil
}

func (s *driveServer) SearchDriveItems(ctx context.Context, request *connect.Request[drivev1.SearchDriveItemsRequest]) (*connect.Response[drivev1.SearchDriveItemsResponse], error) {
	page, err := s.service.SearchItems(ctx, request.Msg.GetQuery(), int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	items, next := pageProto(page)
	return connect.NewResponse(&drivev1.SearchDriveItemsResponse{Items: items, NextPageToken: next}), nil
}

func (s *driveServer) CreateDirectory(ctx context.Context, request *connect.Request[drivev1.CreateDirectoryRequest]) (*connect.Response[drivev1.CreateDirectoryResponse], error) {
	item, err := s.service.CreateDirectory(ctx, stringPointer(request.Msg.GetParentUid()), request.Msg.GetName())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&drivev1.CreateDirectoryResponse{Item: itemProto(item)}), nil
}

func (s *driveServer) RenameDriveItem(ctx context.Context, request *connect.Request[drivev1.RenameDriveItemRequest]) (*connect.Response[drivev1.RenameDriveItemResponse], error) {
	item, err := s.service.RenameItem(ctx, request.Msg.GetItemUid(), request.Msg.GetName())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&drivev1.RenameDriveItemResponse{Item: itemProto(item)}), nil
}

func (s *driveServer) MoveDriveItem(ctx context.Context, request *connect.Request[drivev1.MoveDriveItemRequest]) (*connect.Response[drivev1.MoveDriveItemResponse], error) {
	item, err := s.service.MoveItem(ctx, request.Msg.GetItemUid(), stringPointer(request.Msg.GetParentUid()))
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&drivev1.MoveDriveItemResponse{Item: itemProto(item)}), nil
}

func (s *driveServer) DeleteDriveItem(ctx context.Context, request *connect.Request[drivev1.DeleteDriveItemRequest]) (*connect.Response[drivev1.DeleteDriveItemResponse], error) {
	item, err := s.service.TrashItem(ctx, request.Msg.GetItemUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&drivev1.DeleteDriveItemResponse{Item: itemProto(item)}), nil
}

func (s *driveServer) RestoreDriveItem(ctx context.Context, request *connect.Request[drivev1.RestoreDriveItemRequest]) (*connect.Response[drivev1.RestoreDriveItemResponse], error) {
	item, err := s.service.RestoreItem(ctx, request.Msg.GetItemUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&drivev1.RestoreDriveItemResponse{Item: itemProto(item)}), nil
}

func (s *driveServer) PurgeDriveItem(ctx context.Context, request *connect.Request[drivev1.PurgeDriveItemRequest]) (*connect.Response[drivev1.PurgeDriveItemResponse], error) {
	if err := s.service.PurgeItem(ctx, request.Msg.GetItemUid()); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&drivev1.PurgeDriveItemResponse{}), nil
}

func (s *driveServer) EmptyTrash(ctx context.Context, _ *connect.Request[drivev1.EmptyTrashRequest]) (*connect.Response[drivev1.EmptyTrashResponse], error) {
	if err := s.service.EmptyTrash(ctx); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&drivev1.EmptyTrashResponse{}), nil
}

func (s *driveServer) ImportDriveFile(ctx context.Context, request *connect.Request[drivev1.ImportDriveFileRequest]) (*connect.Response[drivev1.ImportDriveFileResponse], error) {
	item, err := s.service.ImportFile(ctx, request.Msg.GetUploadUid(), stringPointer(request.Msg.GetParentUid()), request.Msg.GetName())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&drivev1.ImportDriveFileResponse{Item: itemProto(item)}), nil
}

func (s *driveServer) GetDownload(ctx context.Context, request *connect.Request[drivev1.GetDownloadRequest]) (*connect.Response[drivev1.GetDownloadResponse], error) {
	item, err := s.service.GetItem(ctx, request.Msg.GetItemUid())
	if err != nil {
		return nil, connectError(err)
	}
	if item.Kind != domain.ItemKindFile {
		return nil, connectError(fmt.Errorf("%w: item is not a file", domain.ErrConflict))
	}
	return connect.NewResponse(&drivev1.GetDownloadResponse{Item: itemProto(item), DownloadUrl: "/v1/items/" + item.UID + "/content"}), nil
}

type shareServer struct{ service *app.DriveService }

func (s *shareServer) CreateShareLink(ctx context.Context, request *connect.Request[drivev1.CreateShareLinkRequest]) (*connect.Response[drivev1.CreateShareLinkResponse], error) {
	expireTime, err := timestampFromProto(request.Msg.GetExpireTime())
	if err != nil {
		return nil, connectError(fmt.Errorf("%w: invalid expiry", domain.ErrInvalidArgument))
	}
	share, url, err := s.service.CreateShare(ctx, request.Msg.GetTargetUid(), expireTime)
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&drivev1.CreateShareLinkResponse{ShareLink: shareProto(share), ShareUrl: url}), nil
}

func (s *shareServer) ListShareLinks(ctx context.Context, request *connect.Request[drivev1.ListShareLinksRequest]) (*connect.Response[drivev1.ListShareLinksResponse], error) {
	page, err := s.service.ListShareLinks(ctx, request.Msg.GetTargetUid(), int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	links := make([]*drivev1.ShareLink, 0, len(page.ShareLinks))
	for _, link := range page.ShareLinks {
		links = append(links, shareProto(link))
	}
	return connect.NewResponse(&drivev1.ListShareLinksResponse{ShareLinks: links, NextPageToken: page.NextPageToken}), nil
}

func (s *shareServer) RevokeShareLink(ctx context.Context, request *connect.Request[drivev1.RevokeShareLinkRequest]) (*connect.Response[drivev1.RevokeShareLinkResponse], error) {
	share, err := s.service.RevokeShare(ctx, request.Msg.GetShareUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&drivev1.RevokeShareLinkResponse{ShareLink: shareProto(share)}), nil
}

func (s *shareServer) ResolveShare(ctx context.Context, request *connect.Request[drivev1.ResolveShareRequest]) (*connect.Response[drivev1.ResolveShareResponse], error) {
	share, item, err := s.service.ResolveShare(ctx, request.Msg.GetShareToken())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&drivev1.ResolveShareResponse{ShareLink: shareProto(share), Target: itemProto(item)}), nil
}

func (s *shareServer) ListSharedItems(ctx context.Context, request *connect.Request[drivev1.ListSharedItemsRequest]) (*connect.Response[drivev1.ListSharedItemsResponse], error) {
	page, err := s.service.ListSharedItems(ctx, request.Msg.GetShareToken(), stringPointer(request.Msg.GetParentUid()), int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	items, next := pageProto(page)
	return connect.NewResponse(&drivev1.ListSharedItemsResponse{Items: items, NextPageToken: next}), nil
}

func (s *shareServer) GetSharedDownload(ctx context.Context, request *connect.Request[drivev1.GetSharedDownloadRequest]) (*connect.Response[drivev1.GetSharedDownloadResponse], error) {
	_, item, err := s.service.SharedItem(ctx, request.Msg.GetShareToken(), request.Msg.GetItemUid())
	if err != nil {
		return nil, connectError(err)
	}
	if item.Kind != domain.ItemKindFile {
		return nil, connectError(fmt.Errorf("%w: item is not a file", domain.ErrConflict))
	}
	url := "/v1/shares/" + request.Msg.GetShareToken() + "/items/" + item.UID + "/content"
	return connect.NewResponse(&drivev1.GetSharedDownloadResponse{Item: itemProto(item), DownloadUrl: url}), nil
}
