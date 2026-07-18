import type { PlayableTrack } from "@realtime-me/library-contracts";

export const PLAYBACK_MODES = [
  "sequential",
  "repeat-all",
  "repeat-one",
  "shuffle",
] as const;

export type PlaybackMode = (typeof PLAYBACK_MODES)[number];

export type PlaybackQueuePage = Readonly<{
  tracks: PlayableTrack[];
  nextPageToken: string;
}>;

export type PlaybackQueuePageLoader = (
  pageToken: string,
  signal: AbortSignal,
) => Promise<PlaybackQueuePage>;

export type PlaybackQueueSelection = Readonly<{
  tracks: readonly PlayableTrack[];
  startIndex: number;
  nextPageToken: string;
  loadNextPage: PlaybackQueuePageLoader;
}>;

export type PlaybackAdapterState = Readonly<{
  paused: boolean;
  position: number;
  duration: number;
}>;

export interface PlaybackAdapter {
  load(resource: string): Promise<void>;
  play(): Promise<void>;
  pause(): Promise<void>;
  seek(seconds: number): Promise<void>;
  setVolume(volume: number): Promise<void>;
  destroy(): void;
}

export type PlaybackAdapterEvents = Readonly<{
  onState: (state: PlaybackAdapterState) => void;
  onEnded: () => void;
  onError: (error: unknown) => void;
}>;
