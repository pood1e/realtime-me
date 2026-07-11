package transport

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	contentv1 "example.com/cloud-drive/api/gen/cloud/content/v1"
	"example.com/cloud-drive/api/internal/domain"
)

func (router *httpRouter) servePrivateRaw(writer http.ResponseWriter, request *http.Request) {
	switch {
	case strings.HasPrefix(request.URL.Path, "/v1/uploads/"):
		router.servePrivateUpload(writer, request)
	case strings.HasPrefix(request.URL.Path, "/v1/items/"):
		router.serveDriveContent(writer, request)
	case strings.HasPrefix(request.URL.Path, "/v1/books/"):
		router.serveBookFile(writer, request)
	case strings.HasPrefix(request.URL.Path, "/v1/tracks/"):
		router.serveTrackFile(writer, request)
	case strings.HasPrefix(request.URL.Path, "/v1/images/"):
		router.servePrivateImageFile(writer, request)
	default:
		http.NotFound(writer, request)
	}
}

func (router *httpRouter) servePrivateUpload(writer http.ResponseWriter, request *http.Request) {
	segments := pathSegments(request.URL.Path, "/v1/uploads/")
	if len(segments) == 2 && segments[1] == "chunks" && request.Method == http.MethodPut {
		router.writeRawChunk(writer, request, segments[0])
		return
	}
	if len(segments) == 1 && request.Method == http.MethodGet {
		upload, err := router.suite.Content.GetUpload(request.Context(), segments[0])
		if err != nil {
			writeDomainError(writer, err, false)
			return
		}
		writeProtoJSON(writer, http.StatusOK, &contentv1.GetUploadResponse{Upload: uploadProto(upload)})
		return
	}
	methodNotAllowed(writer, http.MethodGet, http.MethodPut)
}

func (router *httpRouter) writeRawChunk(writer http.ResponseWriter, request *http.Request, uploadUID string) {
	start, end, total, err := parseContentRange(request.Header.Get("Content-Range"))
	if err != nil || (request.ContentLength >= 0 && request.ContentLength != end-start+1) {
		writeDomainError(writer, fmt.Errorf("%w: invalid upload range", domain.ErrInvalidArgument), false)
		return
	}
	request.Body = http.MaxBytesReader(writer, request.Body, maxChunkBodyBytes+1)
	data, err := io.ReadAll(request.Body)
	if err != nil || int64(len(data)) != end-start+1 {
		writeDomainError(writer, fmt.Errorf("%w: chunk body does not match range", domain.ErrInvalidArgument), false)
		return
	}
	upload, err := router.suite.Content.WriteUploadChunk(request.Context(), uploadUID, start, total, data)
	if err != nil {
		writeDomainError(writer, err, false)
		return
	}
	writeProtoJSON(writer, http.StatusOK, &contentv1.WriteUploadChunkResponse{Upload: uploadProto(upload)})
}

func (router *httpRouter) serveDriveContent(writer http.ResponseWriter, request *http.Request) {
	if !getOrHead(writer, request) {
		return
	}
	segments := pathSegments(request.URL.Path, "/v1/items/")
	if len(segments) != 2 || segments[1] != "content" {
		http.NotFound(writer, request)
		return
	}
	item, err := router.suite.Drive.GetItem(request.Context(), segments[0])
	if err != nil {
		writeDomainError(writer, err, false)
		return
	}
	file, err := router.suite.Drive.OpenItem(request.Context(), item)
	if err != nil {
		writeDomainError(writer, err, false)
		return
	}
	serveFile(writer, request, file, item.Name, item.ContentType, item.UpdateTime, "private, no-store", false)
}

func (router *httpRouter) serveBookFile(writer http.ResponseWriter, request *http.Request) {
	if !getOrHead(writer, request) {
		return
	}
	segments := pathSegments(request.URL.Path, "/v1/books/")
	if len(segments) != 2 {
		http.NotFound(writer, request)
		return
	}
	var file *os.File
	var name, contentType string
	var modified time.Time
	var err error
	if segments[1] == "content" {
		var book domain.Book
		file, book, err = router.suite.Books.OpenContent(request.Context(), segments[0])
		name, modified = book.OriginalFileName, book.UpdateTime
		if book.Format == domain.BookFormatPDF {
			contentType = "application/pdf"
		} else {
			contentType = "application/epub+zip"
		}
	} else if segments[1] == "cover" {
		var book domain.Book
		file, book, err = router.suite.Books.OpenCover(request.Context(), segments[0])
		name, contentType, modified = "cover", "", book.UpdateTime
	} else {
		http.NotFound(writer, request)
		return
	}
	if err != nil {
		writeDomainError(writer, err, false)
		return
	}
	serveFile(writer, request, file, name, contentType, modified, "private, no-store", false)
}

func (router *httpRouter) serveTrackFile(writer http.ResponseWriter, request *http.Request) {
	if !getOrHead(writer, request) {
		return
	}
	segments := pathSegments(request.URL.Path, "/v1/tracks/")
	if len(segments) != 2 {
		http.NotFound(writer, request)
		return
	}
	var file *os.File
	var track domain.Track
	var err error
	contentType, name := "", "artwork"
	if segments[1] == "content" {
		file, track, err = router.suite.Music.Library.OpenContent(request.Context(), segments[0])
		contentType, name = track.ContentType, track.OriginalFileName
	} else if segments[1] == "artwork" {
		file, track, err = router.suite.Music.Library.OpenArtwork(request.Context(), segments[0])
	} else {
		http.NotFound(writer, request)
		return
	}
	if err != nil {
		writeDomainError(writer, err, false)
		return
	}
	serveFile(writer, request, file, name, contentType, track.UpdateTime, "private, no-store", false)
}

func (router *httpRouter) servePrivateImageFile(writer http.ResponseWriter, request *http.Request) {
	if !getOrHead(writer, request) {
		return
	}
	segments := pathSegments(request.URL.Path, "/v1/images/")
	if len(segments) != 2 {
		http.NotFound(writer, request)
		return
	}
	var file *os.File
	var image domain.Image
	var err error
	contentType, attachment := "image/webp", false
	if segments[1] == "original" {
		file, image, err = router.suite.Images.OpenOriginal(request.Context(), segments[0])
		contentType, attachment = image.ContentType, image.ContentType == "image/svg+xml"
	} else if segments[1] == "preview" {
		file, image, err = router.suite.Images.OpenPreview(request.Context(), segments[0])
	} else {
		http.NotFound(writer, request)
		return
	}
	if err != nil {
		writeDomainError(writer, err, false)
		return
	}
	serveFile(writer, request, file, image.OriginalFileName, contentType, image.UpdateTime, "private, no-store", attachment)
}
