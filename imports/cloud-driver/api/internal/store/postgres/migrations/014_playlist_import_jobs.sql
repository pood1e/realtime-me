CREATE TABLE music_playlist_imports (
    uid TEXT PRIMARY KEY,
    provider_id TEXT NOT NULL CHECK (provider_id ~ '^[a-z][a-z0-9_.-]{0,63}$'),
    source TEXT NOT NULL CHECK (source <> '' AND length(source) <= 2048),
    status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    playlist_uid TEXT REFERENCES music_playlists(uid) ON DELETE SET NULL,
    failure_code TEXT NOT NULL DEFAULT '',
    create_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX music_playlist_imports_update_idx
    ON music_playlist_imports (update_time, uid);

ALTER TABLE processing_jobs
    DROP CONSTRAINT processing_jobs_kind_check,
    ADD CONSTRAINT processing_jobs_kind_check
        CHECK (kind IN ('book', 'track', 'image', 'wallpaper', 'music_download', 'music_artwork', 'upload_finalize', 'playlist_import'));
