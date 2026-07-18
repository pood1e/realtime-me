CREATE TABLE content_tombstones (
    content_uid TEXT PRIMARY KEY,
    storage_keys TEXT[] NOT NULL CHECK (cardinality(storage_keys) > 0),
    delete_after TIMESTAMPTZ NOT NULL,
    create_time TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX content_tombstones_delete_idx
    ON content_tombstones (delete_after, content_uid);

CREATE INDEX uploads_claim_retention_idx
    ON uploads (claim_time, uid)
    WHERE status = 'claimed';
