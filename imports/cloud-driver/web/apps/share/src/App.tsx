import type { ReactNode } from "react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { AlertTriangle, Download, FilePlus2, Folder, Grid2X2, Link2Off, ShieldCheck } from "lucide-react";
import type { DriveItem } from "@cloud-drive/contracts";
import {
  apiBaseUrl,
  Breadcrumbs,
  Dialog,
  DriveItemList,
  EmptyState,
  InlineError,
  isImage,
  isPdf,
  isText,
  isUnavailableShareError,
  LoadingIndicator,
  PublicShareClient,
  type ResolvedShare,
  shareLinkExpiresAt,
  ToastProvider,
} from "@cloud-drive/shared";
import { driveItemIsDirectory, driveItemName, driveItemUid } from "@cloud-drive/shared";
import { DEFAULT_SHARE_API_BASE } from "./config";

type ViewState = "loading" | "ready" | "unavailable" | "error";
type Trail = Readonly<{ id: string; label: string }>;

const API_BASE = apiBaseUrl(import.meta.env.VITE_SHARE_API_BASE, DEFAULT_SHARE_API_BASE);

function getShareToken(): string | undefined {
  const queryToken = new URLSearchParams(window.location.search).get("token");
  if (queryToken) {
    return queryToken;
  }
  const segments = window.location.pathname.split("/").filter(Boolean);
  const marker = segments.findIndex((segment) => segment === "s" || segment === "share");
  if (marker >= 0) {
    return segments[marker + 1];
  }
  return segments.length === 1 ? segments[0] : undefined;
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : "暂时无法打开此分享。";
}

function asDownloadUrl(url: string): string {
  const target = new URL(url);
  target.searchParams.set("download", "1");
  return target.toString();
}

export function App() {
  return <ToastProvider><SharedWorkspace /></ToastProvider>;
}

function SharedWorkspace() {
  const client = useMemo(() => new PublicShareClient(API_BASE), []);
  const token = useMemo(getShareToken, []);
  const [state, setState] = useState<ViewState>(token ? "loading" : "unavailable");
  const [resolution, setResolution] = useState<ResolvedShare>();
  const [trail, setTrail] = useState<readonly Trail[]>([]);
  const [items, setItems] = useState<DriveItem[]>([]);
  const [loadingItems, setLoadingItems] = useState(false);
  const [error, setError] = useState<string>();
  const [preview, setPreview] = useState<DriveItem>();

  useEffect(() => {
    if (!token) {
      return;
    }
    const controller = new AbortController();
    void client.resolveShare(token, controller.signal).then((next) => {
      setResolution(next);
      setTrail([{ id: "", label: driveItemName(next.target) }]);
      setState("ready");
    }).catch((resolveError) => {
      setError(errorMessage(resolveError));
      setState(isUnavailableShareError(resolveError) ? "unavailable" : "error");
    });
    return () => controller.abort();
  }, [client, token]);

  const parentUid = trail.at(-1)?.id ?? "";
  const targetIsFolder = resolution ? driveItemIsDirectory(resolution.target) : false;
  const refreshItems = useCallback(() => {
    if (!resolution || !token || !targetIsFolder) {
      return;
    }
    const controller = new AbortController();
    setLoadingItems(true);
    setError(undefined);
    void client.listSharedItems(token, parentUid, controller.signal).then(setItems).catch((listError) => {
      setError(errorMessage(listError));
    }).finally(() => setLoadingItems(false));
    return () => controller.abort();
  }, [client, parentUid, resolution, targetIsFolder, token]);

  useEffect(() => refreshItems(), [refreshItems]);

  const openItem = (item: DriveItem) => {
    if (driveItemIsDirectory(item)) {
      const uid = driveItemUid(item);
      setTrail((current) => [...current, { id: uid, label: driveItemName(item) }]);
      return;
    }
    setPreview(item);
  };

  if (state === "loading") {
    return <ShareFrame><CenteredState><LoadingIndicator label="正在打开分享" /></CenteredState></ShareFrame>;
  }
  if (state === "unavailable") {
    return <ShareFrame><CenteredState><EmptyState icon={<Link2Off className="size-6" />} title="此分享链接不可用" detail="它可能已过期、被撤销，或链接地址不完整。" /></CenteredState></ShareFrame>;
  }
  if (state === "error" || !resolution || !token) {
    return <ShareFrame><CenteredState><EmptyState icon={<AlertTriangle className="size-6" />} title="无法打开分享" detail={error ?? "请稍后重试。"} /></CenteredState></ShareFrame>;
  }

  const visibleItems = targetIsFolder ? items : [resolution.target];
  return (
    <ShareFrame>
      <header className="shrink-0 border-b border-white/[0.08] pb-5">
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0">
            <div className="mb-3 flex items-center gap-2 text-xs text-sky-200"><ShieldCheck className="size-4" />只读分享</div>
            <Breadcrumbs items={trail} onNavigate={(id) => setTrail((current) => current.slice(0, current.findIndex((segment) => segment.id === id) + 1))} />
          </div>
          <div className="hidden shrink-0 items-center gap-2 text-xs text-slate-500 sm:flex"><Grid2X2 className="size-4 text-sky-300" />个人云盘</div>
        </div>
        <p className="mt-3 text-sm text-slate-500">有效期至 {shareLinkExpiresAt(resolution.shareLink)?.toLocaleString("zh-CN") ?? "—"}</p>
      </header>
      <section className="mt-5 min-h-0 flex-1 overflow-y-auto pb-6">
        {error ? <InlineError message={error} onRetry={() => refreshItems()} /> : null}
        {loadingItems ? <LoadingIndicator label="正在读取分享内容" /> : null}
        {!loadingItems && !error ? <DriveItemList items={visibleItems} onOpen={openItem} empty={<EmptyState icon={<Folder className="size-6" />} title="此文件夹为空" />} /> : null}
      </section>
      {preview ? <PublicPreview item={preview} token={token} client={client} open onClose={() => setPreview(undefined)} /> : null}
    </ShareFrame>
  );
}

function ShareFrame({ children }: { children: ReactNode }) {
  return <main className="h-dvh w-full overflow-hidden bg-slate-950 text-slate-100"><section className="flex h-full w-full flex-col px-4 py-5 sm:px-6 sm:py-6 lg:px-8 lg:py-8 2xl:px-10">{children}</section></main>;
}

function CenteredState({ children }: { children: ReactNode }) {
  return <div className="flex min-h-0 flex-1 items-center justify-center">{children}</div>;
}

function PublicPreview({ item, token, client, open, onClose }: { item: DriveItem; token: string; client: PublicShareClient; open: boolean; onClose: () => void }) {
  const [text, setText] = useState<string>();
  const [error, setError] = useState<string>();
  const url = client.contentUrl(token, driveItemUid(item));
  useEffect(() => { if (!isText(item)) return; const controller = new AbortController(); setText(undefined); setError(undefined); void client.readText(url, controller.signal).then(setText).catch((loadError) => setError(errorMessage(loadError))); return () => controller.abort(); }, [client, item, url]);
  return <Dialog open={open} title={driveItemName(item)} size="preview" onClose={onClose}>{error ? <InlineError message={error} /> : null}<div className="min-h-48">{isImage(item) ? <img src={url} referrerPolicy="no-referrer" alt={driveItemName(item)} className="max-h-[calc(100dvh-10rem)] w-full rounded-lg object-contain sm:max-h-[calc(90dvh-10rem)]" /> : isPdf(item) ? <iframe src={url} referrerPolicy="no-referrer" title={driveItemName(item)} className="h-[calc(100dvh-10rem)] min-h-80 w-full rounded-lg border border-white/10 bg-white sm:h-[calc(90dvh-10rem)]" /> : isText(item) ? text === undefined ? <LoadingIndicator label="正在载入文本" /> : <pre className="max-h-[calc(100dvh-10rem)] overflow-auto rounded-lg border border-white/10 bg-slate-950 p-3 text-xs leading-5 text-slate-300 sm:max-h-[calc(90dvh-10rem)]">{text}</pre> : <EmptyState icon={<FilePlus2 className="size-6" />} title="此文件不支持在线预览" detail="你可以下载后在本地打开。" />}</div><div className="mt-4 flex justify-end"><a href={asDownloadUrl(url)} referrerPolicy="no-referrer" className="inline-flex h-10 items-center gap-2 rounded-lg bg-sky-500 px-4 text-sm font-medium text-slate-950 hover:bg-sky-300"><Download className="size-4" />下载</a></div></Dialog>;
}
