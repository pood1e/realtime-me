import { useCallback, useEffect, useRef, useState } from "react";
import { PLAYBACK_MODES, type PlaybackMode } from "./playback-types";
import type {
  PlaybackQueuePage,
  PlaybackQueueSelection,
} from "./playback-types";
import {
  advanceQueue,
  appendQueuePage,
  canAdvanceQueue,
  EMPTY_PLAYBACK_QUEUE,
  queueNeedsPage,
  queueWithMode,
  replayQueue,
  retreatQueue,
  selectedQueue,
  selectQueueIndex,
  type PlaybackQueueState,
} from "./playback-queue-state";

const PREVIOUS_RESTART_THRESHOLD_SECONDS = 3;
const MAX_EMPTY_PAGE_FETCHES = 5;

export function usePlaybackQueue({
  initialMode,
  onModeChange,
  random,
}: {
  initialMode: PlaybackMode;
  onModeChange: (mode: PlaybackMode) => void;
  random: () => number;
}) {
  const [mode, setMode] = useState(initialMode);
  const modeRef = useRef(mode);
  const [state, setRenderedState] =
    useState<PlaybackQueueState>(EMPTY_PLAYBACK_QUEUE);
  const stateRef = useRef(state);
  const generation = useRef(0);
  const loadRequest = useRef<AbortController | undefined>(undefined);
  modeRef.current = mode;

  const commit = useCallback(
    (update: (current: PlaybackQueueState) => PlaybackQueueState) => {
      const next = update(stateRef.current);
      stateRef.current = next;
      setRenderedState(next);
      return next;
    },
    [],
  );

  useEffect(() => () => loadRequest.current?.abort(), []);

  const start = useCallback(
    (selection: PlaybackQueueSelection) => {
      if (!validSelection(selection)) return;
      generation.current += 1;
      loadRequest.current?.abort();
      loadRequest.current = undefined;
      const next = selectedQueue(
        selection,
        mode,
        stateRef.current.playbackSequence + 1,
        random,
      );
      stateRef.current = next;
      setRenderedState(next);
    },
    [mode, random],
  );

  const loadMore = useCallback(async (): Promise<boolean> => {
    const initial = stateRef.current;
    if (initial.loadingMore || !initial.nextPageToken || !initial.loadNextPage)
      return false;

    const requestGeneration = generation.current;
    const controller = new AbortController();
    loadRequest.current?.abort();
    loadRequest.current = controller;
    commit((current) => ({ ...current, loadingMore: true, loadError: "" }));

    try {
      const page = await loadPlayablePage(
        initial.nextPageToken,
        initial.loadNextPage,
        controller.signal,
      );
      if (!requestIsCurrent(controller, requestGeneration, generation.current))
        return false;
      const previousLength = stateRef.current.tracks.length;
      commit((current) =>
        appendQueuePage(current, page, modeRef.current, random),
      );
      return stateRef.current.tracks.length > previousLength;
    } catch (error) {
      if (!requestIsCurrent(controller, requestGeneration, generation.current))
        return false;
      commit((current) => ({
        ...current,
        loadingMore: false,
        loadError: message(error),
      }));
      return false;
    } finally {
      if (loadRequest.current === controller) loadRequest.current = undefined;
    }
  }, [commit, random]);

  const next = useCallback(
    async (reason: "manual" | "ended" = "manual") => {
      let current = stateRef.current;
      if (current.currentIndex < 0 || current.loadingMore) return;
      if (reason === "ended" && mode === "repeat-one") {
        commit(replayQueue);
        return;
      }
      if (queueNeedsPage(current, mode)) {
        await loadMore();
        current = stateRef.current;
        if (current.loadError) return;
      }
      commit((queue) => advanceQueue(queue, mode, random));
    },
    [commit, loadMore, mode, random],
  );

  const previous = useCallback(
    (position: number) => {
      commit((current) =>
        position > PREVIOUS_RESTART_THRESHOLD_SECONDS
          ? replayQueue(current)
          : retreatQueue(current, mode),
      );
    },
    [commit, mode],
  );

  const playIndex = useCallback(
    (index: number) => {
      commit((current) => selectQueueIndex(current, index, mode, random));
    },
    [commit, mode, random],
  );

  const cycleMode = useCallback(() => {
    const nextMode = nextPlaybackMode(mode);
    setMode(nextMode);
    onModeChange(nextMode);
    commit((current) => queueWithMode(current, nextMode, random));
  }, [commit, mode, onModeChange, random]);

  return {
    ...state,
    currentTrack: state.tracks[state.currentIndex],
    mode,
    canNext: canAdvanceQueue(state, mode),
    canPrevious: state.currentIndex >= 0,
    hasMore: Boolean(state.nextPageToken),
    start,
    next,
    previous,
    playIndex,
    cycleMode,
    loadMore,
  };
}

export type PlaybackQueueController = ReturnType<typeof usePlaybackQueue>;

async function loadPlayablePage(
  firstPageToken: string,
  loadPage: NonNullable<PlaybackQueueState["loadNextPage"]>,
  signal: AbortSignal,
): Promise<PlaybackQueuePage> {
  let pageToken = firstPageToken;
  let page: PlaybackQueuePage = { tracks: [], nextPageToken: pageToken };
  for (
    let attempt = 0;
    attempt < MAX_EMPTY_PAGE_FETCHES && pageToken;
    attempt += 1
  ) {
    page = await loadPage(pageToken, signal);
    pageToken = page.nextPageToken;
    if (page.tracks.length) break;
  }
  return page;
}

function validSelection(selection: PlaybackQueueSelection): boolean {
  return (
    selection.startIndex >= 0 && selection.startIndex < selection.tracks.length
  );
}

function requestIsCurrent(
  controller: AbortController,
  requestGeneration: number,
  currentGeneration: number,
): boolean {
  return !controller.signal.aborted && requestGeneration === currentGeneration;
}

function nextPlaybackMode(mode: PlaybackMode): PlaybackMode {
  const index = PLAYBACK_MODES.indexOf(mode);
  return PLAYBACK_MODES[(index + 1) % PLAYBACK_MODES.length] ?? "sequential";
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : "播放队列加载失败";
}
