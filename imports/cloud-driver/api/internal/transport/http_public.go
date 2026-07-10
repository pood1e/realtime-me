package transport

import (
	"fmt"
	"net/http"
	"strconv"

	drivev1 "example.com/cloud-drive/api/gen/cloud/drive/v1"
	"example.com/cloud-drive/api/internal/domain"
)

func (router *httpRouter) servePublicImage(writer http.ResponseWriter, request *http.Request) {
	if !getOrHead(writer, request) {
		return
	}
	segments := pathSegments(request.URL.Path, "/i/")
	if len(segments) != 1 {
		http.NotFound(writer, request)
		return
	}
	file, image, contentType, err := router.suite.Images.OpenPublicLink(request.Context(), segments[0])
	if err != nil {
		writeDomainError(writer, err, true)
		return
	}
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
	serveFile(writer, request, file, image.OriginalFileName, contentType, image.UpdateTime, "no-store", false)
}

func (router *httpRouter) serveWallpaperFile(writer http.ResponseWriter, request *http.Request) {
	if !getOrHead(writer, request) {
		return
	}
	segments := pathSegments(request.URL.Path, "/v1/wallpapers/")
	if len(segments) != 2 {
		http.NotFound(writer, request)
		return
	}
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
	if segments[1] == "original" {
		file, wallpaper, err := router.suite.Wallpapers.OpenOriginal(request.Context(), segments[0])
		if err != nil {
			writeDomainError(writer, err, true)
			return
		}
		serveFile(writer, request, file, wallpaper.Title, wallpaper.ContentType, wallpaper.UpdateTime, "public, max-age=300, s-maxage=3600", wallpaper.ContentType == "image/svg+xml")
		return
	}
	width, err := strconv.Atoi(segments[1])
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	file, wallpaper, variant, err := router.suite.Wallpapers.OpenVariant(request.Context(), segments[0], width)
	if err != nil {
		writeDomainError(writer, err, true)
		return
	}
	serveFile(writer, request, file, wallpaper.Title+".webp", variant.ContentType, wallpaper.UpdateTime, "public, max-age=300, s-maxage=3600", false)
}

func (router *httpRouter) servePublicShareRaw(writer http.ResponseWriter, request *http.Request) {
	if !getOrHead(writer, request) {
		return
	}
	segments := pathSegments(request.URL.Path, "/v1/shares/")
	if len(segments) == 1 {
		share, item, err := router.suite.Drive.ResolveShare(request.Context(), segments[0])
		if err != nil {
			writeDomainError(writer, err, true)
			return
		}
		writeProtoJSON(writer, http.StatusOK, &drivev1.ResolveShareResponse{ShareLink: shareProto(share), Target: itemProto(item)})
		return
	}
	if len(segments) == 2 && segments[1] == "items" {
		pageSize, err := queryPageSize(request.URL.Query().Get("pageSize"))
		if err != nil {
			writeDomainError(writer, err, true)
			return
		}
		parentUID := request.URL.Query().Get("parentUid")
		page, err := router.suite.Drive.ListSharedItems(
			request.Context(),
			segments[0],
			&parentUID,
			pageSize,
			request.URL.Query().Get("pageToken"),
		)
		if err != nil {
			writeDomainError(writer, err, true)
			return
		}
		items := make([]*drivev1.DriveItem, 0, len(page.Items))
		for _, item := range page.Items {
			items = append(items, itemProto(item))
		}
		writeProtoJSON(writer, http.StatusOK, &drivev1.ListSharedItemsResponse{
			Items: items, NextPageToken: page.NextPageToken,
		})
		return
	}
	if len(segments) == 4 && segments[1] == "items" && segments[3] == "content" {
		_, item, err := router.suite.Drive.SharedItem(request.Context(), segments[0], segments[2])
		if err != nil {
			writeDomainError(writer, err, true)
			return
		}
		file, err := router.suite.Drive.OpenItem(request.Context(), item)
		if err != nil {
			writeDomainError(writer, err, true)
			return
		}
		serveFile(writer, request, file, item.Name, item.ContentType, item.UpdateTime, "private, no-store", false)
		return
	}
	http.NotFound(writer, request)
}

func publicAssetPreflight(writer http.ResponseWriter, request *http.Request) bool {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
	if request.Method != http.MethodOptions {
		return false
	}
	writer.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	writer.Header().Set("Access-Control-Allow-Headers", "Range")
	writer.WriteHeader(http.StatusNoContent)
	return true
}

func queryPageSize(value string) (int, error) {
	if value == "" {
		return 200, nil
	}
	pageSize, err := strconv.Atoi(value)
	if err != nil || pageSize <= 0 {
		return 0, fmt.Errorf("%w: invalid page size", domain.ErrInvalidArgument)
	}
	return pageSize, nil
}
