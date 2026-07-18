import type { PlayableTrack } from "@realtime-me/library-contracts";
import type {
  PlaybackMode,
  PlaybackQueuePage,
  PlaybackQueuePageLoader,
  PlaybackQueueSelection,
} from "./playback-types";

export type PlaybackQueueState = Readonly<{
  tracks: PlayableTrack[];
  currentIndex: number;
  nextPageToken: string;
  loadNextPage: PlaybackQueuePageLoader | undefined;
  shuffleHistory: number[];
  shuffleUpcoming: number[];
  playbackSequence: number;
  loadingMore: boolean;
  loadError: string;
}>;

export const EMPTY_PLAYBACK_QUEUE: PlaybackQueueState = {
  tracks: [],
  currentIndex: -1,
  nextPageToken: "",
  loadNextPage: undefined,
  shuffleHistory: [],
  shuffleUpcoming: [],
  playbackSequence: 0,
  loadingMore: false,
  loadError: "",
};

export function selectedQueue(
  selection: PlaybackQueueSelection,
  mode: PlaybackMode,
  playbackSequence: number,
  random: () => number,
): PlaybackQueueState {
  const tracks = [...selection.tracks];
  return {
    tracks,
    currentIndex: selection.startIndex,
    nextPageToken: selection.nextPageToken,
    loadNextPage: selection.loadNextPage,
    shuffleHistory: [],
    shuffleUpcoming:
      mode === "shuffle" ? shuffledIndexes(tracks.length, selection.startIndex, random) : [],
    playbackSequence,
    loadingMore: false,
    loadError: "",
  };
}

export function appendQueuePage(
  current: PlaybackQueueState,
  page: PlaybackQueuePage,
  mode: PlaybackMode,
  random: () => number,
): PlaybackQueueState {
  const firstNewIndex = current.tracks.length;
  const tracks = [...current.tracks, ...page.tracks];
  const newIndexes = Array.from(
    { length: page.tracks.length },
    (_, index) => firstNewIndex + index,
  );
  return {
    ...current,
    tracks,
    nextPageToken: page.nextPageToken,
    loadingMore: false,
    loadError: "",
    shuffleUpcoming:
      mode === "shuffle"
        ? shuffle([...current.shuffleUpcoming, ...newIndexes], random)
        : current.shuffleUpcoming,
  };
}

export function replayQueue(current: PlaybackQueueState): PlaybackQueueState {
  return {
    ...current,
    playbackSequence: current.playbackSequence + 1,
  };
}

export function advanceQueue(
  current: PlaybackQueueState,
  mode: PlaybackMode,
  random: () => number,
): PlaybackQueueState {
  if (current.currentIndex < 0 || current.tracks.length === 0) return current;
  if (mode === "shuffle") return advanceShuffledQueue(current, random);
  if (current.currentIndex < current.tracks.length - 1) {
    return selectIndex(current, current.currentIndex + 1);
  }
  return mode === "repeat-all" ? selectIndex(current, 0) : current;
}

export function retreatQueue(current: PlaybackQueueState, mode: PlaybackMode): PlaybackQueueState {
  if (current.currentIndex < 0) return current;
  if (mode === "shuffle") return retreatShuffledQueue(current);
  if (current.currentIndex > 0) return selectIndex(current, current.currentIndex - 1);
  if (mode === "repeat-all" && current.tracks.length > 1)
    return selectIndex(current, current.tracks.length - 1);
  return replayQueue(current);
}

export function selectQueueIndex(
  current: PlaybackQueueState,
  index: number,
  mode: PlaybackMode,
  random: () => number,
): PlaybackQueueState {
  if (index < 0 || index >= current.tracks.length) return current;
  return {
    ...current,
    currentIndex: index,
    shuffleHistory: [],
    shuffleUpcoming:
      mode === "shuffle" ? shuffledIndexes(current.tracks.length, index, random) : [],
    playbackSequence: current.playbackSequence + 1,
  };
}

export function queueWithMode(
  current: PlaybackQueueState,
  mode: PlaybackMode,
  random: () => number,
): PlaybackQueueState {
  return {
    ...current,
    shuffleHistory: [],
    shuffleUpcoming:
      mode === "shuffle"
        ? shuffledIndexes(current.tracks.length, current.currentIndex, random)
        : [],
  };
}

export function queueNeedsPage(current: PlaybackQueueState, mode: PlaybackMode): boolean {
  if (!current.nextPageToken || !current.loadNextPage) return false;
  return mode === "shuffle"
    ? current.shuffleUpcoming.length === 0
    : current.currentIndex >= current.tracks.length - 1;
}

export function canAdvanceQueue(current: PlaybackQueueState, mode: PlaybackMode): boolean {
  if (current.currentIndex < 0) return false;
  if (current.nextPageToken) return true;
  if (mode === "shuffle") return current.tracks.length > 1;
  if (current.currentIndex < current.tracks.length - 1) return true;
  return mode === "repeat-all" && current.tracks.length > 1;
}

function selectIndex(current: PlaybackQueueState, currentIndex: number): PlaybackQueueState {
  return {
    ...current,
    currentIndex,
    playbackSequence: current.playbackSequence + 1,
  };
}

function advanceShuffledQueue(
  current: PlaybackQueueState,
  random: () => number,
): PlaybackQueueState {
  const upcoming = current.shuffleUpcoming.length
    ? current.shuffleUpcoming
    : shuffledIndexes(current.tracks.length, current.currentIndex, random);
  const nextIndex = upcoming[0];
  if (nextIndex === undefined) return current;
  return {
    ...current,
    currentIndex: nextIndex,
    shuffleHistory: [...current.shuffleHistory, current.currentIndex],
    shuffleUpcoming: upcoming.slice(1),
    playbackSequence: current.playbackSequence + 1,
  };
}

function retreatShuffledQueue(current: PlaybackQueueState): PlaybackQueueState {
  const previousIndex = current.shuffleHistory.at(-1);
  if (previousIndex === undefined) return replayQueue(current);
  return {
    ...current,
    currentIndex: previousIndex,
    shuffleHistory: current.shuffleHistory.slice(0, -1),
    shuffleUpcoming: [current.currentIndex, ...current.shuffleUpcoming],
    playbackSequence: current.playbackSequence + 1,
  };
}

function shuffledIndexes(length: number, excludedIndex: number, random: () => number): number[] {
  const indexes = Array.from({ length }, (_, index) => index).filter(
    (index) => index !== excludedIndex,
  );
  return shuffle(indexes, random);
}

function shuffle(values: number[], random: () => number): number[] {
  const result = [...values];
  for (let index = result.length - 1; index > 0; index -= 1) {
    const target = Math.floor(random() * (index + 1));
    const value = result[index];
    const replacement = result[target];
    if (value === undefined || replacement === undefined) continue;
    result[index] = replacement;
    result[target] = value;
  }
  return result;
}
