ALTER TABLE music_provider_connections
    DROP CONSTRAINT IF EXISTS music_provider_connections_provider_check;
ALTER TABLE music_provider_connection_attempts
    DROP CONSTRAINT IF EXISTS music_provider_connection_attempts_provider_check;
ALTER TABLE playback_history
    DROP CONSTRAINT IF EXISTS playback_history_provider_check,
    DROP CONSTRAINT IF EXISTS playback_history_source_check;
ALTER TABLE tracks
    DROP CONSTRAINT IF EXISTS tracks_source_check;
ALTER TABLE music_playlists
    DROP CONSTRAINT IF EXISTS music_playlists_provider_check;
ALTER TABLE music_playlist_tracks
    DROP CONSTRAINT IF EXISTS music_playlist_tracks_provider_check;

ALTER TABLE music_provider_connections RENAME COLUMN provider TO provider_id;
ALTER TABLE music_provider_connection_attempts RENAME COLUMN provider TO provider_id;
ALTER TABLE playback_history RENAME COLUMN provider TO provider_id;
ALTER TABLE music_playlists RENAME COLUMN provider TO provider_id;
ALTER TABLE music_playlist_tracks RENAME COLUMN provider TO provider_id;

ALTER TABLE music_provider_connections
    ADD CONSTRAINT music_provider_connections_provider_id_check
        CHECK (provider_id ~ '^[a-z][a-z0-9_.-]{0,63}$');
ALTER TABLE music_provider_connection_attempts
    ADD CONSTRAINT music_provider_connection_attempts_provider_id_check
        CHECK (provider_id ~ '^[a-z][a-z0-9_.-]{0,63}$');
ALTER TABLE playback_history
    ADD CONSTRAINT playback_history_provider_id_check
        CHECK (provider_id ~ '^[a-z][a-z0-9_.-]{0,63}$'),
    ADD CONSTRAINT playback_history_source_check
        CHECK (
            (provider_id = 'local' AND track_uid IS NOT NULL AND external_track_id = track_uid)
            OR (provider_id <> 'local' AND track_uid IS NULL)
        );
ALTER TABLE music_playlists
    ADD CONSTRAINT music_playlists_provider_id_check
        CHECK (provider_id ~ '^[a-z][a-z0-9_.-]{0,63}$');
ALTER TABLE music_playlist_tracks
    ADD CONSTRAINT music_playlist_tracks_provider_id_check
        CHECK (provider_id ~ '^[a-z][a-z0-9_.-]{0,63}$');

CREATE TABLE music_track_sources (
    provider_id TEXT NOT NULL CHECK (provider_id ~ '^[a-z][a-z0-9_.-]{0,63}$'),
    external_track_id TEXT NOT NULL CHECK (external_track_id <> ''),
    track_uid TEXT NOT NULL REFERENCES tracks(uid) ON DELETE CASCADE,
    create_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (provider_id, external_track_id)
);

CREATE INDEX music_track_sources_track_idx
    ON music_track_sources (track_uid, provider_id, external_track_id);

INSERT INTO music_track_sources (provider_id, external_track_id, track_uid)
SELECT source_provider, source_track_id, uid
FROM tracks
WHERE source_provider IS NOT NULL
ON CONFLICT (provider_id, external_track_id) DO NOTHING;

INSERT INTO music_track_sources (provider_id, external_track_id, track_uid)
SELECT provider_id, external_track_id, local_track_uid
FROM music_playlist_tracks
WHERE local_track_uid IS NOT NULL
ON CONFLICT (provider_id, external_track_id) DO NOTHING;

DROP INDEX IF EXISTS tracks_source_idx;
ALTER TABLE tracks
    DROP COLUMN source_provider,
    DROP COLUMN source_track_id;

CREATE INDEX music_playlist_tracks_local_track_idx
    ON music_playlist_tracks (local_track_uid)
    WHERE local_track_uid IS NOT NULL;
