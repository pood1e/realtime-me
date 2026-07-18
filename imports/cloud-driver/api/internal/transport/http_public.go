package transport

import (
	"net/http"
	"strconv"
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
