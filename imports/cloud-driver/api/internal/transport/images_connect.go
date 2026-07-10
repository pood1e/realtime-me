package transport

import (
	"context"

	"connectrpc.com/connect"
	imagesv1 "example.com/cloud-drive/api/gen/cloud/images/v1"
	"example.com/cloud-drive/api/internal/app"
	"example.com/cloud-drive/api/internal/domain"
)

type imageServer struct{ service *app.ImageService }

func (s *imageServer) GetImage(ctx context.Context, request *connect.Request[imagesv1.GetImageRequest]) (*connect.Response[imagesv1.GetImageResponse], error) {
	image, err := s.service.Get(ctx, request.Msg.GetImageUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&imagesv1.GetImageResponse{Image: imageProto(image)}), nil
}

func (s *imageServer) ListImages(ctx context.Context, request *connect.Request[imagesv1.ListImagesRequest]) (*connect.Response[imagesv1.ListImagesResponse], error) {
	page, err := s.service.List(ctx, request.Msg.GetQuery(), stringPointer(request.Msg.GetAlbumUid()), request.Msg.GetTrashed(), int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	images := make([]*imagesv1.Image, 0, len(page.Images))
	for _, image := range page.Images {
		images = append(images, imageProto(image))
	}
	return connect.NewResponse(&imagesv1.ListImagesResponse{Images: images, NextPageToken: page.NextPageToken}), nil
}

func (s *imageServer) ImportImage(ctx context.Context, request *connect.Request[imagesv1.ImportImageRequest]) (*connect.Response[imagesv1.ImportImageResponse], error) {
	image, err := s.service.Import(ctx, request.Msg.GetUploadUid(), stringPointer(request.Msg.GetAlbumUid()))
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&imagesv1.ImportImageResponse{Image: imageProto(image)}), nil
}

func (s *imageServer) UpdateImage(ctx context.Context, request *connect.Request[imagesv1.UpdateImageRequest]) (*connect.Response[imagesv1.UpdateImageResponse], error) {
	image, err := s.service.Update(ctx, domain.Image{UID: request.Msg.GetImageUid(), DisplayName: request.Msg.GetDisplayName(), AlbumUID: stringPointer(request.Msg.GetAlbumUid())})
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&imagesv1.UpdateImageResponse{Image: imageProto(image)}), nil
}

func (s *imageServer) DeleteImage(ctx context.Context, request *connect.Request[imagesv1.DeleteImageRequest]) (*connect.Response[imagesv1.DeleteImageResponse], error) {
	image, err := s.service.Trash(ctx, request.Msg.GetImageUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&imagesv1.DeleteImageResponse{Image: imageProto(image)}), nil
}

func (s *imageServer) RestoreImage(ctx context.Context, request *connect.Request[imagesv1.RestoreImageRequest]) (*connect.Response[imagesv1.RestoreImageResponse], error) {
	image, err := s.service.Restore(ctx, request.Msg.GetImageUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&imagesv1.RestoreImageResponse{Image: imageProto(image)}), nil
}

func (s *imageServer) PurgeImage(ctx context.Context, request *connect.Request[imagesv1.PurgeImageRequest]) (*connect.Response[imagesv1.PurgeImageResponse], error) {
	if err := s.service.Purge(ctx, request.Msg.GetImageUid()); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&imagesv1.PurgeImageResponse{}), nil
}

func (s *imageServer) EmptyImageTrash(ctx context.Context, _ *connect.Request[imagesv1.EmptyImageTrashRequest]) (*connect.Response[imagesv1.EmptyImageTrashResponse], error) {
	if err := s.service.EmptyTrash(ctx); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&imagesv1.EmptyImageTrashResponse{}), nil
}

func (s *imageServer) RetryImageProcessing(ctx context.Context, request *connect.Request[imagesv1.RetryImageProcessingRequest]) (*connect.Response[imagesv1.RetryImageProcessingResponse], error) {
	image, err := s.service.RetryProcessing(ctx, request.Msg.GetImageUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&imagesv1.RetryImageProcessingResponse{Image: imageProto(image)}), nil
}

func (s *imageServer) ListImageAlbums(ctx context.Context, _ *connect.Request[imagesv1.ListImageAlbumsRequest]) (*connect.Response[imagesv1.ListImageAlbumsResponse], error) {
	albums, err := s.service.ListAlbums(ctx)
	if err != nil {
		return nil, connectError(err)
	}
	result := make([]*imagesv1.ImageAlbum, 0, len(albums))
	for _, album := range albums {
		result = append(result, imageAlbumProto(album))
	}
	return connect.NewResponse(&imagesv1.ListImageAlbumsResponse{Albums: result}), nil
}

func (s *imageServer) CreateImageAlbum(ctx context.Context, request *connect.Request[imagesv1.CreateImageAlbumRequest]) (*connect.Response[imagesv1.CreateImageAlbumResponse], error) {
	album, err := s.service.CreateAlbum(ctx, request.Msg.GetDisplayName())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&imagesv1.CreateImageAlbumResponse{Album: imageAlbumProto(album)}), nil
}

func (s *imageServer) DeleteImageAlbum(ctx context.Context, request *connect.Request[imagesv1.DeleteImageAlbumRequest]) (*connect.Response[imagesv1.DeleteImageAlbumResponse], error) {
	if err := s.service.DeleteAlbum(ctx, request.Msg.GetAlbumUid()); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&imagesv1.DeleteImageAlbumResponse{}), nil
}

func (s *imageServer) ListImageLinks(ctx context.Context, request *connect.Request[imagesv1.ListImageLinksRequest]) (*connect.Response[imagesv1.ListImageLinksResponse], error) {
	links, err := s.service.ListLinks(ctx, request.Msg.GetImageUid())
	if err != nil {
		return nil, connectError(err)
	}
	result := make([]*imagesv1.ImageLink, 0, len(links))
	for _, link := range links {
		result = append(result, imageLinkProto(link, s.service.PublicLinkURL(link.UID)))
	}
	return connect.NewResponse(&imagesv1.ListImageLinksResponse{ImageLinks: result}), nil
}

func (s *imageServer) CreateImageLink(ctx context.Context, request *connect.Request[imagesv1.CreateImageLinkRequest]) (*connect.Response[imagesv1.CreateImageLinkResponse], error) {
	link, err := s.service.CreateLink(ctx, request.Msg.GetImageUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&imagesv1.CreateImageLinkResponse{ImageLink: imageLinkProto(link, s.service.PublicLinkURL(link.UID))}), nil
}

func (s *imageServer) RevokeImageLink(ctx context.Context, request *connect.Request[imagesv1.RevokeImageLinkRequest]) (*connect.Response[imagesv1.RevokeImageLinkResponse], error) {
	link, err := s.service.RevokeLink(ctx, request.Msg.GetImageLinkUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&imagesv1.RevokeImageLinkResponse{ImageLink: imageLinkProto(link, s.service.PublicLinkURL(link.UID))}), nil
}
