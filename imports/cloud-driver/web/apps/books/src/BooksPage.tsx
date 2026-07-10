import { useCallback, useEffect, useMemo, useState } from "react";
import { Library, Plus, Search } from "lucide-react";
import { BookFormat } from "@cloud-drive/contracts";
import type { Book, Shelf } from "@cloud-drive/contracts";
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
  useToast,
} from "@cloud-drive/shared";
import { BookCard } from "./BookCard";
import { BookReader } from "./BookReader";
import { API_BASE, APP_LINKS } from "./config";

type Filter = "all" | "pdf" | "epub" | "trash";

export function BooksPage() {
  const client = useMemo(() => new BooksClient(API_BASE), []);
  const uploader = useMemo(() => new UploadClient(API_BASE), []);
  const { showToast } = useToast();
  const [books, setBooks] = useState<Book[]>([]);
  const [shelves, setShelves] = useState<Shelf[]>([]);
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState<Filter>("all");
  const [shelfUid, setShelfUid] = useState("all");
  const [loading, setLoading] = useState(true);
  const [reader, setReader] = useState<Book>();

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const format =
        filter === "pdf"
          ? BookFormat.PDF
          : filter === "epub"
            ? BookFormat.EPUB
            : undefined;
      const [nextBooks, nextShelves] = await Promise.all([
        client.list({
          query,
          format,
          shelfUid: shelfUid === "all" ? undefined : shelfUid,
          trashed: filter === "trash",
        }),
        client.shelves(),
      ]);
      setBooks(nextBooks);
      setShelves(nextShelves);
    } catch (error) {
      showToast(message(error), "error");
    } finally {
      setLoading(false);
    }
  }, [client, filter, query, shelfUid, showToast]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 180);
    return () => window.clearTimeout(timer);
  }, [load]);

  const upload = async (files: File[]) => {
    for (const file of files)
      try {
        const uid = await uploader.upload(file);
        await client.importUpload(uid);
        showToast(`${file.name} 已加入书架`);
      } catch (error) {
        showToast(`${file.name}: ${message(error)}`, "error");
      }
    await load();
  };
  const createShelf = async () => {
    const name = window.prompt("书架名称");
    if (!name?.trim()) return;
    try {
      await client.createShelf(name.trim());
      await load();
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  const remove = async (book: Book) => {
    if (filter === "trash") {
      if (!window.confirm("永久删除这本书？")) return;
      await client.purge(book.uid);
    } else await client.trash(book.uid);
    await load();
  };
  const restore = async (book: Book) => {
    await client.restore(book.uid);
    await load();
  };

  return (
    <PrivateAppShell
      app="books"
      title="书架"
      subtitle={`${books.length} 本书`}
      apiBase={API_BASE}
      links={APP_LINKS}
      actions={
        filter === "trash" ? (
          <Button
            variant="destructive"
            onClick={() => void emptyTrash(client, load, showToast)}
          >
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
          onValueChange={(value) => setFilter(value as Filter)}
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
            {shelves.map((shelf) => (
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
      {loading ? (
        <LoadingIndicator label="正在整理书架" />
      ) : books.length ? (
        <div className="grid grid-cols-2 gap-x-4 gap-y-7 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-6 2xl:grid-cols-8">
          {books.map((book) => (
            <BookCard
              key={book.uid}
              book={book}
              client={client}
              trashed={filter === "trash"}
              onOpen={() => setReader(book)}
              onRemove={() => void remove(book)}
              onRestore={() => void restore(book)}
            />
          ))}
        </div>
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
async function emptyTrash(
  client: BooksClient,
  reload: () => Promise<void>,
  toast: (message: string, variant?: "default" | "error") => void,
) {
  if (!window.confirm("永久删除书籍回收站？")) return;
  try {
    await client.emptyTrash();
    await reload();
    toast("回收站已清空");
  } catch (error) {
    toast(message(error), "error");
  }
}
