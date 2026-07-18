package transport

import (
	"context"

	"connectrpc.com/connect"
	wallpapersv1 "example.com/cloud-drive/api/gen/cloud/wallpapers/v1"
	"example.com/cloud-drive/api/internal/app"
)

type wallpaperAdminServer struct{ service *app.WallpaperService }

func (s *wallpaperAdminServer) ListPublishedWallpapers(ctx context.Context, request *connect.Request[wallpapersv1.ListPublishedWallpapersRequest]) (*connect.Response[wallpapersv1.ListPublishedWallpapersResponse], error) {
	page, err := s.service.List(ctx, request.Msg.GetQuery(), request.Msg.GetTag(), wallpaperOrientationDomain(request.Msg.GetOrientation()), int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	result := make([]*wallpapersv1.Wallpaper, 0, len(page.Wallpapers))
	for _, wallpaper := range page.Wallpapers {
		result = append(result, wallpaperProto(wallpaper, s.service.OriginalURL(wallpaper.UID), func(width int) string { return s.service.VariantURL(wallpaper.UID, width) }))
	}
	return connect.NewResponse(&wallpapersv1.ListPublishedWallpapersResponse{Wallpapers: result, NextPageToken: page.NextPageToken}), nil
}

func (s *wallpaperAdminServer) PublishWallpaper(ctx context.Context, request *connect.Request[wallpapersv1.PublishWallpaperRequest]) (*connect.Response[wallpapersv1.PublishWallpaperResponse], error) {
	wallpaper, err := s.service.Publish(ctx, request.Msg.GetImageUid(), request.Msg.GetTitle(), request.Msg.GetTags())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&wallpapersv1.PublishWallpaperResponse{Wallpaper: wallpaperProto(wallpaper, s.service.OriginalURL(wallpaper.UID), func(width int) string { return s.service.VariantURL(wallpaper.UID, width) })}), nil
}

func (s *wallpaperAdminServer) UpdateWallpaper(ctx context.Context, request *connect.Request[wallpapersv1.UpdateWallpaperRequest]) (*connect.Response[wallpapersv1.UpdateWallpaperResponse], error) {
	wallpaper, err := s.service.Update(ctx, request.Msg.GetWallpaperUid(), request.Msg.GetTitle(), request.Msg.GetTags())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&wallpapersv1.UpdateWallpaperResponse{Wallpaper: wallpaperProto(wallpaper, s.service.OriginalURL(wallpaper.UID), func(width int) string { return s.service.VariantURL(wallpaper.UID, width) })}), nil
}

func (s *wallpaperAdminServer) UnpublishWallpaper(ctx context.Context, request *connect.Request[wallpapersv1.UnpublishWallpaperRequest]) (*connect.Response[wallpapersv1.UnpublishWallpaperResponse], error) {
	if err := s.service.Unpublish(ctx, request.Msg.GetWallpaperUid()); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&wallpapersv1.UnpublishWallpaperResponse{}), nil
}

type wallpaperPublicServer struct{ service *app.WallpaperService }

func (s *wallpaperPublicServer) GetWallpaper(ctx context.Context, request *connect.Request[wallpapersv1.GetWallpaperRequest]) (*connect.Response[wallpapersv1.GetWallpaperResponse], error) {
	wallpaper, err := s.service.Get(ctx, request.Msg.GetWallpaperUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&wallpapersv1.GetWallpaperResponse{Wallpaper: wallpaperProto(wallpaper, s.service.OriginalURL(wallpaper.UID), func(width int) string { return s.service.VariantURL(wallpaper.UID, width) })}), nil
}

func (s *wallpaperPublicServer) ListWallpapers(ctx context.Context, request *connect.Request[wallpapersv1.ListWallpapersRequest]) (*connect.Response[wallpapersv1.ListWallpapersResponse], error) {
	page, err := s.service.List(ctx, request.Msg.GetQuery(), request.Msg.GetTag(), wallpaperOrientationDomain(request.Msg.GetOrientation()), int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	result := make([]*wallpapersv1.Wallpaper, 0, len(page.Wallpapers))
	for _, wallpaper := range page.Wallpapers {
		result = append(result, wallpaperProto(wallpaper, s.service.OriginalURL(wallpaper.UID), func(width int) string { return s.service.VariantURL(wallpaper.UID, width) }))
	}
	return connect.NewResponse(&wallpapersv1.ListWallpapersResponse{Wallpapers: result, NextPageToken: page.NextPageToken}), nil
}
