package app

import "example.com/cloud-drive/api/internal/domain"

func musicProviderCapabilities(adapter domain.MusicProviderAdapter) []domain.MusicProviderCapability {
	capabilities := make([]domain.MusicProviderCapability, 0, 5)
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
	return capabilities
}
