import { useCallback, useEffect, useRef, useState } from "react";
import type { Track } from "@cloud-drive/contracts";
import type { MusicClient, TrackListOptions } from "@cloud-drive/shared";

const TRACK_PAGE_SIZE = 50;

export type LocalLibraryMode = "all" | "favorites" | "trash";

type CatalogParameters = Readonly<{
  client: MusicClient;
  query: string;
  mode: LocalLibraryMode;
  onError: (error: unknown) => void;
}>;

export function useLocalTrackCatalog({
  client,
  query,
  mode,
  onError,
}: CatalogParameters) {
  const [tracks, setTracks] = useState<Track[]>([]);
  const [nextPageToken, setNextPageToken] = useState("");
  const [initialLoading, setInitialLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [loadMoreFailed, setLoadMoreFailed] = useState(false);
  const requestVersion = useRef(0);
  const firstPageRequest = useRef<AbortController | undefined>(undefined);
  const nextPageRequest = useRef<AbortController | undefined>(undefined);

  const cancelRequests = useCallback(() => {
    firstPageRequest.current?.abort();
    nextPageRequest.current?.abort();
    firstPageRequest.current = undefined;
    nextPageRequest.current = undefined;
  }, []);

  const loadFirstPage = useCallback(async () => {
    cancelRequests();
    const controller = new AbortController();
    const version = ++requestVersion.current;
    firstPageRequest.current = controller;
    setTracks([]);
    setNextPageToken("");
    setInitialLoading(true);
    setLoadingMore(false);
    setLoadMoreFailed(false);
    try {
      const page = await client.trackPage(
        listOptions(query, mode),
        controller.signal,
      );
      if (controller.signal.aborted || requestVersion.current !== version)
        return;
      setTracks(page.tracks);
      setNextPageToken(page.nextPageToken);
    } catch (error) {
      if (!controller.signal.aborted) onError(error);
    } finally {
      if (firstPageRequest.current === controller)
        firstPageRequest.current = undefined;
      if (requestVersion.current === version) setInitialLoading(false);
    }
  }, [cancelRequests, client, mode, onError, query]);

  const loadMore = useCallback(async () => {
    if (initialLoading || nextPageRequest.current || !nextPageToken) return;
    const controller = new AbortController();
    const version = requestVersion.current;
    nextPageRequest.current = controller;
    setLoadingMore(true);
    setLoadMoreFailed(false);
    try {
      const page = await client.trackPage(
        listOptions(query, mode, nextPageToken),
        controller.signal,
      );
      if (controller.signal.aborted || requestVersion.current !== version)
        return;
      setTracks((current) => appendUnique(current, page.tracks));
      setNextPageToken(page.nextPageToken);
    } catch (error) {
      if (!controller.signal.aborted) {
        setLoadMoreFailed(true);
        onError(error);
      }
    } finally {
      if (nextPageRequest.current === controller)
        nextPageRequest.current = undefined;
      if (requestVersion.current === version) setLoadingMore(false);
    }
  }, [client, initialLoading, mode, nextPageToken, onError, query]);

  useEffect(() => {
    const timer = window.setTimeout(() => void loadFirstPage(), 180);
    return () => {
      window.clearTimeout(timer);
      requestVersion.current++;
      cancelRequests();
    };
  }, [cancelRequests, loadFirstPage]);

  const updateTrack = useCallback((track: Track) => {
    setTracks((current) =>
      current.map((candidate) =>
        candidate.uid === track.uid ? track : candidate,
      ),
    );
  }, []);

  const removeTrack = useCallback((trackUid: string) => {
    setTracks((current) => current.filter((track) => track.uid !== trackUid));
  }, []);

  return {
    tracks,
    initialLoading,
    loadingMore,
    loadMoreFailed,
    hasMore: nextPageToken !== "",
    loadMore,
    refresh: loadFirstPage,
    updateTrack,
    removeTrack,
  };
}

function listOptions(
  query: string,
  mode: LocalLibraryMode,
  pageToken = "",
): TrackListOptions {
  return {
    query,
    favorites: mode === "favorites",
    trashed: mode === "trash",
    pageSize: TRACK_PAGE_SIZE,
    pageToken,
  };
}

function appendUnique(current: Track[], next: Track[]): Track[] {
  const known = new Set(current.map((track) => track.uid));
  return [...current, ...next.filter((track) => !known.has(track.uid))];
}
