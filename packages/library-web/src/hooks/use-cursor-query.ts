import type { InfiniteData, QueryKey } from "@tanstack/react-query";
import { useInfiniteQuery } from "@tanstack/react-query";

export type CursorPage<T> = Readonly<{
  items: T[];
  nextPageToken: string;
}>;

export function useCursorQuery<T>({
  queryKey,
  loadPage,
  enabled = true,
  pollInterval,
  shouldPoll,
}: {
  queryKey: QueryKey;
  loadPage: (pageToken: string, signal: AbortSignal) => Promise<CursorPage<T>>;
  enabled?: boolean;
  pollInterval?: number;
  shouldPoll?: (items: T[]) => boolean;
}) {
  const query = useInfiniteQuery<
    CursorPage<T>,
    Error,
    InfiniteData<CursorPage<T>, string>,
    QueryKey,
    string
  >({
    queryKey,
    enabled,
    refetchInterval:
      pollInterval && shouldPoll
        ? (current) => {
            const items = current.state.data?.pages.flatMap((page) => page.items) ?? [];
            return shouldPoll(items) ? pollInterval : false;
          }
        : false,
    initialPageParam: "",
    queryFn: ({ pageParam, signal }) => loadPage(pageParam, signal),
    getNextPageParam: (page) => page.nextPageToken || undefined,
  });
  return {
    ...query,
    items: query.data?.pages.flatMap((page) => page.items) ?? [],
  };
}
