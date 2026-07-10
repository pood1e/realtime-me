import { lazy, Suspense } from "react";
import { BookFormat } from "@cloud-drive/contracts";
import type { Book } from "@cloud-drive/contracts";
import { BooksClient, Dialog, LoadingIndicator } from "@cloud-drive/shared";

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
    <Dialog open title={book.title} size="preview" onClose={onClose}>
      <Suspense fallback={<LoadingIndicator label="正在加载阅读器" />}>
        {book.format === BookFormat.PDF ? (
          <PdfReader book={book} client={client} />
        ) : (
          <EpubReader book={book} client={client} />
        )}
      </Suspense>
    </Dialog>
  );
}
