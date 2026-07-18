package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type audioProbe struct {
	Duration string `json:"duration"`
	Tags     struct {
		Title       string `json:"title"`
		Artist      string `json:"artist"`
		Album       string `json:"album"`
		AlbumArtist string `json:"album_artist"`
		Track       string `json:"track"`
		Disc        string `json:"disc"`
		Date        string `json:"date"`
	} `json:"tags"`
}

func probeAudio(ctx context.Context, source string) (audioProbe, error) {
	output, err := commandOutput(ctx, "ffprobe", "-v", "error", "-show_entries", "format=duration:format_tags", "-of", "json", source)
	if err != nil {
		return audioProbe{}, err
	}
	var wrapper struct {
		Format audioProbe `json:"format"`
	}
	if err := json.Unmarshal([]byte(output), &wrapper); err != nil {
		return audioProbe{}, fmt.Errorf("parse ffprobe output: %w", err)
	}
	return wrapper.Format, nil
}

func vipsDimension(ctx context.Context, path, field string) (int, error) {
	output, err := commandOutput(ctx, "vipsheader", "-f", field, path)
	if err != nil {
		return 0, err
	}
	value, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil || value < 1 {
		return 0, fmt.Errorf("invalid %s from vipsheader", field)
	}
	return value, nil
}

func dominantColor(ctx context.Context, source, workDir string) (string, error) {
	path := filepath.Join(workDir, "dominant.png")
	if err := runCommand(ctx, "vipsthumbnail", source, "--size", "1x1", "--output", path); err != nil {
		return "", err
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	image, err := png.Decode(file)
	if err != nil {
		return "", err
	}
	r, g, b, _ := image.At(image.Bounds().Min.X, image.Bounds().Min.Y).RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8), nil
}
