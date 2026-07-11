import {
  MusicProvider,
  type PlayableTrack,
  type Track,
} from "@cloud-drive/contracts";

export function localPlayableTrack(track: Track): PlayableTrack {
  return {
    $typeName: "cloud.music.v1.PlayableTrack",
    provider: MusicProvider.LOCAL,
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

export function providerLabel(provider: MusicProvider): string {
  switch (provider) {
    case MusicProvider.LOCAL:
      return "本地音乐";
    case MusicProvider.QQ_MUSIC:
      return "QQ 音乐";
    case MusicProvider.NETEASE_CLOUD_MUSIC:
      return "网易云音乐";
    case MusicProvider.SPOTIFY:
      return "Spotify";
    default:
      return "未知来源";
  }
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
