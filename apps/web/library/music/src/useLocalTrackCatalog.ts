import { ProcessingStatus, type Track } from "@realtime-me/library-contracts";
import type {
  CursorPage,
  InfiniteData,
  MusicClient,
  TrackListOptions,
} from "@realtime-me/library-web";
import { useCursorQuery, useQueryClient } from "@realtime-me/library-web";
import { useCallback, useDeferredValue, useEffect, useMemo } from "react";
import { localPlayableTrack } from "./music-model";

const TRACK_PAGE_SIZE = 50;

export type LocalLibraryMode = "all" | "favorites" | "trash";

type CatalogParameters = Readonly<{
  client: MusicClient;
  query: string;
  mode: LocalLibraryMode;
  onError: (error: unknown) => void;
}>;

export function useLocalTrackCatalog({ client, query, mode, onError }: CatalogParameters) {
  const deferredQuery = useDeferredValue(query.trim());
  const queryKey = useMemo(
    () => ["music-tracks", deferredQuery, mode] as const,
    [deferredQuery, mode],
  );
  const queryClient = useQueryClient();
  const catalog = useCursorQuery<Track>({
    queryKey,
    pollInterval: 2_500,
    shouldPoll: (tracks) =>
      tracks.some((track) => track.processingStatus === ProcessingStatus.PENDING),
    loadPage: async (pageToken, signal) => {
      const page = await client.library.trackPage(
        listOptions(deferredQuery, mode, pageToken),
        signal,
      );
      return { items: page.tracks, nextPageToken: page.nextPageToken };
    },
  });

  useEffect(() => {
    if (catalog.error) onError(catalog.error);
  }, [catalog.error, onError]);

  const updatePages = useCallback(
    (update: (tracks: Track[]) => Track[]) => {
      queryClient.setQueryData<InfiniteData<CursorPage<Track>, string>>(queryKey, (current) =>
        current
          ? {
              ...current,
              pages: current.pages.map((page) => ({
                ...page,
                items: update(page.items),
              })),
            }
          : current,
      );
    },
    [queryClient, queryKey],
  );
  const updateTrack = useCallback(
    (track: Track) => {
      updatePages((tracks) =>
        tracks.map((candidate) => (candidate.uid === track.uid ? track : candidate)),
      );
    },
    [updatePages],
  );
  const removeTrack = useCallback(
    (trackUid: string) => {
      updatePages((tracks) => tracks.filter((track) => track.uid !== trackUid));
    },
    [updatePages],
  );
  const loadMore = useCallback(async () => {
    await catalog.fetchNextPage();
  }, [catalog.fetchNextPage]);
  const refresh = useCallback(async () => {
    await catalog.refetch();
  }, [catalog.refetch]);
  const loadPlaybackPage = useCallback(
    async (pageToken: string, signal: AbortSignal) => {
      const page = await client.library.trackPage(
        listOptions(deferredQuery, mode, pageToken),
        signal,
      );
      return {
        tracks: page.tracks.map(localPlayableTrack),
        nextPageToken: page.nextPageToken,
      };
    },
    [client, deferredQuery, mode],
  );

  return {
    tracks: catalog.items,
    initialLoading: catalog.isPending,
    loadingMore: catalog.isFetchingNextPage,
    loadMoreFailed: catalog.isFetchNextPageError,
    hasMore: catalog.hasNextPage,
    loadMore,
    nextPageToken: catalog.data?.pages.at(-1)?.nextPageToken ?? "",
    loadPlaybackPage,
    refresh,
    updateTrack,
    removeTrack,
  };
}

function listOptions(query: string, mode: LocalLibraryMode, pageToken = ""): TrackListOptions {
  return {
    query,
    favorites: mode === "favorites",
    trashed: mode === "trash",
    pageSize: TRACK_PAGE_SIZE,
    pageToken,
  };
}
