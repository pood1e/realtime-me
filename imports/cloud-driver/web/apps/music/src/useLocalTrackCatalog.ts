import { useCallback, useDeferredValue, useEffect, useMemo } from "react";
import { ProcessingStatus, type Track } from "@cloud-drive/contracts";
import type {
  CursorPage,
  InfiniteData,
  MusicClient,
  TrackListOptions,
} from "@cloud-drive/shared";
import { useCursorQuery, useQueryClient } from "@cloud-drive/shared";

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
      tracks.some(
        (track) => track.processingStatus === ProcessingStatus.PENDING,
      ),
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
      queryClient.setQueryData<InfiniteData<CursorPage<Track>, string>>(
        queryKey,
        (current) =>
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
        tracks.map((candidate) =>
          candidate.uid === track.uid ? track : candidate,
        ),
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

  return {
    tracks: catalog.items,
    initialLoading: catalog.isPending,
    loadingMore: catalog.isFetchingNextPage,
    loadMoreFailed: catalog.isFetchNextPageError,
    hasMore: catalog.hasNextPage,
    loadMore,
    refresh,
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
