ALTER TABLE processing_jobs
    DROP CONSTRAINT processing_jobs_kind_check,
    ADD CONSTRAINT processing_jobs_kind_check
        CHECK (kind IN ('book', 'track', 'image', 'wallpaper', 'music_download', 'music_artwork'));

INSERT INTO processing_jobs (uid, kind, resource_uid, status)
SELECT 'music-artwork-' || track.uid, 'music_artwork', track.uid, 'pending'
FROM tracks track
WHERE track.delete_time IS NULL
  AND track.source_provider IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM content_artifacts artifact
      WHERE artifact.content_uid = track.content_uid
        AND artifact.kind = 'track_artwork'
        AND artifact.variant = 'default'
  )
  AND EXISTS (
      SELECT 1 FROM music_playlist_tracks item
      WHERE item.local_track_uid = track.uid
        AND item.artwork_url <> ''
  )
ON CONFLICT (kind, resource_uid) DO UPDATE
SET status = 'pending',
    attempts = 0,
    available_time = now(),
    lease_until = NULL,
    error_code = '',
    update_time = now();
