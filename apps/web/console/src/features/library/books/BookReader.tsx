import type { Book } from "@realtime-me/library-contracts";
import { BookFormat } from "@realtime-me/library-contracts";
import { AppDialog, type BooksClient, LoadingIndicator } from "@realtime-me/library-web";
import { lazy, Suspense } from "react";

const PdfReader = lazy(() =>
  import("./PdfReader").then((module) => ({ default: module.PdfReader })),
);
const EpubReader = lazy(() =>
  import("./EpubReader").then((module) => ({ default: module.EpubReader })),
);

export function BookReader({
  book,
  client,
  onClose,
}: {
  book: Book;
  client: BooksClient;
  onClose: () => void;
}) {
  return (
    <AppDialog open title={book.title} size="preview" onClose={onClose}>
      <Suspense fallback={<LoadingIndicator label="正在加载阅读器" />}>
        {book.format === BookFormat.PDF ? (
          <PdfReader book={book} client={client} />
        ) : (
          <EpubReader book={book} client={client} />
        )}
      </Suspense>
    </AppDialog>
  );
}
