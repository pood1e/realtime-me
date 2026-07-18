CREATE TABLE music_provider_connections (
    provider TEXT PRIMARY KEY CHECK (provider IN ('qq_music', 'netease_cloud_music', 'spotify')),
    status TEXT NOT NULL CHECK (status IN ('connected', 'reconnect_required', 'unavailable')),
    account_id TEXT NOT NULL,
    display_name TEXT NOT NULL,
    avatar_url TEXT NOT NULL DEFAULT '',
    membership TEXT NOT NULL DEFAULT '',
    membership_expire_time TIMESTAMPTZ,
    encrypted_credentials BYTEA NOT NULL,
    create_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE music_provider_connection_attempts (
    uid TEXT PRIMARY KEY,
    provider TEXT NOT NULL CHECK (provider IN ('qq_music', 'netease_cloud_music', 'spotify')),
    status TEXT NOT NULL CHECK (status IN ('waiting', 'scanned', 'connected', 'expired', 'refused', 'failed')),
    qr_image BYTEA,
    qr_content_type TEXT NOT NULL DEFAULT '',
    qr_payload TEXT NOT NULL DEFAULT '',
    authorization_url TEXT NOT NULL DEFAULT '',
    state_hash BYTEA UNIQUE,
    encrypted_state BYTEA NOT NULL,
    create_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    expire_time TIMESTAMPTZ NOT NULL,
    consumed_time TIMESTAMPTZ
);

CREATE INDEX music_provider_attempt_expire_idx
    ON music_provider_connection_attempts (expire_time)
    WHERE consumed_time IS NULL;

ALTER TABLE playback_history
    ADD COLUMN provider TEXT NOT NULL DEFAULT 'local',
    ADD COLUMN external_track_id TEXT,
    ADD COLUMN title TEXT,
    ADD COLUMN artists TEXT[],
    ADD COLUMN album TEXT,
    ADD COLUMN duration_ms BIGINT,
    ADD COLUMN artwork_url TEXT,
    ADD COLUMN provider_url TEXT,
    ADD COLUMN playable BOOLEAN,
    ADD COLUMN lyrics_available BOOLEAN;

UPDATE playback_history history
SET external_track_id = track.uid,
    title = track.title,
    artists = track.artists,
    album = track.album,
    duration_ms = track.duration_ms,
    artwork_url = CASE
        WHEN EXISTS (
            SELECT 1 FROM content_artifacts artifact
            WHERE artifact.content_uid = track.content_uid
              AND artifact.kind = 'track_artwork'
              AND artifact.variant = 'default'
        ) THEN '/v1/tracks/' || track.uid || '/artwork'
        ELSE ''
    END,
    provider_url = '',
    playable = TRUE,
    lyrics_available = FALSE
FROM tracks track
WHERE history.track_uid = track.uid;

ALTER TABLE playback_history
    ALTER COLUMN track_uid DROP NOT NULL,
    ALTER COLUMN external_track_id SET NOT NULL,
    ALTER COLUMN title SET NOT NULL,
    ALTER COLUMN artists SET NOT NULL,
    ALTER COLUMN album SET NOT NULL,
    ALTER COLUMN duration_ms SET NOT NULL,
    ALTER COLUMN artwork_url SET NOT NULL,
    ALTER COLUMN provider_url SET NOT NULL,
    ALTER COLUMN playable SET NOT NULL,
    ALTER COLUMN lyrics_available SET NOT NULL,
    ADD CONSTRAINT playback_history_provider_check
        CHECK (provider IN ('local', 'qq_music', 'netease_cloud_music', 'spotify')),
    ADD CONSTRAINT playback_history_source_check
        CHECK (
            (provider = 'local' AND track_uid IS NOT NULL AND external_track_id = track_uid)
            OR (provider <> 'local' AND track_uid IS NULL)
        );
