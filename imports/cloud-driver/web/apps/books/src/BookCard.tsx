import { BookOpen, MoreHorizontal, Trash2 } from "lucide-react";
import { BookFormat, BookProcessingStatus } from "@cloud-drive/contracts";
import type { Book } from "@cloud-drive/contracts";
import {
  Badge,
  BooksClient,
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  formatBytes,
} from "@cloud-drive/shared";

export function BookCard({
  book,
  client,
  trashed,
  onOpen,
  onRemove,
  onRestore,
}: {
  book: Book;
  client: BooksClient;
  trashed: boolean;
  onOpen: () => void;
  onRemove: () => void;
  onRestore: () => void;
}) {
  const cover = book.coverUrl ? client.coverUrl(book) : "";
  return (
    <article className="group min-w-0">
      <button
        type="button"
        className="relative aspect-[2/3] w-full overflow-hidden rounded-xl border bg-card text-left shadow-lg transition-transform hover:-translate-y-1"
        onClick={onOpen}
        disabled={trashed}
      >
        {cover ? (
          <img
            src={cover}
            alt=""
            loading="lazy"
            decoding="async"
            className="h-full w-full object-cover"
          />
        ) : (
          <div className="grid h-full place-items-center bg-gradient-to-br from-primary/15 to-card">
            <BookOpen className="size-10 text-primary/70" />
          </div>
        )}
        <Badge className="absolute top-2 left-2" variant="secondary">
          {book.format === BookFormat.PDF ? "PDF" : "EPUB"}
        </Badge>
      </button>
      <div className="mt-3 flex items-start gap-2">
        <BookSummary book={book} />
        <BookMenu
          bookTitle={book.title}
          trashed={trashed}
          onRemove={onRemove}
          onRestore={onRestore}
        />
      </div>
    </article>
  );
}

function BookSummary({ book }: { book: Book }) {
  return (
    <div className="min-w-0 flex-1">
      <h2 className="truncate text-sm font-medium" title={book.title}>
        {book.title}
      </h2>
      <p className="mt-1 truncate text-xs text-muted-foreground">
        {book.authors.join("、") || "未知作者"}
      </p>
      <p className="mt-1 text-[11px] text-muted-foreground">
        {formatBytes(book.sizeBytes)}
        {book.processingStatus === BookProcessingStatus.PENDING
          ? " · 处理中"
          : ""}
      </p>
    </div>
  );
}

function BookMenu({
  bookTitle,
  trashed,
  onRemove,
  onRestore,
}: {
  bookTitle: string;
  trashed: boolean;
  onRemove: () => void;
  onRestore: () => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size="icon-xs"
          aria-label={`打开《${bookTitle}》操作菜单`}
          title="打开书籍操作菜单"
        >
          <MoreHorizontal />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {trashed ? (
          <DropdownMenuItem onSelect={onRestore}>恢复</DropdownMenuItem>
        ) : (
          <DropdownMenuItem variant="destructive" onSelect={onRemove}>
            <Trash2 />
            移入回收站
          </DropdownMenuItem>
        )}
        {trashed ? (
          <DropdownMenuItem variant="destructive" onSelect={onRemove}>
            永久删除
          </DropdownMenuItem>
        ) : null}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
