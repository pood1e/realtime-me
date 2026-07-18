package app

import "github.com/pood1e/realtime-me/services/library/internal/domain"

func musicProviderCapabilities(adapter domain.MusicProviderAdapter) []domain.MusicProviderCapability {
	capabilities := make([]domain.MusicProviderCapability, 0, 7)
	if _, supported := adapter.(domain.MusicLoginStarter); supported {
		capabilities = append(capabilities, domain.MusicProviderAccountConnection)
	}
	if _, supported := adapter.(domain.MusicCatalogSearcher); supported {
		capabilities = append(capabilities, domain.MusicProviderCatalogSearch)
	}
	if _, supported := adapter.(domain.MusicPlaybackResolver); supported {
		capabilities = append(capabilities, domain.MusicProviderPlayback)
	}
	if _, supported := adapter.(domain.MusicLyricsProvider); supported {
		capabilities = append(capabilities, domain.MusicProviderLyrics)
	}
	if _, supported := adapter.(domain.MusicPlaybackTokenProvider); supported {
		capabilities = append(capabilities, domain.MusicProviderBrowserToken)
	}
	if _, supported := adapter.(domain.MusicPlaylistImporter); supported {
		capabilities = append(capabilities, domain.MusicProviderPlaylistImport)
	}
	if _, supported := adapter.(domain.MusicTrackDownloader); supported {
		capabilities = append(capabilities, domain.MusicProviderLocalDownload)
	}
	return capabilities
}
