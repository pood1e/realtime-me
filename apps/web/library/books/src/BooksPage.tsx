import { useCallback, useMemo, useState } from "react";
import { Library, Plus, Search } from "lucide-react";
import type { Book } from "@realtime-me/library-contracts";
import {
  BooksClient,
  Button,
  EmptyState,
  Input,
  LoadingIndicator,
  PrivateAppShell,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  UploadButton,
  UploadClient,
  useDialog,
  useToast,
} from "@realtime-me/library-web";
import { BookGrid } from "./BookGrid";
import { BookReader } from "./BookReader";
import { API_BASE, APP_LINKS } from "./config";
import { useBookCatalog } from "./useBookCatalog";
import type { BookFilter } from "./useBookCatalog";

export function BooksPage() {
  const client = useMemo(() => new BooksClient(API_BASE), []);
  const uploader = useMemo(() => new UploadClient(API_BASE), []);
  const { showToast } = useToast();
  const { confirm, prompt } = useDialog();
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState<BookFilter>("all");
  const [shelfUid, setShelfUid] = useState("all");
  const [reader, setReader] = useState<Book>();
  const onLoadError = useCallback(
    (error: unknown) => showToast(message(error), "error"),
    [showToast],
  );
  const catalog = useBookCatalog({
    client,
    query,
    filter,
    shelfUid,
    onError: onLoadError,
  });

  const upload = async (files: File[]) => {
    for (const file of files)
      try {
        const uid = await uploader.upload(file);
        await client.importUpload(uid);
        showToast(`${file.name} 已加入书架`);
      } catch (error) {
        showToast(`${file.name}: ${message(error)}`, "error");
      }
    await catalog.refresh();
  };
  const createShelf = async () => {
    const name = await prompt({ title: "新建书架", label: "书架名称" });
    if (!name) return;
    try {
      await client.createShelf(name);
      await catalog.refresh();
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  const remove = async (book: Book) => {
    if (filter === "trash") {
      if (
        !(await confirm({
          title: "永久删除书籍",
          description: `“${book.title}”将无法恢复。`,
          confirmLabel: "永久删除",
          destructive: true,
        }))
      )
        return;
      await client.purge(book.uid);
    } else await client.trash(book.uid);
    await catalog.refresh();
  };
  const emptyTrash = async () => {
    if (
      !(await confirm({
        title: "清空书籍回收站",
        description: "回收站中的全部书籍将被永久删除，此操作无法撤销。",
        confirmLabel: "永久删除全部书籍",
        destructive: true,
      }))
    )
      return;
    try {
      await client.emptyTrash();
      await catalog.refresh();
      showToast("回收站已清空");
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  const restore = async (book: Book) => {
    await client.restore(book.uid);
    await catalog.refresh();
  };

  return (
    <PrivateAppShell
      app="books"
      title="书架"
      subtitle={catalogSubtitle(
        catalog.books.length,
        catalog.hasMore,
        catalog.initialLoading,
      )}
      apiBase={API_BASE}
      links={APP_LINKS}
      actions={
        filter === "trash" ? (
          <Button variant="destructive" onClick={() => void emptyTrash()}>
            清空
          </Button>
        ) : (
          <UploadButton
            accept=".pdf,.epub,application/pdf,application/epub+zip"
            onFiles={upload}
            label="导入书籍"
          />
        )
      }
    >
      <div className="mb-7 flex flex-col gap-3 lg:flex-row lg:items-center">
        <div className="relative min-w-0 flex-1">
          <Search className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="搜索书名、作者或系列"
            className="pl-9"
          />
        </div>
        <Select
          value={filter}
          onValueChange={(value) => setFilter(value as BookFilter)}
        >
          <SelectTrigger className="w-full lg:w-40">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部格式</SelectItem>
            <SelectItem value="pdf">PDF</SelectItem>
            <SelectItem value="epub">EPUB</SelectItem>
            <SelectItem value="trash">回收站</SelectItem>
          </SelectContent>
        </Select>
        <Select value={shelfUid} onValueChange={setShelfUid}>
          <SelectTrigger className="w-full lg:w-48">
            <SelectValue placeholder="全部书架" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部书架</SelectItem>
            {catalog.shelves.map((shelf) => (
              <SelectItem key={shelf.uid} value={shelf.uid}>
                {shelf.displayName} · {shelf.bookCount}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button variant="outline" onClick={() => void createShelf()}>
          <Plus />
          新建书架
        </Button>
      </div>
      {catalog.initialLoading ? (
        <LoadingIndicator label="正在整理书架" />
      ) : catalog.books.length ? (
        <BookGrid
          books={catalog.books}
          client={client}
          trashed={filter === "trash"}
          hasMore={catalog.hasMore}
          loadingMore={catalog.loadingMore}
          loadMoreFailed={catalog.loadMoreFailed}
          onLoadMore={catalog.loadMore}
          onOpen={setReader}
          onRemove={(book) => void remove(book)}
          onRestore={(book) => void restore(book)}
        />
      ) : (
        <EmptyState
          icon={<Library className="size-6" />}
          title="书架还是空的"
          detail="导入 PDF 或 EPUB 后，系统会自动提取封面与元数据。"
        />
      )}
      {reader ? (
        <BookReader
          book={reader}
          client={client}
          onClose={() => setReader(undefined)}
        />
      ) : null}
    </PrivateAppShell>
  );
}

function message(error: unknown) {
  return error instanceof Error ? error.message : "操作未完成";
}

function catalogSubtitle(count: number, hasMore: boolean, loading: boolean) {
  if (loading) return "正在加载";
  return hasMore ? `已加载 ${count} 本` : `${count} 本书`;
}
