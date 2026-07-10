CREATE TABLE drive_items (
    uid TEXT PRIMARY KEY,
    parent_uid TEXT REFERENCES drive_items(uid) DEFERRABLE INITIALLY DEFERRED,
    name TEXT NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('file', 'directory')),
    size_bytes BIGINT NOT NULL DEFAULT 0 CHECK (size_bytes >= 0),
    content_type TEXT NOT NULL DEFAULT '',
    storage_key TEXT NOT NULL DEFAULT '',
    create_time TIMESTAMPTZ NOT NULL,
    update_time TIMESTAMPTZ NOT NULL,
    delete_time TIMESTAMPTZ
);

CREATE INDEX drive_items_parent_visible_idx ON drive_items (parent_uid, name, uid) WHERE delete_time IS NULL;
CREATE INDEX drive_items_deleted_idx ON drive_items (delete_time) WHERE delete_time IS NOT NULL;

CREATE TABLE uploads (
    uid TEXT PRIMARY KEY,
    item_uid TEXT NOT NULL UNIQUE,
    parent_uid TEXT,
    file_name TEXT NOT NULL,
    content_type TEXT NOT NULL DEFAULT '',
    total_size_bytes BIGINT NOT NULL CHECK (total_size_bytes >= 0),
    received_bytes BIGINT NOT NULL DEFAULT 0 CHECK (received_bytes >= 0),
    chunk_size_bytes BIGINT NOT NULL CHECK (chunk_size_bytes > 0),
    status TEXT NOT NULL CHECK (status IN ('active', 'completed', 'expired')),
    create_time TIMESTAMPTZ NOT NULL,
    expire_time TIMESTAMPTZ NOT NULL,
    complete_time TIMESTAMPTZ
);

CREATE INDEX uploads_expire_idx ON uploads (expire_time) WHERE status = 'active';

CREATE TABLE upload_chunks (
    upload_uid TEXT NOT NULL REFERENCES uploads(uid) ON DELETE CASCADE,
    start_offset BIGINT NOT NULL CHECK (start_offset >= 0),
    end_offset BIGINT NOT NULL CHECK (end_offset > start_offset),
    checksum BYTEA NOT NULL,
    PRIMARY KEY (upload_uid, start_offset)
);

CREATE INDEX upload_chunks_range_idx ON upload_chunks (upload_uid, start_offset);

CREATE TABLE share_links (
    uid TEXT PRIMARY KEY,
    target_uid TEXT NOT NULL REFERENCES drive_items(uid),
    token_hash BYTEA NOT NULL UNIQUE,
    create_time TIMESTAMPTZ NOT NULL,
    expire_time TIMESTAMPTZ NOT NULL,
    revoke_time TIMESTAMPTZ
);

CREATE INDEX share_links_token_active_idx ON share_links (token_hash) WHERE revoke_time IS NULL;
