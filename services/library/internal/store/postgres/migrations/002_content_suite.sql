CREATE TABLE content_objects (
    uid TEXT PRIMARY KEY,
    sha256 BYTEA UNIQUE,
    size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
    content_type TEXT NOT NULL,
    storage_key TEXT NOT NULL UNIQUE,
    create_time TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO content_objects (uid, size_bytes, content_type, storage_key, create_time)
SELECT storage_key, size_bytes, content_type, 'blobs/' || storage_key, create_time
FROM drive_items
WHERE kind = 'file';

ALTER TABLE drive_items ADD COLUMN content_uid TEXT REFERENCES content_objects(uid);
UPDATE drive_items SET content_uid = storage_key WHERE kind = 'file';
ALTER TABLE drive_items
    ADD CONSTRAINT drive_items_content_kind_check CHECK (
        (kind = 'directory' AND content_uid IS NULL) OR
        (kind = 'file' AND content_uid IS NOT NULL)
    );
ALTER TABLE drive_items DROP COLUMN size_bytes;
ALTER TABLE drive_items DROP COLUMN content_type;
ALTER TABLE drive_items DROP COLUMN storage_key;
CREATE INDEX drive_items_content_idx ON drive_items (content_uid) WHERE content_uid IS NOT NULL;

DROP TABLE upload_chunks;
DROP TABLE uploads;

CREATE TABLE uploads (
    uid TEXT PRIMARY KEY,
    file_name TEXT NOT NULL,
    content_type TEXT NOT NULL DEFAULT '',
    total_size_bytes BIGINT NOT NULL CHECK (total_size_bytes >= 0),
    received_bytes BIGINT NOT NULL DEFAULT 0 CHECK (received_bytes >= 0),
    chunk_size_bytes BIGINT NOT NULL CHECK (chunk_size_bytes > 0),
    status TEXT NOT NULL CHECK (status IN ('active', 'claimed', 'expired')),
    claimed_resource_uid TEXT,
    create_time TIMESTAMPTZ NOT NULL,
    expire_time TIMESTAMPTZ NOT NULL,
    claim_time TIMESTAMPTZ,
    CHECK ((status = 'claimed') = (claimed_resource_uid IS NOT NULL))
);

CREATE INDEX uploads_expire_idx ON uploads (expire_time) WHERE status = 'active';

CREATE TABLE upload_chunks (
    upload_uid TEXT NOT NULL REFERENCES uploads(uid) ON DELETE CASCADE,
    start_offset BIGINT NOT NULL CHECK (start_offset >= 0),
    end_offset BIGINT NOT NULL CHECK (end_offset > start_offset),
    checksum BYTEA NOT NULL,
    PRIMARY KEY (upload_uid, start_offset)
);

CREATE TABLE processing_jobs (
    uid TEXT PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('book', 'track', 'image', 'wallpaper')),
    resource_uid TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    available_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    lease_until TIMESTAMPTZ,
    error_code TEXT NOT NULL DEFAULT '',
    create_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (kind, resource_uid)
);

CREATE INDEX processing_jobs_claim_idx ON processing_jobs (available_time, create_time)
WHERE status IN ('pending', 'running');

CREATE TABLE worker_state (
    singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
    heartbeat_time TIMESTAMPTZ NOT NULL
);

CREATE TABLE content_artifacts (
    uid TEXT PRIMARY KEY,
    content_uid TEXT NOT NULL REFERENCES content_objects(uid) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    variant TEXT NOT NULL,
    content_type TEXT NOT NULL,
    storage_key TEXT NOT NULL UNIQUE,
    width INTEGER NOT NULL DEFAULT 0 CHECK (width >= 0),
    height INTEGER NOT NULL DEFAULT 0 CHECK (height >= 0),
    create_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (content_uid, kind, variant)
);

CREATE TABLE books (
    uid TEXT PRIMARY KEY,
    content_uid TEXT NOT NULL UNIQUE REFERENCES content_objects(uid),
    title TEXT NOT NULL,
    authors TEXT[] NOT NULL DEFAULT '{}',
    series TEXT NOT NULL DEFAULT '',
    series_number TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    format TEXT NOT NULL CHECK (format IN ('pdf', 'epub')),
    original_file_name TEXT NOT NULL,
    page_count INTEGER NOT NULL DEFAULT 0 CHECK (page_count >= 0),
    processing_status TEXT NOT NULL CHECK (processing_status IN ('pending', 'ready', 'failed')),
    create_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    delete_time TIMESTAMPTZ
);

CREATE INDEX books_visible_idx ON books (title, uid) WHERE delete_time IS NULL;
CREATE INDEX books_deleted_idx ON books (delete_time) WHERE delete_time IS NOT NULL;

CREATE TABLE shelves (
    uid TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    create_time TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE shelf_books (
    shelf_uid TEXT NOT NULL REFERENCES shelves(uid) ON DELETE CASCADE,
    book_uid TEXT NOT NULL REFERENCES books(uid) ON DELETE CASCADE,
    create_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (shelf_uid, book_uid)
);

CREATE TABLE reading_progress (
    book_uid TEXT PRIMARY KEY REFERENCES books(uid) ON DELETE CASCADE,
    progress_percent REAL NOT NULL CHECK (progress_percent >= 0 AND progress_percent <= 1),
    location_kind TEXT NOT NULL CHECK (location_kind IN ('pdf', 'epub')),
    pdf_page_number INTEGER NOT NULL DEFAULT 0 CHECK (pdf_page_number >= 0),
    pdf_page_count INTEGER NOT NULL DEFAULT 0 CHECK (pdf_page_count >= 0),
    epub_cfi TEXT NOT NULL DEFAULT '',
    update_time TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE tracks (
    uid TEXT PRIMARY KEY,
    content_uid TEXT NOT NULL UNIQUE REFERENCES content_objects(uid),
    title TEXT NOT NULL,
    artists TEXT[] NOT NULL DEFAULT '{}',
    album TEXT NOT NULL DEFAULT '',
    album_artist TEXT NOT NULL DEFAULT '',
    track_number INTEGER NOT NULL DEFAULT 0 CHECK (track_number >= 0),
    disc_number INTEGER NOT NULL DEFAULT 0 CHECK (disc_number >= 0),
    year INTEGER NOT NULL DEFAULT 0 CHECK (year >= 0),
    duration_ms BIGINT NOT NULL DEFAULT 0 CHECK (duration_ms >= 0),
    original_file_name TEXT NOT NULL,
    favorite BOOLEAN NOT NULL DEFAULT FALSE,
    processing_status TEXT NOT NULL CHECK (processing_status IN ('pending', 'ready', 'failed')),
    create_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    delete_time TIMESTAMPTZ
);

CREATE INDEX tracks_visible_idx ON tracks (title, uid) WHERE delete_time IS NULL;
CREATE INDEX tracks_deleted_idx ON tracks (delete_time) WHERE delete_time IS NOT NULL;

CREATE TABLE playback_history (
    uid TEXT PRIMARY KEY,
    track_uid TEXT NOT NULL REFERENCES tracks(uid) ON DELETE CASCADE,
    play_time TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX playback_history_time_idx ON playback_history (play_time DESC, uid DESC);

CREATE TABLE image_albums (
    uid TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    create_time TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE images (
    uid TEXT PRIMARY KEY,
    content_uid TEXT NOT NULL UNIQUE REFERENCES content_objects(uid),
    album_uid TEXT REFERENCES image_albums(uid) ON DELETE SET NULL,
    display_name TEXT NOT NULL,
    original_file_name TEXT NOT NULL,
    width INTEGER NOT NULL DEFAULT 0 CHECK (width >= 0),
    height INTEGER NOT NULL DEFAULT 0 CHECK (height >= 0),
    processing_status TEXT NOT NULL CHECK (processing_status IN ('pending', 'ready', 'failed')),
    create_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    delete_time TIMESTAMPTZ
);

CREATE INDEX images_visible_idx ON images (display_name, uid) WHERE delete_time IS NULL;
CREATE INDEX images_deleted_idx ON images (delete_time) WHERE delete_time IS NOT NULL;

CREATE TABLE image_links (
    uid TEXT PRIMARY KEY,
    image_uid TEXT NOT NULL REFERENCES images(uid) ON DELETE CASCADE,
    create_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoke_time TIMESTAMPTZ
);

CREATE INDEX image_links_active_idx ON image_links (uid) WHERE revoke_time IS NULL;

CREATE TABLE wallpapers (
    uid TEXT PRIMARY KEY,
    image_uid TEXT NOT NULL UNIQUE REFERENCES images(uid),
    title TEXT NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}',
    dominant_color TEXT NOT NULL DEFAULT '',
    publish_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX wallpapers_publish_idx ON wallpapers (publish_time DESC, uid DESC);
