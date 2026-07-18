ALTER TABLE uploads
    DROP CONSTRAINT IF EXISTS uploads_status_check,
    DROP CONSTRAINT IF EXISTS uploads_check;

ALTER TABLE uploads
    ADD COLUMN sealed_sha256 BYTEA,
    ADD COLUMN sealed_size_bytes BIGINT,
    ADD COLUMN sealed_content_type TEXT,
    ADD COLUMN sealed_storage_key TEXT,
    ADD COLUMN sealed_content_uid TEXT REFERENCES content_objects(uid),
    ADD COLUMN failure_code TEXT NOT NULL DEFAULT '',
    ADD COLUMN finalize_time TIMESTAMPTZ,
    ADD CONSTRAINT uploads_status_check CHECK (
        status IN ('active', 'finalizing', 'sealed', 'claimed', 'failed', 'expired')
    ),
    ADD CONSTRAINT uploads_claim_check CHECK (
        (status = 'claimed') = (claimed_resource_uid IS NOT NULL)
    ),
    ADD CONSTRAINT uploads_sealed_check CHECK (
        status <> 'sealed'
        OR (
            octet_length(sealed_sha256) = 32
            AND sealed_size_bytes = total_size_bytes
            AND sealed_content_type <> ''
            AND sealed_storage_key <> ''
            AND sealed_content_uid IS NOT NULL
            AND finalize_time IS NOT NULL
        )
    );

CREATE INDEX uploads_sealed_content_idx
    ON uploads (sealed_content_uid)
    WHERE sealed_content_uid IS NOT NULL;

ALTER TABLE processing_jobs
    DROP CONSTRAINT processing_jobs_kind_check,
    ADD CONSTRAINT processing_jobs_kind_check
        CHECK (kind IN ('book', 'track', 'image', 'wallpaper', 'music_download', 'music_artwork', 'upload_finalize'));
