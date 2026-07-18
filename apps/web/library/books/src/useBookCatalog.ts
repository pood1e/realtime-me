import { useCallback, useDeferredValue, useEffect } from "react";
import { BookFormat, ProcessingStatus } from "@realtime-me/library-contracts";
import type { Book } from "@realtime-me/library-contracts";
import type { BookListOptions, BooksClient } from "@realtime-me/library-web";
import { useCursorQuery, useQuery } from "@realtime-me/library-web";

const BOOK_PAGE_SIZE = 32;

export type BookFilter = "all" | "pdf" | "epub" | "trash";

type CatalogParameters = Readonly<{
  client: BooksClient;
  query: string;
  filter: BookFilter;
  shelfUid: string;
  onError: (error: unknown) => void;
}>;

export function useBookCatalog({
  client,
  query,
  filter,
  shelfUid,
  onError,
}: CatalogParameters) {
  const deferredQuery = useDeferredValue(query.trim());
  const catalog = useCursorQuery<Book>({
    queryKey: ["books", deferredQuery, filter, shelfUid],
    pollInterval: 2_500,
    shouldPoll: (books) =>
      books.some((book) => book.processingStatus === ProcessingStatus.PENDING),
    loadPage: async (pageToken, signal) => {
      const page = await client.listPage(
        listOptions(deferredQuery, filter, shelfUid, pageToken),
        signal,
      );
      return { items: page.books, nextPageToken: page.nextPageToken };
    },
  });
  const shelves = useQuery({
    queryKey: ["book-shelves"],
    queryFn: ({ signal }) => client.shelves(signal),
  });

  useEffect(() => {
    if (catalog.error) onError(catalog.error);
  }, [catalog.error, onError]);
  useEffect(() => {
    if (shelves.error) onError(shelves.error);
  }, [onError, shelves.error]);

  const loadMore = useCallback(async () => {
    await catalog.fetchNextPage();
  }, [catalog.fetchNextPage]);
  const refresh = useCallback(async () => {
    await Promise.all([catalog.refetch(), shelves.refetch()]);
  }, [catalog.refetch, shelves.refetch]);

  return {
    books: catalog.items,
    shelves: shelves.data ?? [],
    initialLoading: catalog.isPending,
    loadingMore: catalog.isFetchingNextPage,
    loadMoreFailed: catalog.isFetchNextPageError,
    hasMore: catalog.hasNextPage,
    loadMore,
    refresh,
  };
}

function listOptions(
  query: string,
  filter: BookFilter,
  shelfUid: string,
  pageToken = "",
): BookListOptions {
  const format =
    filter === "pdf"
      ? BookFormat.PDF
      : filter === "epub"
        ? BookFormat.EPUB
        : undefined;
  return {
    query,
    ...(format === undefined ? {} : { format }),
    ...(shelfUid === "all" ? {} : { shelfUid }),
    trashed: filter === "trash",
    pageSize: BOOK_PAGE_SIZE,
    pageToken,
  };
}
