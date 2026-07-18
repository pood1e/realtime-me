import type { ReactNode } from "react";
import { useEffect, useMemo, useState } from "react";
import {
  AlertTriangle,
  Download,
  FilePlus2,
  Folder,
  Link2Off,
  ShieldCheck,
} from "lucide-react";
import type { DriveItem } from "@realtime-me/library-contracts";
import {
  Breadcrumbs,
  Button,
  AppDialog,
  DriveItemView,
  DriveViewModeToggle,
  EmptyState,
  InlineError,
  InfiniteScrollSentinel,
  LoadingIndicator,
  PublicShareClient,
  driveItemIsDirectory,
  driveItemName,
  driveItemUid,
  isImage,
  isPdf,
  isText,
  isUnavailableShareError,
  shareLinkExpiresAt,
  useDriveViewMode,
  useCursorQuery,
} from "@realtime-me/library-web";
import type { ResolvedShare } from "@realtime-me/library-web";

import { API_BASE } from "./config";

type ViewState = "loading" | "ready" | "unavailable" | "error";
type Trail = Readonly<{ id: string; label: string }>;

export function App() {
  const client = useMemo(() => new PublicShareClient(API_BASE), []);
  const token = useMemo(readShareToken, []);
  const [state, setState] = useState<ViewState>(
    token ? "loading" : "unavailable",
  );
  const [share, setShare] = useState<ResolvedShare>();
  const [trail, setTrail] = useState<readonly Trail[]>([]);
  const [preview, setPreview] = useState<DriveItem>();
  const [error, setError] = useState("");
  const [viewMode, setViewMode] = useDriveViewMode(
    "cloud-drive.share.view-mode",
  );

  useEffect(() => {
    if (!token) return;
    const controller = new AbortController();
    void client
      .resolveShare(token, controller.signal)
      .then((resolved) => {
        setShare(resolved);
        setTrail([{ id: "", label: driveItemName(resolved.target) }]);
        setState("ready");
      })
      .catch((resolveError: unknown) => {
        if (controller.signal.aborted) return;
        setError(errorMessage(resolveError));
        setState(
          isUnavailableShareError(resolveError) ? "unavailable" : "error",
        );
      });
    return () => controller.abort();
  }, [client, token]);

  const parentUid = trail.at(-1)?.id ?? "";
  const folderShare = Boolean(share && driveItemIsDirectory(share.target));
  const catalog = useCursorQuery({
    queryKey: ["shared-items", share?.shareLink.uid ?? "", parentUid],
    enabled: Boolean(share && token && folderShare),
    loadPage: async (pageToken, signal) => {
      if (!token) return { items: [], nextPageToken: "" };
      return client.listSharedItemsPage(token, parentUid, pageToken, signal);
    },
  });

  if (state === "loading") {
    return (
      <ShareFrame>
        <LoadingIndicator label="正在打开分享" />
      </ShareFrame>
    );
  }
  if (state === "unavailable") {
    return (
      <ShareFrame>
        <EmptyState
          icon={<Link2Off className="size-6" />}
          title="此分享链接不可用"
          detail="它可能已过期、被撤销，或地址不完整。"
        />
      </ShareFrame>
    );
  }
  if (state === "error" || !share || !token) {
    return (
      <ShareFrame>
        <EmptyState
          icon={<AlertTriangle className="size-6" />}
          title="无法打开分享"
          detail={error || "请稍后重试。"}
        />
      </ShareFrame>
    );
  }

  const visibleItems = folderShare ? catalog.items : [share.target];
  return (
    <main className="min-h-dvh bg-background p-4 text-foreground sm:p-7 lg:p-10">
      <section className="mx-auto w-full max-w-[112rem]">
        <header className="mb-6 border-b pb-5">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0">
              <p className="mb-3 flex items-center gap-2 text-xs font-medium text-primary">
                <ShieldCheck className="size-4" />
                只读分享
              </p>
              <Breadcrumbs
                items={trail}
                onNavigate={(id) => {
                  const index = trail.findIndex((segment) => segment.id === id);
                  setTrail(trail.slice(0, index + 1));
                }}
              />
            </div>
            <DriveViewModeToggle mode={viewMode} onChange={setViewMode} />
          </div>
          <p className="mt-3 text-xs text-muted-foreground">
            有效期至{" "}
            {shareLinkExpiresAt(share.shareLink)?.toLocaleString("zh-CN") ??
              "—"}
          </p>
        </header>
        {catalog.error ? (
          <InlineError
            message={errorMessage(catalog.error)}
            onRetry={() => void catalog.refetch()}
          />
        ) : null}
        {catalog.isPending && folderShare ? (
          <LoadingIndicator label="正在读取分享内容" />
        ) : null}
        {(!folderShare || !catalog.isPending) && !catalog.error ? (
          <>
            <DriveItemView
              mode={viewMode}
              items={visibleItems}
              onOpen={(item) => {
                if (driveItemIsDirectory(item)) {
                  setTrail((current) => [
                    ...current,
                    { id: driveItemUid(item), label: driveItemName(item) },
                  ]);
                } else {
                  setPreview(item);
                }
              }}
              empty={
                <EmptyState
                  icon={<Folder className="size-6" />}
                  title="文件夹为空"
                />
              }
            />
            {folderShare ? (
              <InfiniteScrollSentinel
                hasMore={catalog.hasNextPage}
                loading={catalog.isFetchingNextPage}
                failed={catalog.isFetchNextPageError}
                loadingLabel="继续加载分享内容"
                completeLabel="已加载全部分享内容"
                onLoadMore={() => void catalog.fetchNextPage()}
              />
            ) : null}
          </>
        ) : null}
      </section>
      {preview ? (
        <PublicPreview
          item={preview}
          token={token}
          client={client}
          onClose={() => setPreview(undefined)}
        />
      ) : null}
    </main>
  );
}

function ShareFrame({ children }: { children: ReactNode }) {
  return (
    <main className="grid min-h-dvh place-items-center bg-background p-5 text-foreground">
      {children}
    </main>
  );
}

function PublicPreview({
  item,
  token,
  client,
  onClose,
}: {
  item: DriveItem;
  token: string;
  client: PublicShareClient;
  onClose: () => void;
}) {
  const url = client.contentUrl(token, driveItemUid(item));
  const [text, setText] = useState<string>();
  const [error, setError] = useState("");

  useEffect(() => {
    if (!isText(item)) return;
    const controller = new AbortController();
    void fetch(url, {
      signal: controller.signal,
      referrerPolicy: "no-referrer",
    })
      .then((response) => response.text())
      .then(setText)
      .catch((loadError: unknown) => {
        if (!controller.signal.aborted) setError(errorMessage(loadError));
      });
    return () => controller.abort();
  }, [item, url]);

  return (
    <AppDialog
      open
      title={driveItemName(item)}
      size="preview"
      onClose={onClose}
    >
      {error ? <InlineError message={error} /> : null}
      <div className="min-h-56 max-h-[70dvh] overflow-auto rounded-lg bg-muted/30 p-2">
        {isImage(item) ? (
          <img
            src={url}
            alt={driveItemName(item)}
            className="mx-auto max-h-[68dvh] object-contain"
          />
        ) : isPdf(item) ? (
          <iframe
            src={url}
            title={driveItemName(item)}
            className="h-[68dvh] w-full rounded bg-white"
          />
        ) : isText(item) ? (
          text === undefined ? (
            <LoadingIndicator label="正在载入文本" />
          ) : (
            <pre className="whitespace-pre-wrap p-3 text-xs">{text}</pre>
          )
        ) : (
          <EmptyState
            icon={<FilePlus2 className="size-6" />}
            title="不支持在线预览"
            detail="可以下载后在本地打开。"
          />
        )}
      </div>
      <div className="mt-4 flex justify-end">
        <Button asChild>
          <a href={`${url}?download=1`}>
            <Download />
            下载
          </a>
        </Button>
      </div>
    </AppDialog>
  );
}

function readShareToken(): string | undefined {
  const queryToken = new URLSearchParams(window.location.search).get("token");
  if (queryToken) return queryToken;
  const segments = window.location.pathname.split("/").filter(Boolean);
  const marker = segments.findIndex(
    (segment) => segment === "s" || segment === "share",
  );
  if (marker >= 0) return segments[marker + 1];
  return segments.length === 1 ? segments[0] : undefined;
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : "暂时无法打开此分享。";
}
