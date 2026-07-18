package transport

import (
	"fmt"
	"time"

	booksv1 "github.com/pood1e/realtime-me/services/library/gen/cloud/books/v1"
	contentv1 "github.com/pood1e/realtime-me/services/library/gen/cloud/content/v1"
	drivev1 "github.com/pood1e/realtime-me/services/library/gen/cloud/drive/v1"
	imagesv1 "github.com/pood1e/realtime-me/services/library/gen/cloud/images/v1"
	musicv1 "github.com/pood1e/realtime-me/services/library/gen/cloud/music/v1"
	wallpapersv1 "github.com/pood1e/realtime-me/services/library/gen/cloud/wallpapers/v1"
	"github.com/pood1e/realtime-me/services/library/internal/domain"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func itemProto(item domain.Item) *drivev1.DriveItem {
	result := &drivev1.DriveItem{Uid: item.UID, Name: item.Name, Kind: itemKindProto(item.Kind), SizeBytes: item.SizeBytes,
		ContentType: item.ContentType, CreateTime: timestamppb.New(item.CreateTime), UpdateTime: timestamppb.New(item.UpdateTime)}
	if item.ParentUID != nil {
		result.ParentUid = *item.ParentUID
	}
	if item.DeleteTime != nil {
		result.DeleteTime = timestamppb.New(*item.DeleteTime)
	}
	return result
}

func itemKindProto(kind domain.ItemKind) drivev1.DriveItemKind {
	if kind == domain.ItemKindDirectory {
		return drivev1.DriveItemKind_DRIVE_ITEM_KIND_DIRECTORY
	}
	return drivev1.DriveItemKind_DRIVE_ITEM_KIND_FILE
}

func pageProto(page domain.Page) ([]*drivev1.DriveItem, string) {
	items := make([]*drivev1.DriveItem, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, itemProto(item))
	}
	return items, page.NextPageToken
}

func uploadProto(upload domain.Upload) *contentv1.Upload {
	result := &contentv1.Upload{Uid: upload.UID, FileName: upload.FileName, ContentType: upload.ContentType,
		TotalSizeBytes: upload.TotalSizeBytes, ReceivedBytes: upload.ReceivedBytes, ChunkSizeBytes: upload.ChunkSizeBytes,
		Status: uploadStatusProto(upload.Status), CreateTime: timestamppb.New(upload.CreateTime), ExpireTime: timestamppb.New(upload.ExpireTime),
		ClaimedResourceUid: upload.ClaimedUID, FailureCode: upload.FailureCode}
	for _, chunk := range upload.Chunks {
		result.Chunks = append(result.Chunks, &contentv1.UploadChunk{StartOffset: chunk.StartOffset, EndOffset: chunk.EndOffset})
	}
	return result
}

func uploadStatusProto(status domain.UploadStatus) contentv1.UploadStatus {
	switch status {
	case domain.UploadStatusActive:
		return contentv1.UploadStatus_UPLOAD_STATUS_ACTIVE
	case domain.UploadStatusFinalizing:
		return contentv1.UploadStatus_UPLOAD_STATUS_FINALIZING
	case domain.UploadStatusSealed:
		return contentv1.UploadStatus_UPLOAD_STATUS_SEALED
	case domain.UploadStatusClaimed:
		return contentv1.UploadStatus_UPLOAD_STATUS_CLAIMED
	case domain.UploadStatusFailed:
		return contentv1.UploadStatus_UPLOAD_STATUS_FAILED
	case domain.UploadStatusExpired:
		return contentv1.UploadStatus_UPLOAD_STATUS_EXPIRED
	default:
		return contentv1.UploadStatus_UPLOAD_STATUS_UNSPECIFIED
	}
}

func shareProto(share domain.ShareLink) *drivev1.ShareLink {
	result := &drivev1.ShareLink{Uid: share.UID, TargetUid: share.TargetUID,
		CreateTime: timestamppb.New(share.CreateTime), ExpireTime: timestamppb.New(share.ExpireTime)}
	if share.RevokeTime != nil {
		result.RevokeTime = timestamppb.New(*share.RevokeTime)
	}
	return result
}

func bookProto(book domain.Book) *booksv1.Book {
	result := &booksv1.Book{Uid: book.UID, Title: book.Title, Authors: book.Authors, Series: book.Series,
		SeriesNumber: book.SeriesNumber, Description: book.Description, Format: bookFormatProto(book.Format),
		OriginalFileName: book.OriginalFileName, SizeBytes: book.SizeBytes, PageCount: int32(book.PageCount),
		ContentUrl: "/v1/books/" + book.UID + "/content", ProcessingStatus: processingStatusProto(book.ProcessingStatus),
		CreateTime: timestamppb.New(book.CreateTime), UpdateTime: timestamppb.New(book.UpdateTime)}
	if book.CoverStorageKey != "" {
		result.CoverUrl = "/v1/books/" + book.UID + "/cover"
	}
	if book.DeleteTime != nil {
		result.DeleteTime = timestamppb.New(*book.DeleteTime)
	}
	return result
}

func bookFormatProto(format domain.BookFormat) booksv1.BookFormat {
	if format == domain.BookFormatEPUB {
		return booksv1.BookFormat_BOOK_FORMAT_EPUB
	}
	if format == domain.BookFormatPDF {
		return booksv1.BookFormat_BOOK_FORMAT_PDF
	}
	return booksv1.BookFormat_BOOK_FORMAT_UNSPECIFIED
}

func bookFormatDomain(format booksv1.BookFormat) domain.BookFormat {
	switch format {
	case booksv1.BookFormat_BOOK_FORMAT_PDF:
		return domain.BookFormatPDF
	case booksv1.BookFormat_BOOK_FORMAT_EPUB:
		return domain.BookFormatEPUB
	default:
		return ""
	}
}

func processingStatusProto(status domain.ProcessingStatus) contentv1.ProcessingStatus {
	switch status {
	case domain.ProcessingStatusPending:
		return contentv1.ProcessingStatus_PROCESSING_STATUS_PENDING
	case domain.ProcessingStatusReady:
		return contentv1.ProcessingStatus_PROCESSING_STATUS_READY
	case domain.ProcessingStatusFailed:
		return contentv1.ProcessingStatus_PROCESSING_STATUS_FAILED
	default:
		return contentv1.ProcessingStatus_PROCESSING_STATUS_UNSPECIFIED
	}
}

func shelfProto(shelf domain.Shelf) *booksv1.Shelf {
	return &booksv1.Shelf{Uid: shelf.UID, DisplayName: shelf.DisplayName, BookCount: int32(shelf.BookCount), CreateTime: timestamppb.New(shelf.CreateTime)}
}

func readingProgressProto(progress domain.ReadingProgress) *booksv1.ReadingProgress {
	result := &booksv1.ReadingProgress{BookUid: progress.BookUID, ProgressPercent: progress.ProgressPercent}
	if !progress.UpdateTime.IsZero() {
		result.UpdateTime = timestamppb.New(progress.UpdateTime)
	}
	if progress.LocationKind == "pdf" {
		result.Location = &booksv1.ReadingProgress_Pdf{Pdf: &booksv1.PdfLocation{PageNumber: int32(progress.PDFPageNumber), PageCount: int32(progress.PDFPageCount)}}
	} else if progress.LocationKind == "epub" {
		result.Location = &booksv1.ReadingProgress_Epub{Epub: &booksv1.EpubLocation{Cfi: progress.EPUBCFI}}
	}
	return result
}

func readingProgressDomain(progress *booksv1.ReadingProgress) domain.ReadingProgress {
	result := domain.ReadingProgress{BookUID: progress.GetBookUid(), ProgressPercent: progress.GetProgressPercent()}
	if pdf := progress.GetPdf(); pdf != nil {
		result.LocationKind, result.PDFPageNumber, result.PDFPageCount = "pdf", int(pdf.GetPageNumber()), int(pdf.GetPageCount())
	}
	if epub := progress.GetEpub(); epub != nil {
		result.LocationKind, result.EPUBCFI = "epub", epub.GetCfi()
	}
	return result
}

func trackProto(track domain.Track) *musicv1.Track {
	result := &musicv1.Track{Uid: track.UID, Title: track.Title, Artists: track.Artists, Album: track.Album,
		AlbumArtist: track.AlbumArtist, TrackNumber: int32(track.TrackNumber), DiscNumber: int32(track.DiscNumber),
		Year: int32(track.Year), Duration: durationpb.New(track.Duration), OriginalFileName: track.OriginalFileName,
		ContentType: track.ContentType, SizeBytes: track.SizeBytes, ContentUrl: "/v1/tracks/" + track.UID + "/content",
		Favorite: track.Favorite, ProcessingStatus: processingStatusProto(track.ProcessingStatus),
		CreateTime: timestamppb.New(track.CreateTime), UpdateTime: timestamppb.New(track.UpdateTime)}
	if track.ArtworkStorageKey != "" {
		result.ArtworkUrl = "/v1/tracks/" + track.UID + "/artwork"
	}
	if track.DeleteTime != nil {
		result.DeleteTime = timestamppb.New(*track.DeleteTime)
	}
	return result
}

func playbackProto(entry domain.PlaybackEntry) *musicv1.PlaybackEntry {
	return &musicv1.PlaybackEntry{Uid: entry.UID, Track: playableTrackProto(entry.Track), PlayTime: timestamppb.New(entry.PlayTime)}
}

func imageProto(image domain.Image) *imagesv1.Image {
	result := &imagesv1.Image{Uid: image.UID, DisplayName: image.DisplayName, OriginalFileName: image.OriginalFileName,
		ContentType: image.ContentType, SizeBytes: image.SizeBytes, Width: int32(image.Width), Height: int32(image.Height),
		OriginalUrl: "/v1/images/" + image.UID + "/original", ProcessingStatus: processingStatusProto(image.ProcessingStatus),
		CreateTime: timestamppb.New(image.CreateTime), UpdateTime: timestamppb.New(image.UpdateTime)}
	if image.AlbumUID != nil {
		result.AlbumUid = *image.AlbumUID
	}
	if image.PreviewStorageKey != "" {
		result.PreviewUrl = "/v1/images/" + image.UID + "/preview"
	}
	if image.DeleteTime != nil {
		result.DeleteTime = timestamppb.New(*image.DeleteTime)
	}
	return result
}

func imageAlbumProto(album domain.ImageAlbum) *imagesv1.ImageAlbum {
	return &imagesv1.ImageAlbum{Uid: album.UID, DisplayName: album.DisplayName, ImageCount: int32(album.ImageCount), CreateTime: timestamppb.New(album.CreateTime)}
}

func imageLinkProto(link domain.ImageLink, publicURL string) *imagesv1.ImageLink {
	result := &imagesv1.ImageLink{Uid: link.UID, ImageUid: link.ImageUID, PublicUrl: publicURL, CreateTime: timestamppb.New(link.CreateTime)}
	if link.RevokeTime != nil {
		result.RevokeTime = timestamppb.New(*link.RevokeTime)
	}
	return result
}

func wallpaperProto(wallpaper domain.Wallpaper, originalURL string, variantURL func(int) string) *wallpapersv1.Wallpaper {
	result := &wallpapersv1.Wallpaper{Uid: wallpaper.UID, ImageUid: wallpaper.ImageUID, Title: wallpaper.Title,
		Tags: wallpaper.Tags, Orientation: wallpaperOrientationProto(wallpaper.Width, wallpaper.Height), Width: int32(wallpaper.Width),
		Height: int32(wallpaper.Height), DominantColor: wallpaper.DominantColor, OriginalUrl: originalURL,
		PublishTime: timestamppb.New(wallpaper.PublishTime), UpdateTime: timestamppb.New(wallpaper.UpdateTime)}
	for _, variant := range wallpaper.Variants {
		result.Variants = append(result.Variants, &wallpapersv1.WallpaperVariant{Width: int32(variant.Width), Height: int32(variant.Height), Url: variantURL(variant.Width)})
	}
	return result
}

func wallpaperOrientationProto(width, height int) wallpapersv1.WallpaperOrientation {
	if width > height {
		return wallpapersv1.WallpaperOrientation_WALLPAPER_ORIENTATION_LANDSCAPE
	}
	if height > width {
		return wallpapersv1.WallpaperOrientation_WALLPAPER_ORIENTATION_PORTRAIT
	}
	if width > 0 {
		return wallpapersv1.WallpaperOrientation_WALLPAPER_ORIENTATION_SQUARE
	}
	return wallpapersv1.WallpaperOrientation_WALLPAPER_ORIENTATION_UNSPECIFIED
}

func wallpaperOrientationDomain(value wallpapersv1.WallpaperOrientation) string {
	switch value {
	case wallpapersv1.WallpaperOrientation_WALLPAPER_ORIENTATION_LANDSCAPE:
		return "landscape"
	case wallpapersv1.WallpaperOrientation_WALLPAPER_ORIENTATION_PORTRAIT:
		return "portrait"
	case wallpapersv1.WallpaperOrientation_WALLPAPER_ORIENTATION_SQUARE:
		return "square"
	default:
		return ""
	}
}

func timestampFromProto(value *timestamppb.Timestamp) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	if err := value.CheckValid(); err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}
	result := value.AsTime().UTC()
	return &result, nil
}

func stringPointer(value string) *string { return &value }
