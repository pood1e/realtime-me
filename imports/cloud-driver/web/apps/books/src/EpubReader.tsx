import { create } from "@bufbuild/protobuf";
import { useEffect, useRef, useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import ePub from "epubjs";
import type { Rendition } from "epubjs";
import { ReadingProgressSchema } from "@cloud-drive/contracts";
import type { Book } from "@cloud-drive/contracts";
import {
  BooksClient,
  Button,
  InlineError,
  LoadingIndicator,
} from "@cloud-drive/shared";

type EpubLocation = Readonly<{
  start: Readonly<{ cfi: string; percentage?: number }>;
}>;

export function EpubReader({
  book,
  client,
}: {
  book: Book;
  client: BooksClient;
}) {
  const host = useRef<HTMLDivElement>(null);
  const rendition = useRef<Rendition | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!host.current) return;
    let active = true;
    let publication: ReturnType<typeof ePub> | undefined;
    void fetchBook(client.contentUrl(book))
      .then(async (data) => {
        if (!active || !host.current) return;
        publication = ePub(data);
        const next = publication.renderTo(host.current, {
          width: "100%",
          height: "64vh",
          spread: "auto",
        });
        rendition.current = next;
        const progress = await client.progress(book.uid).catch(() => undefined);
        await next.display(
          progress?.location.case === "epub"
            ? progress.location.value.cfi
            : undefined,
        );
        next.on("relocated", (location: EpubLocation) => {
          void saveLocation(client, book.uid, location);
        });
        setLoading(false);
      })
      .catch((loadError: unknown) => {
        if (!active) return;
        setError(errorMessage(loadError));
        setLoading(false);
      });
    return () => {
      active = false;
      rendition.current?.destroy();
      publication?.destroy();
    };
  }, [book, client]);

  return (
    <div className="space-y-3">
      {loading ? <LoadingIndicator label="正在打开 EPUB" /> : null}
      {error ? <InlineError message={error} /> : null}
      <div
        ref={host}
        className="overflow-hidden rounded-lg bg-white text-black"
      />
      <div className="flex justify-center gap-3">
        <Button variant="outline" onClick={() => rendition.current?.prev()}>
          <ChevronLeft />
          上一页
        </Button>
        <Button variant="outline" onClick={() => rendition.current?.next()}>
          下一页
          <ChevronRight />
        </Button>
      </div>
    </div>
  );
}

async function fetchBook(url: string): Promise<ArrayBuffer> {
  const response = await fetch(url, { credentials: "include" });
  if (!response.ok) throw new Error(`EPUB 请求失败（${response.status}）`);
  return response.arrayBuffer();
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : "无法打开 EPUB。";
}

async function saveLocation(
  client: BooksClient,
  bookUid: string,
  location: EpubLocation,
): Promise<void> {
  await client
    .saveProgress(
      create(ReadingProgressSchema, {
        bookUid,
        progressPercent: location.start.percentage ?? 0,
        location: {
          case: "epub",
          value: { cfi: location.start.cfi },
        },
      }),
    )
    .catch(() => undefined);
}
