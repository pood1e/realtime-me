package worker

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"

	"example.com/cloud-drive/api/internal/domain"
)

func (w *Worker) processBook(ctx context.Context, job domain.ProcessingJob, workDir string) error {
	book, content, err := w.store.GetBookForProcessing(ctx, job.ResourceUID)
	if err != nil {
		return err
	}
	source, err := w.materialize(ctx, content, workDir, extensionForContent(content.ContentType))
	if err != nil {
		return err
	}
	var title string
	var authors []string
	var pageCount int
	var coverPath string
	switch book.Format {
	case domain.BookFormatPDF:
		title, authors, pageCount, coverPath, err = extractPDF(ctx, source, workDir)
	case domain.BookFormatEPUB:
		title, authors, coverPath, err = extractEPUB(ctx, source, workDir)
	default:
		err = errors.New("unsupported book format")
	}
	if err != nil {
		return err
	}
	var cover *domain.Artifact
	if coverPath != "" {
		storageKey, err := w.files.PublishArtifact(coverPath, content.UID, "book_cover", "default", filepath.Ext(coverPath))
		if err != nil {
			return err
		}
		cover = &domain.Artifact{UID: uuid.NewString(), ContentUID: content.UID, Kind: "book_cover", Variant: "default", ContentType: contentTypeForExtension(filepath.Ext(coverPath)), StorageKey: storageKey}
	}
	return w.store.CompleteBookProcessing(ctx, book.UID, title, authors, pageCount, cover)
}

func (w *Worker) processTrack(ctx context.Context, job domain.ProcessingJob, workDir string) error {
	track, content, err := w.store.GetTrackForProcessing(ctx, job.ResourceUID)
	if err != nil {
		return err
	}
	source, err := w.materialize(ctx, content, workDir, extensionForContent(content.ContentType))
	if err != nil {
		return err
	}
	metadata, err := probeAudio(ctx, source)
	if err != nil {
		return err
	}
	track.Title = firstNonEmpty(metadata.Tags.Title, track.Title)
	track.Artists = splitArtists(metadata.Tags.Artist)
	track.Album = metadata.Tags.Album
	track.AlbumArtist = metadata.Tags.AlbumArtist
	track.TrackNumber = leadingInteger(metadata.Tags.Track)
	track.DiscNumber = leadingInteger(metadata.Tags.Disc)
	track.Year = leadingInteger(metadata.Tags.Date)
	seconds, _ := strconv.ParseFloat(metadata.Duration, 64)
	track.Duration = time.Duration(seconds * float64(time.Second))
	var artwork *domain.Artifact
	artworkPath := filepath.Join(workDir, "artwork.jpg")
	if err := runCommand(ctx, "ffmpeg", "-y", "-v", "error", "-i", source, "-map", "0:v:0", "-frames:v", "1", artworkPath); err == nil {
		storageKey, err := w.files.PublishArtifact(artworkPath, content.UID, "track_artwork", "default", "jpg")
		if err != nil {
			return err
		}
		artwork = &domain.Artifact{UID: uuid.NewString(), ContentUID: content.UID, Kind: "track_artwork", Variant: "default", ContentType: "image/jpeg", StorageKey: storageKey}
	}
	return w.store.CompleteTrackProcessing(ctx, track, artwork)
}

func (w *Worker) processImage(ctx context.Context, job domain.ProcessingJob, workDir string) error {
	image, content, err := w.store.GetImageForProcessing(ctx, job.ResourceUID)
	if err != nil {
		return err
	}
	source, err := w.materialize(ctx, content, workDir, extensionForContent(content.ContentType))
	if err != nil {
		return err
	}
	width, err := vipsDimension(ctx, source, "width")
	if err != nil {
		return err
	}
	height, err := vipsDimension(ctx, source, "height")
	if err != nil {
		return err
	}
	previewPath := filepath.Join(workDir, "preview.webp")
	if err := runCommand(ctx, "vipsthumbnail", source, "--size", "1600x1600", "--output", previewPath+"[Q=82,strip]"); err != nil {
		return err
	}
	storageKey, err := w.files.PublishArtifact(previewPath, content.UID, "image_preview", "default", "webp")
	if err != nil {
		return err
	}
	preview := &domain.Artifact{UID: uuid.NewString(), ContentUID: content.UID, Kind: "image_preview", Variant: "default", ContentType: "image/webp", StorageKey: storageKey, Width: width, Height: height}
	return w.store.CompleteImageProcessing(ctx, image.UID, width, height, preview)
}

func (w *Worker) processWallpaper(ctx context.Context, job domain.ProcessingJob, workDir string) error {
	wallpaper, content, err := w.store.GetWallpaperForProcessing(ctx, job.ResourceUID)
	if err != nil {
		return err
	}
	source, err := w.materialize(ctx, content, workDir, extensionForContent(content.ContentType))
	if err != nil {
		return err
	}
	variants := make([]domain.Artifact, 0, 3)
	for _, requestedWidth := range []int{640, 1280, 1920} {
		target := filepath.Join(workDir, fmt.Sprintf("wallpaper-%d.webp", requestedWidth))
		if err := runCommand(ctx, "vipsthumbnail", source, "--size", fmt.Sprintf("%dx", requestedWidth), "--output", target+"[Q=85,strip]"); err != nil {
			return err
		}
		width, err := vipsDimension(ctx, target, "width")
		if err != nil {
			return err
		}
		height, err := vipsDimension(ctx, target, "height")
		if err != nil {
			return err
		}
		storageKey, err := w.files.PublishArtifact(target, content.UID, "wallpaper", strconv.Itoa(requestedWidth), "webp")
		if err != nil {
			return err
		}
		variants = append(variants, domain.Artifact{UID: uuid.NewString(), ContentUID: content.UID, Kind: "wallpaper", Variant: strconv.Itoa(requestedWidth), ContentType: "image/webp", StorageKey: storageKey, Width: width, Height: height})
	}
	color, err := dominantColor(ctx, source, workDir)
	if err != nil {
		w.logger.Warn("dominant color extraction failed", "wallpaper_id", wallpaper.UID, "error", err)
	}
	return w.store.CompleteWallpaperProcessing(ctx, wallpaper.UID, color, variants)
}
