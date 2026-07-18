import type { PlayableTrack, Track } from "@realtime-me/library-contracts";
import { LOCAL_PROVIDER_ID } from "@realtime-me/library-web";

export function localPlayableTrack(track: Track): PlayableTrack {
  return {
    $typeName: "cloud.music.v1.PlayableTrack",
    providerId: LOCAL_PROVIDER_ID,
    trackId: track.uid,
    title: track.title || track.originalFileName,
    artists: track.artists,
    album: track.album,
    duration: track.duration,
    artworkUrl: track.artworkUrl,
    providerUrl: "",
    playable: true,
    lyricsAvailable: false,
  };
}

export function durationSeconds(track: PlayableTrack): number {
  if (!track.duration) return 0;
  return Number(track.duration.seconds) + track.duration.nanos / 1_000_000_000;
}

export function clock(value: number): string {
  if (!Number.isFinite(value) || value < 0) return "0:00";
  const minutes = Math.floor(value / 60);
  return `${minutes}:${Math.floor(value % 60)
    .toString()
    .padStart(2, "0")}`;
}
