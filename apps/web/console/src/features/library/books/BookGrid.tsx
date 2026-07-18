import type { Book } from "@realtime-me/library-contracts";
import { type BooksClient, InfiniteScrollSentinel } from "@realtime-me/library-web";
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
      <InfiniteScrollSentinel
        hasMore={hasMore}
        loading={loadingMore}
        failed={loadMoreFailed}
        loadingLabel="继续加载书籍"
        completeLabel="已加载全部书籍"
        onLoadMore={() => void onLoadMore()}
      />
    </>
  );
}
