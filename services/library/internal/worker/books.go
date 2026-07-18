package worker

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func extractPDF(ctx context.Context, source, workDir string) (string, []string, int, string, error) {
	output, err := commandOutput(ctx, "pdfinfo", source)
	if err != nil {
		return "", nil, 0, "", err
	}
	values := parseKeyValueLines(output)
	pageCount, _ := strconv.Atoi(values["Pages"])
	authors := splitArtists(values["Author"])
	coverBase := filepath.Join(workDir, "cover")
	coverPath := coverBase + ".png"
	if err := runCommand(ctx, "pdftoppm", "-f", "1", "-singlefile", "-png", "-scale-to", "800", source, coverBase); err != nil {
		coverPath = ""
	}
	return values["Title"], authors, pageCount, coverPath, nil
}

type epubContainer struct {
	RootFiles []struct {
		FullPath string `xml:"full-path,attr"`
	} `xml:"rootfiles>rootfile"`
}

type epubPackage struct {
	Title    string   `xml:"metadata>title"`
	Creators []string `xml:"metadata>creator"`
	Metadata []struct {
		Name    string `xml:"name,attr"`
		Content string `xml:"content,attr"`
	} `xml:"metadata>meta"`
	Manifest []struct {
		ID         string `xml:"id,attr"`
		Href       string `xml:"href,attr"`
		MediaType  string `xml:"media-type,attr"`
		Properties string `xml:"properties,attr"`
	} `xml:"manifest>item"`
}

func extractEPUB(ctx context.Context, source, workDir string) (string, []string, string, error) {
	archive, err := zip.OpenReader(source)
	if err != nil {
		return "", nil, "", fmt.Errorf("open EPUB: %w", err)
	}
	defer archive.Close()
	containerBody, err := readZipEntry(archive.File, "META-INF/container.xml")
	if err != nil {
		return "", nil, "", err
	}
	var container epubContainer
	if err := xml.Unmarshal(containerBody, &container); err != nil || len(container.RootFiles) == 0 {
		return "", nil, "", errors.New("invalid EPUB container")
	}
	packagePath := container.RootFiles[0].FullPath
	packageBody, err := readZipEntry(archive.File, packagePath)
	if err != nil {
		return "", nil, "", err
	}
	var publication epubPackage
	if err := xml.Unmarshal(packageBody, &publication); err != nil {
		return "", nil, "", fmt.Errorf("parse EPUB package: %w", err)
	}
	coverID := ""
	for _, metadata := range publication.Metadata {
		if strings.EqualFold(metadata.Name, "cover") {
			coverID = metadata.Content
		}
	}
	coverHref := ""
	for _, item := range publication.Manifest {
		if item.ID == coverID || strings.Contains(item.Properties, "cover-image") {
			coverHref = item.Href
			break
		}
	}
	coverPath := ""
	if coverHref != "" {
		entryPath := filepath.ToSlash(filepath.Join(filepath.Dir(packagePath), coverHref))
		body, readErr := readZipEntry(archive.File, entryPath)
		if readErr == nil {
			extension := filepath.Ext(coverHref)
			if extension == "" {
				extension = ".jpg"
			}
			original := filepath.Join(workDir, "epub-cover"+extension)
			if writeErr := os.WriteFile(original, body, 0o600); writeErr == nil {
				coverPath = filepath.Join(workDir, "cover.webp")
				if commandErr := runCommand(ctx, "vipsthumbnail", original, "--size", "800x800", "--output", coverPath+"[Q=85,strip]"); commandErr != nil {
					coverPath = ""
				}
			}
		}
	}
	return cleanMetadata(publication.Title), normalized(publication.Creators), coverPath, nil
}
