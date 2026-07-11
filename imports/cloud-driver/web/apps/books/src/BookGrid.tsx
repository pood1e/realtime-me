import { useEffect, useRef } from "react";
import type { Book } from "@cloud-drive/contracts";
import { BooksClient, Button, LoadingIndicator } from "@cloud-drive/shared";
import { BookCard } from "./BookCard";

type BookGridProps = Readonly<{
  books: Book[];
  client: BooksClient;
  trashed: boolean;
  hasMore: boolean;
  loadingMore: boolean;
  loadMoreFailed: boolean;
  onLoadMore: () => Promise<void>;
  onOpen: (book: Book) => void;
  onRemove: (book: Book) => void;
  onRestore: (book: Book) => void;
}>;

export function BookGrid({
  books,
  client,
  trashed,
  hasMore,
  loadingMore,
  loadMoreFailed,
  onLoadMore,
  onOpen,
  onRemove,
  onRestore,
}: BookGridProps) {
  const sentinel = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const target = sentinel.current;
    if (
      !target ||
      !hasMore ||
      loadingMore ||
      loadMoreFailed ||
      !("IntersectionObserver" in window)
    )
      return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry?.isIntersecting) void onLoadMore();
      },
      { rootMargin: "480px 0px" },
    );
    observer.observe(target);
    return () => observer.disconnect();
  }, [hasMore, loadMoreFailed, loadingMore, onLoadMore]);

  return (
    <>
      <div className="grid grid-cols-2 gap-x-4 gap-y-7 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-6 2xl:grid-cols-8">
        {books.map((book) => (
          <BookCard
            key={book.uid}
            book={book}
            client={client}
            trashed={trashed}
            onOpen={() => onOpen(book)}
            onRemove={() => onRemove(book)}
            onRestore={() => onRestore(book)}
          />
        ))}
      </div>
      {hasMore ? (
        <div
          ref={sentinel}
          className="flex min-h-24 items-center justify-center"
          aria-live="polite"
        >
          {loadingMore ? (
            <LoadingIndicator label="继续加载书籍" />
          ) : (
            <Button variant="outline" onClick={() => void onLoadMore()}>
              {loadMoreFailed ? "重试加载" : "加载更多"}
            </Button>
          )}
        </div>
      ) : (
        <p className="sr-only" aria-live="polite">
          已加载全部书籍
        </p>
      )}
    </>
  );
}
