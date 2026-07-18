import type { PlayableTrack } from "@realtime-me/library-contracts";
import { useEffect, useRef } from "react";

export function useMediaSession({
  track,
  artwork,
  playing,
  position,
  duration,
  canNext,
  onPlay,
  onPause,
  onPrevious,
  onNext,
  onSeek,
}: {
  track: PlayableTrack | undefined;
  artwork: string;
  playing: boolean;
  position: number;
  duration: number;
  canNext: boolean;
  onPlay: () => void;
  onPause: () => void;
  onPrevious: () => void;
  onNext: () => void;
  onSeek: (seconds: number) => void;
}) {
  const handlers = useRef({
    position,
    onPlay,
    onPause,
    onPrevious,
    onNext,
    onSeek,
  });
  handlers.current = {
    position,
    onPlay,
    onPause,
    onPrevious,
    onNext,
    onSeek,
  };

  useEffect(() => {
    if (!("mediaSession" in navigator) || !track) return;
    if (typeof MediaMetadata !== "undefined") {
      navigator.mediaSession.metadata = new MediaMetadata({
        title: track.title,
        artist: track.artists.join("、") || "未知艺人",
        album: track.album,
        artwork: artwork ? [{ src: artwork }] : [],
      });
    }
    return () => {
      navigator.mediaSession.metadata = null;
      navigator.mediaSession.playbackState = "none";
    };
  }, [artwork, track]);

  useEffect(() => {
    if (!("mediaSession" in navigator) || !track) return;
    setAction("play", () => handlers.current.onPlay());
    setAction("pause", () => handlers.current.onPause());
    setAction("previoustrack", () => handlers.current.onPrevious());
    setAction("nexttrack", canNext ? () => handlers.current.onNext() : undefined);
    setSeekAction("seekto", (seconds) => handlers.current.onSeek(seconds));
    setRelativeSeekAction("seekbackward", -1, handlers);
    setRelativeSeekAction("seekforward", 1, handlers);
    return clearMediaSessionActions;
  }, [canNext, track]);

  useEffect(() => {
    if (!("mediaSession" in navigator) || !track) return;
    navigator.mediaSession.playbackState = playing ? "playing" : "paused";
  }, [playing, track]);

  useEffect(() => {
    if (!("mediaSession" in navigator) || !track || !Number.isFinite(duration) || duration <= 0)
      return;
    try {
      navigator.mediaSession.setPositionState({
        duration,
        playbackRate: 1,
        position: Math.min(duration, Math.max(0, position)),
      });
    } catch {
      // Some browsers expose Media Session without position-state support.
    }
  }, [duration, position, track]);
}

function clearMediaSessionActions(): void {
  clearAction("play");
  clearAction("pause");
  clearAction("previoustrack");
  clearAction("nexttrack");
  clearAction("seekto");
  clearAction("seekbackward");
  clearAction("seekforward");
}

function setAction(action: MediaSessionAction, handler: (() => void) | undefined): void {
  try {
    navigator.mediaSession.setActionHandler(action, handler ?? null);
  } catch {
    // Unsupported actions are intentionally omitted from system controls.
  }
}

function setSeekAction(action: MediaSessionAction, onSeek: (seconds: number) => void): void {
  try {
    navigator.mediaSession.setActionHandler(action, (details) => {
      if ("seekTime" in details && details.seekTime !== undefined) onSeek(details.seekTime);
    });
  } catch {
    // Unsupported actions are intentionally omitted from system controls.
  }
}

function clearAction(action: MediaSessionAction): void {
  try {
    navigator.mediaSession.setActionHandler(action, null);
  } catch {
    // The corresponding action was not supported by this browser.
  }
}

function setRelativeSeekAction(
  action: MediaSessionAction,
  direction: -1 | 1,
  handlers: Readonly<{
    current: Readonly<{
      position: number;
      onSeek: (seconds: number) => void;
    }>;
  }>,
): void {
  try {
    navigator.mediaSession.setActionHandler(action, (details) => {
      const offset = "seekOffset" in details ? (details.seekOffset ?? 5) : 5;
      handlers.current.onSeek(Math.max(0, handlers.current.position + direction * offset));
    });
  } catch {
    // Unsupported actions are intentionally omitted from system controls.
  }
}
