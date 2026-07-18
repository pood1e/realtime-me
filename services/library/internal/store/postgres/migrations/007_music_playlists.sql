ALTER TABLE processing_jobs
    DROP CONSTRAINT processing_jobs_kind_check,
    ADD CONSTRAINT processing_jobs_kind_check
        CHECK (kind IN ('book', 'track', 'image', 'wallpaper', 'music_download'));

ALTER TABLE tracks
    ADD COLUMN source_provider TEXT,
    ADD COLUMN source_track_id TEXT,
    ADD CONSTRAINT tracks_source_check CHECK (
        (source_provider IS NULL AND source_track_id IS NULL)
        OR (
            source_provider IN ('qq_music', 'netease_cloud_music')
            AND source_track_id IS NOT NULL
        )
    );

CREATE UNIQUE INDEX tracks_source_idx
    ON tracks (source_provider, source_track_id)
    WHERE source_provider IS NOT NULL;

CREATE TABLE music_playlists (
    uid TEXT PRIMARY KEY,
    provider TEXT NOT NULL CHECK (provider IN ('qq_music', 'netease_cloud_music', 'spotify')),
    external_id TEXT NOT NULL,
    display_name TEXT NOT NULL,
    artwork_url TEXT NOT NULL DEFAULT '',
    provider_url TEXT NOT NULL DEFAULT '',
    download_supported BOOLEAN NOT NULL,
    create_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, external_id)
);

CREATE INDEX music_playlists_update_idx
    ON music_playlists (update_time DESC, uid DESC);

CREATE TABLE music_playlist_tracks (
    uid TEXT PRIMARY KEY,
    playlist_uid TEXT NOT NULL REFERENCES music_playlists(uid) ON DELETE CASCADE,
    position INTEGER NOT NULL CHECK (position > 0),
    provider TEXT NOT NULL CHECK (provider IN ('qq_music', 'netease_cloud_music', 'spotify')),
    external_track_id TEXT NOT NULL,
    title TEXT NOT NULL,
    artists TEXT[] NOT NULL DEFAULT '{}',
    album TEXT NOT NULL DEFAULT '',
    duration_ms BIGINT NOT NULL DEFAULT 0 CHECK (duration_ms >= 0),
    artwork_url TEXT NOT NULL DEFAULT '',
    provider_url TEXT NOT NULL DEFAULT '',
    playable BOOLEAN NOT NULL,
    lyrics_available BOOLEAN NOT NULL,
    download_status TEXT NOT NULL DEFAULT 'not_started'
        CHECK (download_status IN ('not_started', 'pending', 'running', 'completed', 'failed')),
    local_track_uid TEXT REFERENCES tracks(uid) ON DELETE SET NULL,
    UNIQUE (playlist_uid, position)
);

CREATE INDEX music_playlist_tracks_order_idx
    ON music_playlist_tracks (playlist_uid, position, uid);

CREATE INDEX music_playlist_tracks_download_idx
    ON music_playlist_tracks (download_status, uid)
    WHERE download_status IN ('pending', 'running');
