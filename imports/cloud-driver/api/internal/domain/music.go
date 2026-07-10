package domain

import "time"

// Track is one private audio catalog entry.
type Track struct {
	UID               string
	ContentUID        string
	Title             string
	Artists           []string
	Album             string
	AlbumArtist       string
	TrackNumber       int
	DiscNumber        int
	Year              int
	Duration          time.Duration
	OriginalFileName  string
	ContentType       string
	SizeBytes         int64
	ArtworkStorageKey string
	Favorite          bool
	ProcessingStatus  ProcessingStatus
	CreateTime        time.Time
	UpdateTime        time.Time
	DeleteTime        *time.Time
}

// TrackPage is one cursor page of tracks.
type TrackPage struct {
	Tracks        []Track
	NextPageToken string
}

// Album summarizes tracks sharing an album and album artist.
type Album struct {
	UID               string
	Title             string
	AlbumArtist       string
	Year              int
	TrackCount        int
	ArtworkStorageKey string
}

// Artist summarizes one display artist.
type Artist struct {
	UID         string
	DisplayName string
	TrackCount  int
}

// PlaybackEntry records one meaningful playback.
type PlaybackEntry struct {
	UID      string
	Track    Track
	PlayTime time.Time
}

// PlaybackPage is one cursor page of playback entries.
type PlaybackPage struct {
	Entries       []PlaybackEntry
	NextPageToken string
}
