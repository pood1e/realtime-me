import { create } from "@bufbuild/protobuf";
import type { Book } from "@realtime-me/library-contracts";
import { ReadingProgressSchema } from "@realtime-me/library-contracts";
import { type BooksClient, InlineError, LoadingIndicator } from "@realtime-me/library-web";
import { Button } from "@realtime-me/web-ui";
import { ChevronLeft, ChevronRight } from "lucide-react";
import type { PDFDocumentProxy, RenderTask } from "pdfjs-dist";
import { GlobalWorkerOptions, getDocument } from "pdfjs-dist";
import pdfWorker from "pdfjs-dist/build/pdf.worker.min.mjs?url";
import { useEffect, useRef, useState } from "react";

GlobalWorkerOptions.workerSrc = pdfWorker;

export function PdfReader({ book, client }: { book: Book; client: BooksClient }) {
  const canvas = useRef<HTMLCanvasElement>(null);
  const documentRef = useRef<PDFDocumentProxy | null>(null);
  const [page, setPage] = useState(1);
  const [pageCount, setPageCount] = useState(book.pageCount || 1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    const task = getDocument({
      url: client.contentUrl(book),
      withCredentials: true,
    });
    void Promise.all([task.promise, client.progress(book.uid).catch(() => undefined)])
      .then(([pdf, progress]) => {
        if (!active) return;
        documentRef.current = pdf;
        setPageCount(pdf.numPages);
        if (progress?.location.case === "pdf") {
          setPage(Math.max(1, progress.location.value.pageNumber));
        }
        setLoading(false);
      })
      .catch((loadError: unknown) => {
        if (!active) return;
        setError(errorMessage(loadError));
        setLoading(false);
      });
    return () => {
      active = false;
      void task.destroy();
    };
  }, [book, client]);

  useEffect(() => {
    const pdf = documentRef.current;
    const target = canvas.current;
    if (!pdf || !target) return;
    let cancelled = false;
    let renderTask: RenderTask | undefined;
    void pdf
      .getPage(page)
      .then(async (pdfPage) => {
        if (cancelled) return;
        const base = pdfPage.getViewport({ scale: 1 });
        const scale = Math.min(2, Math.max(0.75, 980 / base.width));
        const viewport = pdfPage.getViewport({ scale });
        target.width = viewport.width;
        target.height = viewport.height;
        const context = target.getContext("2d");
        if (context) {
          renderTask = pdfPage.render({
            canvas: target,
            canvasContext: context,
            viewport,
          });
          await renderTask.promise;
        }
      })
      .catch((renderError: unknown) => {
        if (!cancelled) setError(errorMessage(renderError));
      });
    void savePage(client, book.uid, page, pageCount);
    return () => {
      cancelled = true;
      renderTask?.cancel();
    };
  }, [book.uid, client, page, pageCount]);

  if (loading) return <LoadingIndicator label="正在打开 PDF" />;
  if (error) return <InlineError message={error} />;
  return (
    <div className="space-y-3">
      <div className="max-h-[68dvh] overflow-auto rounded-lg bg-muted/40 p-2">
        <canvas ref={canvas} className="mx-auto max-w-full shadow-xl" />
      </div>
      <PageControls page={page} pageCount={pageCount} onPageChange={setPage} />
    </div>
  );
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : "无法打开 PDF。";
}

function PageControls({
  page,
  pageCount,
  onPageChange,
}: {
  page: number;
  pageCount: number;
  onPageChange: (page: number) => void;
}) {
  return (
    <div className="flex items-center justify-center gap-3">
      <Button
        variant="outline"
        size="icon"
        disabled={page <= 1}
        onClick={() => onPageChange(page - 1)}
        aria-label="上一页"
      >
        <ChevronLeft />
      </Button>
      <span className="min-w-24 text-center text-sm text-muted-foreground">
        {page} / {pageCount}
      </span>
      <Button
        variant="outline"
        size="icon"
        disabled={page >= pageCount}
        onClick={() => onPageChange(page + 1)}
        aria-label="下一页"
      >
        <ChevronRight />
      </Button>
    </div>
  );
}

async function savePage(
  client: BooksClient,
  bookUid: string,
  page: number,
  pageCount: number,
): Promise<void> {
  await client
    .saveProgress(
      create(ReadingProgressSchema, {
        bookUid,
        progressPercent: page / pageCount,
        location: {
          case: "pdf",
          value: { pageNumber: page, pageCount },
        },
      }),
    )
    .catch(() => undefined);
}
