import { useCallback, useEffect, useRef, useState } from "react";
import { BookFormat } from "@cloud-drive/contracts";
import type { Book, Shelf } from "@cloud-drive/contracts";
import type { BookListOptions, BooksClient } from "@cloud-drive/shared";

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
  const [books, setBooks] = useState<Book[]>([]);
  const [shelves, setShelves] = useState<Shelf[]>([]);
  const [nextPageToken, setNextPageToken] = useState("");
  const [initialLoading, setInitialLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [loadMoreFailed, setLoadMoreFailed] = useState(false);
  const requestVersion = useRef(0);
  const firstPageRequest = useRef<AbortController | undefined>(undefined);
  const nextPageRequest = useRef<AbortController | undefined>(undefined);
  const shelvesRequest = useRef<AbortController | undefined>(undefined);

  const cancelCatalogRequests = useCallback(() => {
    firstPageRequest.current?.abort();
    nextPageRequest.current?.abort();
    firstPageRequest.current = undefined;
    nextPageRequest.current = undefined;
  }, []);

  const loadFirstPage = useCallback(async () => {
    cancelCatalogRequests();
    const controller = new AbortController();
    const version = ++requestVersion.current;
    firstPageRequest.current = controller;
    setBooks([]);
    setNextPageToken("");
    setInitialLoading(true);
    setLoadingMore(false);
    setLoadMoreFailed(false);
    try {
      const page = await client.listPage(
        listOptions(query, filter, shelfUid),
        controller.signal,
      );
      if (controller.signal.aborted || requestVersion.current !== version)
        return;
      setBooks(page.books);
      setNextPageToken(page.nextPageToken);
    } catch (error) {
      if (!controller.signal.aborted) onError(error);
    } finally {
      if (firstPageRequest.current === controller)
        firstPageRequest.current = undefined;
      if (requestVersion.current === version) setInitialLoading(false);
    }
  }, [cancelCatalogRequests, client, filter, onError, query, shelfUid]);

  const loadMore = useCallback(async () => {
    if (initialLoading || nextPageRequest.current || !nextPageToken) return;
    const controller = new AbortController();
    const version = requestVersion.current;
    nextPageRequest.current = controller;
    setLoadingMore(true);
    setLoadMoreFailed(false);
    try {
      const page = await client.listPage(
        listOptions(query, filter, shelfUid, nextPageToken),
        controller.signal,
      );
      if (controller.signal.aborted || requestVersion.current !== version)
        return;
      setBooks((current) => [...current, ...page.books]);
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
  }, [client, filter, initialLoading, nextPageToken, onError, query, shelfUid]);

  const refreshShelves = useCallback(async () => {
    shelvesRequest.current?.abort();
    const controller = new AbortController();
    shelvesRequest.current = controller;
    try {
      const nextShelves = await client.shelves(controller.signal);
      if (!controller.signal.aborted) setShelves(nextShelves);
    } catch (error) {
      if (!controller.signal.aborted) onError(error);
    } finally {
      if (shelvesRequest.current === controller)
        shelvesRequest.current = undefined;
    }
  }, [client, onError]);

  useEffect(() => {
    const timer = window.setTimeout(() => void loadFirstPage(), 180);
    return () => {
      window.clearTimeout(timer);
      requestVersion.current++;
      cancelCatalogRequests();
    };
  }, [cancelCatalogRequests, loadFirstPage]);

  useEffect(() => {
    void refreshShelves();
    return () => shelvesRequest.current?.abort();
  }, [refreshShelves]);

  const refresh = useCallback(async () => {
    await Promise.all([loadFirstPage(), refreshShelves()]);
  }, [loadFirstPage, refreshShelves]);

  return {
    books,
    shelves,
    initialLoading,
    loadingMore,
    loadMoreFailed,
    hasMore: nextPageToken !== "",
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
    format,
    shelfUid: shelfUid === "all" ? undefined : shelfUid,
    trashed: filter === "trash",
    pageSize: BOOK_PAGE_SIZE,
    pageToken,
  };
}
