-- Older releases classified a track-level QQ Music playback denial as an
-- expired account credential. Let the corrected provider mapping verify the
-- saved credentials again on the next request.
UPDATE music_provider_connections
SET status = 'connected',
    update_time = now()
WHERE provider = 'qq_music'
  AND status = 'reconnect_required';
