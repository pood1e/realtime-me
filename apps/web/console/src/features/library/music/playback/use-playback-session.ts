import { type PlayableTrack, PlaybackQuality } from "@realtime-me/library-contracts";
import { LOCAL_PROVIDER_ID, type MusicClient } from "@realtime-me/library-web";
import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import { createPlaybackAdapter, playbackResource } from "./playback-adapter-factory";
import {
  effectiveVolume,
  emptyPlaybackState,
  mergePlaybackState,
  type PlaybackSessionState,
  playbackErrorMessage,
} from "./playback-session-state";
import type { PlaybackAdapter, PlaybackAdapterEvents } from "./playback-types";

export function usePlaybackSession({
  track,
  playbackSequence,
  client,
  volume,
  muted,
  onEnded,
  onRecorded,
}: {
  track: PlayableTrack | undefined;
  playbackSequence: number;
  client: MusicClient;
  volume: number;
  muted: boolean;
  onEnded: () => void;
  onRecorded: () => void;
}) {
  const [state, setState] = useState<PlaybackSessionState>(() => emptyPlaybackState(track));
  const adapter = useRef<PlaybackAdapter | undefined>(undefined);
  const resolution = useRef<AbortController | undefined>(undefined);
  const generation = useRef(0);
  const recorded = useRef(false);
  const volumeRef = useRef(effectiveVolume(volume, muted));
  const onEndedRef = useRef(onEnded);
  const onRecordedRef = useRef(onRecorded);
  volumeRef.current = effectiveVolume(volume, muted);
  onEndedRef.current = onEnded;
  onRecordedRef.current = onRecorded;

  useEffect(() => {
    generation.current += 1;
    const currentGeneration = generation.current;
    recorded.current = false;
    resolution.current?.abort();
    adapter.current?.destroy();
    adapter.current = undefined;
    setState(emptyPlaybackState(track));
    if (!track) return;

    const controller = new AbortController();
    resolution.current = controller;
    let active = true;

    const recordPlayback = () => {
      if (recorded.current) return;
      recorded.current = true;
      void client.library
        .recordPlayback(track)
        .then(() => onRecordedRef.current())
        .catch(() => {
          if (generation.current === currentGeneration) recorded.current = false;
        });
    };

    const start = async (quality: PlaybackQuality, allowFallback: boolean): Promise<void> => {
      try {
        const descriptor = await client.providers.resolvePlayback(
          track,
          quality,
          controller.signal,
        );
        if (!active || controller.signal.aborted) return;

        let currentAdapter: PlaybackAdapter | undefined;
        let failureHandled = false;
        const handleFailure = (error: unknown) => {
          if (failureHandled || !active || !currentAdapter || adapter.current !== currentAdapter)
            return;
          failureHandled = true;
          if (
            allowFallback &&
            descriptor.providerId !== LOCAL_PROVIDER_ID &&
            descriptor.playback.case === "directAudio"
          ) {
            currentAdapter.destroy();
            adapter.current = undefined;
            void start(PlaybackQuality.HIGH, false);
            return;
          }
          updateError(setState, error);
        };
        const events: PlaybackAdapterEvents = {
          onState: (next) => {
            if (!active || adapter.current !== currentAdapter) return;
            setState((current) => mergePlaybackState(current, next));
            if (!next.paused) recordPlayback();
          },
          onEnded: () => {
            if (active && adapter.current === currentAdapter) onEndedRef.current();
          },
          onError: handleFailure,
        };
        currentAdapter = createPlaybackAdapter(descriptor, client, events);
        if (!currentAdapter) throw new Error("当前浏览器不支持此平台播放器");

        adapter.current?.destroy();
        adapter.current = currentAdapter;
        setState((current) => ({
          ...current,
          descriptor,
          loading: true,
          error: "",
        }));
        try {
          await currentAdapter.setVolume(volumeRef.current);
          await currentAdapter.load(playbackResource(descriptor, client));
        } catch (error) {
          handleFailure(error);
          return;
        }
        if (active && adapter.current === currentAdapter)
          setState((current) => ({ ...current, loading: false }));
      } catch (error) {
        if (active && !controller.signal.aborted) updateError(setState, error);
      }
    };

    void start(PlaybackQuality.BEST_COMPATIBLE, true);
    return () => {
      active = false;
      controller.abort();
      if (resolution.current === controller) resolution.current = undefined;
      adapter.current?.destroy();
      adapter.current = undefined;
    };
  }, [client, playbackSequence, track]);

  useEffect(() => {
    const current = adapter.current;
    if (!current) return;
    void current.setVolume(effectiveVolume(volume, muted)).catch((error) => {
      if (adapter.current === current) updateError(setState, error);
    });
  }, [muted, volume]);

  const play = useCallback(() => {
    const current = adapter.current;
    if (!current) return;
    const operation = async () => {
      if (state.duration > 0 && state.position >= state.duration - 0.25) {
        await current.seek(0);
        setState((value) => ({ ...value, position: 0 }));
      }
      await current.play();
    };
    void operation().catch((error) => {
      if (adapter.current === current) updateError(setState, error);
    });
  }, [state.duration, state.position]);

  const pause = useCallback(() => {
    const current = adapter.current;
    if (!current) return;
    void current.pause().catch((error) => {
      if (adapter.current === current) updateError(setState, error);
    });
  }, []);

  const toggle = useCallback(() => {
    if (state.paused) play();
    else pause();
  }, [pause, play, state.paused]);

  const seek = useCallback(
    (seconds: number) => {
      const current = adapter.current;
      if (!current) return;
      const finite = Number.isFinite(seconds) ? Math.max(0, seconds) : 0;
      const position = state.duration > 0 ? Math.min(finite, state.duration) : finite;
      setState((value) => ({ ...value, position }));
      void current.seek(position).catch((error) => {
        if (adapter.current === current) updateError(setState, error);
      });
    },
    [state.duration],
  );

  return {
    ...state,
    playing: !state.paused,
    ready: Boolean(adapter.current) && !state.loading,
    play,
    pause,
    toggle,
    seek,
  };
}

export type PlaybackSessionController = ReturnType<typeof usePlaybackSession>;

function updateError(
  setState: Dispatch<SetStateAction<PlaybackSessionState>>,
  error: unknown,
): void {
  setState((current) => ({
    ...current,
    paused: true,
    loading: false,
    error: playbackErrorMessage(error),
  }));
}
