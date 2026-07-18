import type { PlaybackDescriptor, PlayableTrack } from "@realtime-me/library-contracts";
import { durationSeconds } from "../music-model";
import type { PlaybackAdapterState } from "./playback-types";

export type PlaybackSessionState = Readonly<{
  descriptor: PlaybackDescriptor | undefined;
  paused: boolean;
  position: number;
  duration: number;
  loading: boolean;
  error: string;
}>;

export function emptyPlaybackState(
  track: PlayableTrack | undefined,
): PlaybackSessionState {
  return {
    descriptor: undefined,
    paused: true,
    position: 0,
    duration: track ? durationSeconds(track) : 0,
    loading: Boolean(track),
    error: "",
  };
}

export function mergePlaybackState(
  current: PlaybackSessionState,
  next: PlaybackAdapterState,
): PlaybackSessionState {
  return {
    ...current,
    paused: next.paused,
    position: next.position,
    duration: next.duration || current.duration,
    loading: false,
    error: "",
  };
}

export function effectiveVolume(volume: number, muted: boolean): number {
  return muted ? 0 : volume;
}

export function playbackErrorMessage(error: unknown): string {
  if (error instanceof DOMException && error.name === "AbortError")
    return "播放请求超时";
  return error instanceof Error ? error.message : "播放失败";
}
