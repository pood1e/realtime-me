import type { ChangeEvent, FormEvent, ReactNode } from "react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  ArchiveRestore,
  Download,
  FilePlus2,
  FolderPlus,
  Grid2X2,
  Link2,
  LoaderCircle,
  LogOut,
  MoreHorizontal,
  MoveRight,
  Pencil,
  Search,
  Share2,
  Trash2,
  UploadCloud,
  X,
} from "lucide-react";
import type { DriveItem, ShareLink } from "@cloud-drive/contracts";
import {
  apiBaseUrl,
  Breadcrumbs,
  Button,
  cn,
  Dialog,
  driveItemDeletedAt,
  driveItemIsDirectory,
  driveItemName,
  driveItemUid,
  DriveItemView,
  DriveViewModeToggle,
  EmptyState,
  FileGlyph,
  formatBytes,
  InlineError,
  IconButton,
  isImage,
  isPdf,
  isText,
  isUnauthenticatedError,
  LoadingIndicator,
  PrivateDriveClient,
  shareLinkExpiresAt,
  shareLinkRevokedAt,
  shareLinkUid,
  ToastProvider,
  useDriveViewMode,
  useToast,
} from "@cloud-drive/shared";
import { DEFAULT_PRIVATE_API_BASE } from "./config";

type View = "drive" | "trash";
type UploadState = "queued" | "uploading" | "completed" | "failed";
type UploadTask = Readonly<{
  id: string;
  file: File;
  parentUid: string;
  state: UploadState;
  uploadedBytes: number;
  uploadUid?: string;
  error?: string;
}>;
type Trail = Readonly<{ id: string; label: string }>;
type DialogState =
  | Readonly<{ kind: "folder" }>
  | Readonly<{ kind: "rename" | "move" | "share" | "preview" | "purge"; item: DriveItem }>
  | Readonly<{ kind: "empty-trash" }>
  | null;
type SessionState = "checking" | "authenticated" | "unauthenticated" | "unavailable";

const API_BASE = apiBaseUrl(import.meta.env.VITE_PRIVATE_API_BASE, DEFAULT_PRIVATE_API_BASE);
const ROOT_TRAIL: readonly Trail[] = [{ id: "", label: "我的文件" }];

function asErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : "操作未完成，请稍后重试。";
}

function futureDate(days: number): string {
  const date = new Date();
  date.setDate(date.getDate() + days);
  return date.toISOString().slice(0, 10);
}

function asDownloadUrl(url: string): string {
  const target = new URL(url);
  target.searchParams.set("download", "1");
  return target.toString();
}

export function App() {
  return (
    <ToastProvider>
      <PrivatePortal />
    </ToastProvider>
  );
}

function PrivatePortal() {
  const client = useMemo(() => new PrivateDriveClient(API_BASE), []);
  const [sessionState, setSessionState] = useState<SessionState>("checking");
  const [sessionError, setSessionError] = useState<string>();

  const checkSession = useCallback(async (signal?: AbortSignal) => {
    setSessionState("checking");
    setSessionError(undefined);
    try {
      await client.getSession(signal);
      if (!signal?.aborted) {
        setSessionState("authenticated");
      }
    } catch (error) {
      if (signal?.aborted) {
        return;
      }
      if (isUnauthenticatedError(error)) {
        setSessionState("unauthenticated");
        return;
      }
      setSessionError(asErrorMessage(error));
      setSessionState("unavailable");
    }
  }, [client]);

  useEffect(() => {
    const controller = new AbortController();
    void checkSession(controller.signal);
    return () => controller.abort();
  }, [checkSession]);

  const logout = useCallback(async () => {
    await client.logout();
    setSessionState("unauthenticated");
  }, [client]);
  const requireLogin = useCallback(() => setSessionState("unauthenticated"), []);

  if (sessionState === "checking") {
    return (
      <SessionFrame>
        <LoadingIndicator label="正在检查登录状态" />
      </SessionFrame>
    );
  }
  if (sessionState === "unauthenticated") {
    return <LoginScreen client={client} onAuthenticated={() => setSessionState("authenticated")} />;
  }
  if (sessionState === "unavailable") {
    return (
      <SessionFrame>
        <InlineError message={sessionError ?? "无法检查登录状态。"} onRetry={() => void checkSession()} />
      </SessionFrame>
    );
  }
  return <DriveWorkspace client={client} onLogout={logout} onUnauthenticated={requireLogin} />;
}

function SessionFrame({ children }: { children: ReactNode }) {
  return (
    <main className="grid min-h-screen place-items-center bg-slate-950 px-4 text-slate-100">
      <section className="w-full max-w-sm rounded-2xl border border-white/10 bg-slate-900/75 p-6 shadow-2xl shadow-black/30">
        {children}
      </section>
    </main>
  );
}

function LoginScreen({ client, onAuthenticated }: { client: PrivateDriveClient; onAuthenticated: () => void }) {
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string>();
  const [submitting, setSubmitting] = useState(false);

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (!password || submitting) {
      return;
    }
    setSubmitting(true);
    setError(undefined);
    try {
      await client.login(password);
      setPassword("");
      onAuthenticated();
    } catch (loginError) {
      setError(
        isUnauthenticatedError(loginError)
          ? "密码错误或尝试过于频繁，请稍后重试。"
          : asErrorMessage(loginError),
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <SessionFrame>
      <div className="mb-7 flex items-center gap-3">
        <div className="grid size-10 place-items-center rounded-xl bg-sky-400 text-slate-950">
          <Grid2X2 className="size-5" />
        </div>
        <div>
          <h1 className="text-base font-semibold text-white">个人云盘</h1>
          <p className="text-sm text-slate-400">输入密码以继续</p>
        </div>
      </div>
      <form className="space-y-4" onSubmit={(event) => void submit(event)}>
        <label className="block text-sm text-slate-300">
          密码
          <input
            aria-invalid={Boolean(error)}
            autoComplete="current-password"
            autoFocus
            className="mt-2 h-10 w-full rounded-lg border border-white/10 bg-slate-950 px-3 text-sm text-white outline-none focus:border-sky-300 focus:ring-2 focus:ring-sky-300/20"
            name="password"
            onChange={(event) => setPassword(event.target.value)}
            required
            type="password"
            value={password}
          />
        </label>
        {error ? <p role="alert" className="text-sm text-rose-300">{error}</p> : null}
        <Button className="w-full" type="submit" disabled={submitting || !password}>
          {submitting ? "登录中" : "登录"}
        </Button>
      </form>
    </SessionFrame>
  );
}

function DriveWorkspace({
  client,
  onLogout,
  onUnauthenticated,
}: {
  client: PrivateDriveClient;
  onLogout: () => Promise<void>;
  onUnauthenticated: () => void;
}) {
  const { showToast } = useToast();
  const [view, setView] = useState<View>("drive");
  const [viewMode, setViewMode] = useDriveViewMode("cloud-drive.private.view-mode");
  const [trail, setTrail] = useState<readonly Trail[]>(ROOT_TRAIL);
  const [items, setItems] = useState<DriveItem[]>([]);
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>();
  const [reloadToken, setReloadToken] = useState(0);
  const [dialog, setDialog] = useState<DialogState>(null);
  const [tasks, setTasks] = useState<UploadTask[]>([]);
  const [busy, setBusy] = useState(false);
  const [loggingOut, setLoggingOut] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const processing = useRef(false);
  const tasksRef = useRef<UploadTask[]>([]);

  const currentDirectoryUid = view === "trash" ? "" : (trail.at(-1)?.id ?? "");
  const isSearching = view === "drive" && query.trim().length > 0;
  const refresh = useCallback(() => setReloadToken((value) => value + 1), []);

  const replaceTasks = useCallback((updater: (current: UploadTask[]) => UploadTask[]) => {
    setTasks((current) => {
      const next = updater(current);
      tasksRef.current = next;
      return next;
    });
  }, []);

  const updateTask = useCallback((id: string, changes: Partial<UploadTask>) => {
    replaceTasks((current) => current.map((task) => (task.id === id ? { ...task, ...changes } : task)));
  }, [replaceTasks]);

  const processQueue = useCallback(async () => {
    if (processing.current) {
      return;
    }
    processing.current = true;
    try {
      while (true) {
        const task = tasksRef.current.find((candidate) => candidate.state === "queued");
        if (!task) {
          return;
        }
        updateTask(task.id, { state: "uploading", error: undefined });
        try {
          await client.uploadFile(task.file, {
            parentUid: task.parentUid,
            resumeUploadUid: task.uploadUid,
            onSession: (uploadUid) => updateTask(task.id, { uploadUid }),
            onProgress: ({ uploadedBytes }) => updateTask(task.id, { uploadedBytes }),
          });
          updateTask(task.id, { state: "completed", uploadedBytes: task.file.size });
          showToast(`${task.file.name} 已上传`);
          refresh();
        } catch (uploadError) {
          updateTask(task.id, { state: "failed", error: asErrorMessage(uploadError) });
        }
      }
    } finally {
      processing.current = false;
    }
  }, [client, refresh, showToast, updateTask]);

  useEffect(() => {
    void processQueue();
  }, [processQueue, tasks]);

  useEffect(() => {
    const controller = new AbortController();
    const fetchItems = async () => {
      setLoading(true);
      setError(undefined);
      try {
        const listed = isSearching
          ? await client.searchDriveItems(query.trim(), controller.signal)
          : view === "trash"
            ? await client.listTrashedItems(controller.signal)
            : await client.listDriveItems(currentDirectoryUid, false, controller.signal);
        setItems(listed.filter((item) => (view === "trash" ? Boolean(driveItemDeletedAt(item)) : !driveItemDeletedAt(item))));
      } catch (loadError) {
        if (!controller.signal.aborted) {
          if (isUnauthenticatedError(loadError)) {
            onUnauthenticated();
            return;
          }
          setError(asErrorMessage(loadError));
        }
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false);
        }
      }
    };
    void fetchItems();
    return () => controller.abort();
  }, [client, currentDirectoryUid, isSearching, onUnauthenticated, query, reloadToken, view]);

  const openItem = useCallback((item: DriveItem) => {
    if (driveItemIsDirectory(item) && view === "drive") {
      setTrail((current) => {
        const uid = driveItemUid(item);
        const existingIndex = current.findIndex((segment) => segment.id === uid);
        return existingIndex >= 0 ? current.slice(0, existingIndex + 1) : [...current, { id: uid, label: driveItemName(item) }];
      });
      setQuery("");
      return;
    }
    if (!driveItemIsDirectory(item)) {
      setDialog({ kind: "preview", item });
    }
  }, [view]);

  const switchView = (nextView: View) => {
    setView(nextView);
    setQuery("");
    if (nextView === "drive") {
      setTrail(ROOT_TRAIL);
    }
  };

  const uploadFiles = (event: ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = [...(event.target.files ?? [])];
    if (!selectedFiles.length) {
      return;
    }
    replaceTasks((current) => [
      ...current,
      ...selectedFiles.map((file) => ({
        id: crypto.randomUUID(),
        file,
        parentUid: currentDirectoryUid,
        state: "queued" as const,
        uploadedBytes: 0,
      })),
    ]);
    event.target.value = "";
  };

  const execute = async <T,>(action: () => Promise<T>, successMessage: string) => {
    setBusy(true);
    try {
      await action();
      setDialog(null);
      showToast(successMessage);
      refresh();
    } catch (actionError) {
      showToast(asErrorMessage(actionError), "error");
    } finally {
      setBusy(false);
    }
  };

  const logout = async () => {
    setLoggingOut(true);
    try {
      await onLogout();
    } catch (logoutError) {
      showToast(asErrorMessage(logoutError), "error");
    } finally {
      setLoggingOut(false);
    }
  };

  const sortItems = [...items].sort((left, right) => {
    const directoryFirst = Number(driveItemIsDirectory(right)) - Number(driveItemIsDirectory(left));
    return directoryFirst || driveItemName(left).localeCompare(driveItemName(right), "zh-CN");
  });

  return (
    <main className="h-dvh w-full overflow-hidden bg-slate-950 text-slate-100">
      <input ref={inputRef} onChange={uploadFiles} type="file" multiple className="hidden" />
      <div className="flex h-full w-full">
        <aside className="hidden h-full w-60 shrink-0 overflow-y-auto border-r border-white/[0.07] bg-slate-950/75 px-4 py-5 lg:flex lg:flex-col">
          <div className="mb-9 flex items-center gap-3 px-2">
            <div className="grid size-9 place-items-center rounded-xl bg-sky-400 text-slate-950"><Grid2X2 className="size-5" /></div>
            <div><p className="text-sm font-semibold tracking-tight text-white">个人云盘</p><p className="text-xs text-slate-500">私有文件空间</p></div>
          </div>
          <nav className="space-y-1">
            <NavButton active={view === "drive"} onClick={() => switchView("drive")} icon={<Grid2X2 className="size-4" />}>文件</NavButton>
            <NavButton active={view === "trash"} onClick={() => switchView("trash")} icon={<Trash2 className="size-4" />}>回收站</NavButton>
          </nav>
          <div className="mt-auto rounded-xl border border-white/[0.08] bg-white/[0.025] p-3 text-xs leading-5 text-slate-400">
            文件保留在你的私有服务器上。
          </div>
        </aside>

        <section className="flex h-full min-w-0 flex-1 flex-col overflow-hidden px-4 py-4 sm:px-6 sm:py-6 xl:px-8 2xl:px-10">
          <header className="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-white/[0.07] pb-4 sm:pb-5">
            <div className="flex min-w-0 items-center gap-3">
              <div className="grid size-9 shrink-0 place-items-center rounded-xl bg-sky-400 text-slate-950 lg:hidden"><Grid2X2 className="size-5" /></div>
              <div className="min-w-0">
                <div className="mb-1 lg:hidden"><span className="text-sm font-semibold text-white">个人云盘</span></div>
                {view === "drive" ? <Breadcrumbs items={trail} onNavigate={(id) => setTrail((current) => current.slice(0, current.findIndex((segment) => segment.id === id) + 1))} /> : <p className="text-sm text-slate-300">回收站</p>}
              </div>
            </div>
            <div className="flex items-center gap-2">
              <DriveViewModeToggle mode={viewMode} onChange={setViewMode} />
              {view === "drive" ? (
                <>
                  <Button variant="outline" className="hidden sm:inline-flex" onClick={() => setDialog({ kind: "folder" })}>
                    <FolderPlus className="size-4" />
                    新建文件夹
                  </Button>
                  <Button onClick={() => inputRef.current?.click()}>
                    <UploadCloud className="size-4" />
                    上传
                  </Button>
                </>
              ) : (
                <Button variant="destructive" disabled={busy || loading || items.length === 0} onClick={() => setDialog({ kind: "empty-trash" })}>
                  <Trash2 className="size-4" />
                  清空回收站
                </Button>
              )}
              <IconButton label="退出登录" disabled={loggingOut} onClick={() => void logout()}>
                <LogOut className="size-4" />
              </IconButton>
            </div>
          </header>

          <div className="mt-4 flex shrink-0 flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex rounded-lg border border-white/[0.08] bg-white/[0.025] p-1 lg:hidden">
              <NavButton compact active={view === "drive"} onClick={() => switchView("drive")} icon={<Grid2X2 className="size-4" />}>文件</NavButton>
              <NavButton compact active={view === "trash"} onClick={() => switchView("trash")} icon={<Trash2 className="size-4" />}>回收站</NavButton>
            </div>
            {view === "drive" ? (
              <label className="flex h-10 w-full max-w-md items-center gap-2 rounded-lg border border-white/[0.09] bg-white/[0.025] px-3 text-slate-400 focus-within:border-sky-300/50 focus-within:ring-2 focus-within:ring-sky-300/20">
                <Search className="size-4 shrink-0" />
                <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="搜索文件和文件夹" className="min-w-0 flex-1 bg-transparent text-sm text-white outline-none placeholder:text-slate-600" />
                {query ? <button aria-label="清除搜索" title="清除搜索" type="button" onClick={() => setQuery("")}><X className="size-4" /></button> : null}
              </label>
            ) : <p className="text-sm text-slate-500">删除的文件会在 30 天后永久清除。</p>}
            {isSearching ? <p className="text-sm text-slate-500">搜索结果</p> : null}
          </div>

          <section className="mt-5 min-h-0 flex-1 overflow-y-auto pb-6">
            {error ? <InlineError message={error} onRetry={refresh} /> : null}
            {loading ? <LoadingIndicator label="正在读取文件" /> : null}
            {!loading && !error ? (
              <DriveItemView
                mode={viewMode}
                items={sortItems}
                onOpen={openItem}
                empty={<EmptyState icon={view === "trash" ? <Trash2 className="size-6" /> : <FilePlus2 className="size-6" />} title={view === "trash" ? "回收站为空" : isSearching ? "没有匹配的文件" : "这里还没有文件"} detail={view === "drive" && !isSearching ? "上传文件或新建文件夹，即可从这里开始管理。" : undefined} action={view === "drive" && !isSearching ? <Button onClick={() => inputRef.current?.click()}><UploadCloud className="size-4" />上传文件</Button> : undefined} />}
                actions={(item) => <ItemActions item={item} inTrash={view === "trash"} onRename={() => setDialog({ kind: "rename", item })} onMove={() => setDialog({ kind: "move", item })} onShare={() => setDialog({ kind: "share", item })} onPreview={() => setDialog({ kind: "preview", item })} onDelete={() => void execute(() => client.deleteDriveItem(driveItemUid(item)), "已移入回收站")} onRestore={() => void execute(() => client.restoreDriveItem(driveItemUid(item)), "已恢复")} onPurge={() => setDialog({ kind: "purge", item })}/>}
              />
            ) : null}
          </section>
          <UploadQueue tasks={tasks} onDismiss={(id) => replaceTasks((current) => current.filter((task) => task.id !== id))} onRetry={(id) => updateTask(id, { state: "queued", error: undefined })} />
        </section>
      </div>

      <FolderDialog open={dialog?.kind === "folder"} busy={busy} onClose={() => setDialog(null)} onCreate={(name) => void execute(() => client.createDirectory(currentDirectoryUid, name), "已创建文件夹")} />
      {dialog?.kind === "rename" ? <RenameDialog item={dialog.item} open busy={busy} onClose={() => setDialog(null)} onRename={(name) => void execute(() => client.renameDriveItem(driveItemUid(dialog.item), name), "已重命名")} /> : null}
      {dialog?.kind === "move" ? <MoveDialog item={dialog.item} open busy={busy} currentDirectories={items.filter(driveItemIsDirectory)} onClose={() => setDialog(null)} onMove={(parentUid) => void execute(() => client.moveDriveItem(driveItemUid(dialog.item), parentUid), "已移动")} /> : null}
      {dialog?.kind === "share" ? <ShareDialog item={dialog.item} client={client} open busy={busy} onClose={() => setDialog(null)} onBusyChange={setBusy} /> : null}
      {dialog?.kind === "preview" ? <PreviewDialog item={dialog.item} client={client} open onClose={() => setDialog(null)} /> : null}
      {dialog?.kind === "purge" ? <DestructiveConfirmationDialog busy={busy} title={`永久删除“${driveItemName(dialog.item)}”？`} description={driveItemIsDirectory(dialog.item) ? "文件夹及其中全部内容将被永久删除，无法恢复。" : "文件将被永久删除，无法恢复。"} confirmLabel="永久删除" onClose={() => setDialog(null)} onConfirm={() => void execute(() => client.purgeDriveItem(driveItemUid(dialog.item)), "已永久删除")} /> : null}
      {dialog?.kind === "empty-trash" ? <DestructiveConfirmationDialog busy={busy} title="清空回收站？" description="回收站中的全部文件和文件夹将被永久删除，无法恢复。" confirmLabel="清空回收站" onClose={() => setDialog(null)} onConfirm={() => void execute(() => client.emptyTrash(), "回收站已清空")} /> : null}
    </main>
  );
}

function NavButton({ active, compact, icon, children, onClick }: { active: boolean; compact?: boolean; icon: ReactNode; children: ReactNode; onClick: () => void }) {
  return <button type="button" onClick={onClick} className={cn("flex items-center gap-3 rounded-lg text-sm transition-colors", compact ? "flex-1 justify-center px-3 py-2" : "w-full px-3 py-2.5", active ? "bg-sky-400/10 text-sky-200" : "text-slate-400 hover:bg-white/[0.04] hover:text-slate-200")}>{icon}{children}</button>;
}

function ItemActions({ item, inTrash, onRename, onMove, onShare, onPreview, onDelete, onRestore, onPurge }: { item: DriveItem; inTrash: boolean; onRename: () => void; onMove: () => void; onShare: () => void; onPreview: () => void; onDelete: () => void; onRestore: () => void; onPurge: () => void }) {
  const [open, setOpen] = useState(false);
  const action = (callback: () => void) => { callback(); setOpen(false); };
  return (
    <div className="relative">
      <Button aria-label="更多操作" title="更多操作" size="icon" variant="ghost" onClick={() => setOpen((value) => !value)}><MoreHorizontal className="size-4" /></Button>
      {open ? <div className="absolute right-0 top-10 z-20 w-36 rounded-xl border border-white/10 bg-slate-800 p-1 shadow-xl shadow-black/30">
        {!inTrash && !driveItemIsDirectory(item) ? <MenuButton icon={<Download className="size-3.5" />} onClick={() => action(onPreview)}>预览与下载</MenuButton> : null}
        {!inTrash ? <><MenuButton icon={<Pencil className="size-3.5" />} onClick={() => action(onRename)}>重命名</MenuButton><MenuButton icon={<MoveRight className="size-3.5" />} onClick={() => action(onMove)}>移动</MenuButton><MenuButton icon={<Share2 className="size-3.5" />} onClick={() => action(onShare)}>创建分享</MenuButton><MenuButton destructive icon={<Trash2 className="size-3.5" />} onClick={() => action(onDelete)}>移入回收站</MenuButton></> : <><MenuButton icon={<ArchiveRestore className="size-3.5" />} onClick={() => action(onRestore)}>恢复</MenuButton><MenuButton destructive icon={<Trash2 className="size-3.5" />} onClick={() => action(onPurge)}>永久删除</MenuButton></>}
      </div> : null}
    </div>
  );
}

function MenuButton({ icon, children, destructive, onClick }: { icon: ReactNode; children: ReactNode; destructive?: boolean; onClick: () => void }) {
  return <button type="button" onClick={onClick} className={cn("flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-xs transition-colors hover:bg-white/[0.08]", destructive ? "text-rose-300" : "text-slate-200")}>{icon}{children}</button>;
}

function DestructiveConfirmationDialog({ busy, title, description, confirmLabel, onClose, onConfirm }: { busy: boolean; title: string; description: string; confirmLabel: string; onClose: () => void; onConfirm: () => void }) {
  const close = () => { if (!busy) onClose(); };
  return <Dialog open title={title} description={description} onClose={close}><div className="flex justify-end gap-2"><Button variant="ghost" disabled={busy} onClick={close}>取消</Button><Button variant="destructive" disabled={busy} onClick={onConfirm}>{busy ? "处理中" : confirmLabel}</Button></div></Dialog>;
}

function FolderDialog({ open, busy, onClose, onCreate }: { open: boolean; busy: boolean; onClose: () => void; onCreate: (name: string) => void }) {
  const [name, setName] = useState("");
  const submit = (event: FormEvent) => { event.preventDefault(); if (name.trim()) onCreate(name.trim()); };
  return <Dialog open={open} title="新建文件夹" description="文件夹会创建在当前目录中。" onClose={onClose}><form className="space-y-4" onSubmit={submit}><input autoFocus value={name} onChange={(event) => setName(event.target.value)} placeholder="文件夹名称" className="h-10 w-full rounded-lg border border-white/10 bg-slate-950 px-3 text-sm text-white outline-none focus:border-sky-300" /><div className="flex justify-end gap-2"><Button variant="ghost" onClick={onClose}>取消</Button><Button type="submit" disabled={busy || !name.trim()}>{busy ? "创建中" : "创建"}</Button></div></form></Dialog>;
}

function RenameDialog({ item, open, busy, onClose, onRename }: { item: DriveItem; open: boolean; busy: boolean; onClose: () => void; onRename: (name: string) => void }) {
  const [name, setName] = useState(driveItemName(item));
  const submit = (event: FormEvent) => { event.preventDefault(); if (name.trim()) onRename(name.trim()); };
  return <Dialog open={open} title="重命名" onClose={onClose}><form className="space-y-4" onSubmit={submit}><input autoFocus value={name} onChange={(event) => setName(event.target.value)} className="h-10 w-full rounded-lg border border-white/10 bg-slate-950 px-3 text-sm text-white outline-none focus:border-sky-300" /><div className="flex justify-end gap-2"><Button variant="ghost" onClick={onClose}>取消</Button><Button type="submit" disabled={busy || !name.trim()}>{busy ? "保存中" : "保存"}</Button></div></form></Dialog>;
}

function MoveDialog({ item, open, busy, currentDirectories, onClose, onMove }: { item: DriveItem; open: boolean; busy: boolean; currentDirectories: readonly DriveItem[]; onClose: () => void; onMove: (parentUid: string) => void }) {
  const [parentUid, setParentUid] = useState("");
  const submit = (event: FormEvent) => { event.preventDefault(); onMove(parentUid); };
  return <Dialog open={open} title="移动项目" description="选择当前可见目录，或将项目移至根目录。" onClose={onClose}><form className="space-y-4" onSubmit={submit}><label className="block text-sm text-slate-300">目标目录<select value={parentUid} onChange={(event) => setParentUid(event.target.value)} className="mt-2 h-10 w-full rounded-lg border border-white/10 bg-slate-950 px-3 text-sm text-white outline-none focus:border-sky-300"><option value="">根目录</option>{currentDirectories.filter((directory) => driveItemUid(directory) !== driveItemUid(item)).map((directory) => <option key={driveItemUid(directory)} value={driveItemUid(directory)}>{driveItemName(directory)}</option>)}</select></label><div className="flex justify-end gap-2"><Button variant="ghost" onClick={onClose}>取消</Button><Button type="submit" disabled={busy}>{busy ? "移动中" : "移动"}</Button></div></form></Dialog>;
}

function ShareDialog({ item, client, open, busy, onClose, onBusyChange }: { item: DriveItem; client: PrivateDriveClient; open: boolean; busy: boolean; onClose: () => void; onBusyChange: (busy: boolean) => void }) {
  const { showToast } = useToast();
  const [expiresOn, setExpiresOn] = useState(futureDate(7));
  const [created, setCreated] = useState<Readonly<{ shareLink: ShareLink; shareUrl: string }>>();
  const [links, setLinks] = useState<ShareLink[]>([]);
  const [linksError, setLinksError] = useState<string>();
  const [revokingUid, setRevokingUid] = useState<string>();

  useEffect(() => {
    if (!open) {
      return undefined;
    }
    const controller = new AbortController();
    setLinksError(undefined);
    void client.listShareLinks(driveItemUid(item), controller.signal).then(setLinks).catch((loadError) => {
      if (!controller.signal.aborted) {
        setLinksError(asErrorMessage(loadError));
      }
    });
    return () => controller.abort();
  }, [client, item, open]);

  const createShare = async (event: FormEvent) => {
    event.preventDefault();
    const selected = new Date(`${expiresOn}T23:59:59`);
    if (Number.isNaN(selected.getTime()) || selected.getTime() <= Date.now()) { showToast("请选择未来 30 天内的日期。", "error"); return; }
    const expiresAt = new Date(Math.min(selected.getTime(), Date.now() + 30 * 24 * 60 * 60 * 1_000));
    onBusyChange(true);
    try {
      const result = await client.createShareLink(driveItemUid(item), expiresAt);
      setCreated(result);
      setLinks((current) => [result.shareLink, ...current.filter((link) => shareLinkUid(link) !== shareLinkUid(result.shareLink))]);
      showToast("分享链接已创建");
    } catch (shareError) { showToast(asErrorMessage(shareError), "error"); } finally { onBusyChange(false); }
  };
  const copy = async () => { if (!created) return; await navigator.clipboard.writeText(created.shareUrl); showToast("链接已复制"); };
  const revoke = async (link: ShareLink) => {
    const uid = shareLinkUid(link);
    setRevokingUid(uid);
    try {
      const revoked = await client.revokeShareLink(uid);
      setLinks((current) => current.map((currentLink) => shareLinkUid(currentLink) === uid ? revoked : currentLink));
      showToast("分享已撤销");
    } catch (revokeError) { showToast(asErrorMessage(revokeError), "error"); } finally { setRevokingUid(undefined); }
  };
  return <Dialog open={open} title="创建分享链接" description={`“${driveItemName(item)}” 将以只读方式分享。`} size="standard" onClose={onClose}><div className="space-y-5">{created ? <div className="rounded-lg border border-sky-300/20 bg-sky-300/[0.06] p-3"><p className="text-xs text-sky-200">新链接（仅本次显示）</p><p className="mt-1 break-all text-sm text-sky-100">{created.shareUrl}</p><div className="mt-3 flex justify-end"><Button size="sm" onClick={() => void copy()}><Link2 className="size-3.5" />复制链接</Button></div></div> : null}<form className="space-y-4" onSubmit={(event) => void createShare(event)}><label className="block text-sm text-slate-300">失效日期<input type="date" min={futureDate(1)} max={futureDate(30)} value={expiresOn} onChange={(event) => setExpiresOn(event.target.value)} className="mt-2 h-10 w-full rounded-lg border border-white/10 bg-slate-950 px-3 text-sm text-white outline-none focus:border-sky-300" /></label><div className="flex justify-end gap-2"><Button variant="ghost" onClick={onClose}>关闭</Button><Button type="submit" disabled={busy}><Share2 className="size-4" />{busy ? "创建中" : "创建链接"}</Button></div></form><div className="border-t border-white/[0.08] pt-4"><p className="mb-2 text-sm font-medium text-slate-200">现有链接</p>{linksError ? <p className="text-xs text-rose-300">{linksError}</p> : links.length ? <div className="space-y-2">{links.map((link) => { const expired = (shareLinkExpiresAt(link)?.getTime() ?? Infinity) <= Date.now(); const revoked = Boolean(shareLinkRevokedAt(link)); return <div key={shareLinkUid(link)} className="flex items-center justify-between gap-3 rounded-lg border border-white/[0.08] px-3 py-2"><div className="min-w-0"><p className="text-xs text-slate-200">到期：{shareLinkExpiresAt(link)?.toLocaleString("zh-CN") ?? "—"}</p><p className="mt-0.5 text-[11px] text-slate-500">{revoked ? "已撤销" : expired ? "已过期" : "有效"}</p></div>{!revoked && !expired ? <Button size="sm" variant="destructive" disabled={revokingUid === shareLinkUid(link)} onClick={() => void revoke(link)}>{revokingUid === shareLinkUid(link) ? "撤销中" : "撤销"}</Button> : null}</div>; })}</div> : <p className="text-xs text-slate-500">尚未创建分享链接。</p>}</div></div></Dialog>;
}

function PreviewDialog({ item, client, open, onClose }: { item: DriveItem; client: PrivateDriveClient; open: boolean; onClose: () => void }) {
  const [url, setUrl] = useState("");
  const [text, setText] = useState<string>();
  const [error, setError] = useState<string>();
  useEffect(() => { const controller = new AbortController(); setText(undefined); setError(undefined); void client.getDownloadUrl(driveItemUid(item), controller.signal).then(setUrl).catch((loadError) => setError(asErrorMessage(loadError))); return () => controller.abort(); }, [client, item]);
  useEffect(() => { if (!url || !isText(item)) return; const controller = new AbortController(); void client.readText(url, controller.signal).then(setText).catch((loadError) => setError(asErrorMessage(loadError))); return () => controller.abort(); }, [client, item, url]);
  return <Dialog open={open} title={driveItemName(item)} size="preview" onClose={onClose}>{error ? <InlineError message={error} /> : null}<div className="min-h-48">{!url ? <LoadingIndicator label="正在准备预览" /> : isImage(item) ? <img src={url} referrerPolicy="no-referrer" alt={driveItemName(item)} className="max-h-[calc(100dvh-10rem)] w-full rounded-lg object-contain sm:max-h-[calc(90dvh-10rem)]" /> : isPdf(item) ? <iframe src={url} referrerPolicy="no-referrer" title={driveItemName(item)} className="h-[calc(100dvh-10rem)] min-h-80 w-full rounded-lg border border-white/10 bg-white sm:h-[calc(90dvh-10rem)]" /> : isText(item) ? text === undefined ? <LoadingIndicator label="正在载入文本" /> : <pre className="max-h-[calc(100dvh-10rem)] overflow-auto rounded-lg border border-white/10 bg-slate-950 p-3 text-xs leading-5 text-slate-300 sm:max-h-[calc(90dvh-10rem)]">{text}</pre> : <EmptyState title="此文件不支持在线预览" detail="你可以下载后在本地打开。" />}</div>{url ? <div className="mt-4 flex justify-end"><a href={asDownloadUrl(url)} referrerPolicy="no-referrer" className="inline-flex h-10 items-center gap-2 rounded-lg bg-sky-500 px-4 text-sm font-medium text-slate-950 hover:bg-sky-300"><Download className="size-4" />下载</a></div> : null}</Dialog>;
}

function UploadQueue({ tasks, onDismiss, onRetry }: { tasks: readonly UploadTask[]; onDismiss: (id: string) => void; onRetry: (id: string) => void }) {
  const visible = tasks.filter((task) => task.state !== "completed" || tasks.length <= 3).slice(-5);
  if (!visible.length) return null;
  return <section className="fixed bottom-4 right-4 z-30 w-[min(24rem,calc(100vw-2rem))] overflow-hidden rounded-xl border border-white/10 bg-slate-900 shadow-2xl shadow-black/40"><div className="flex items-center justify-between border-b border-white/[0.08] px-4 py-3"><p className="text-sm font-medium text-white">上传队列</p><span className="text-xs text-slate-500">{tasks.filter((task) => task.state === "uploading").length} 个进行中</span></div><div className="divide-y divide-white/[0.07]">{visible.map((task) => <div key={task.id} className="px-4 py-3"><div className="flex items-center gap-3"><FilePlus2 className="size-5 shrink-0 text-sky-300" /><div className="min-w-0 flex-1"><p className="truncate text-sm text-slate-200">{task.file.name}</p><p className="mt-0.5 text-xs text-slate-500">{task.state === "uploading" ? `${formatBytes(task.uploadedBytes)} / ${formatBytes(task.file.size)}` : task.state === "queued" ? "等待上传" : task.state === "completed" ? "已完成" : task.error}</p></div>{task.state === "failed" ? <Button size="sm" variant="outline" onClick={() => onRetry(task.id)}>重试</Button> : task.state === "completed" ? <Button size="icon" variant="ghost" aria-label="关闭上传记录" title="关闭上传记录" onClick={() => onDismiss(task.id)}><X className="size-4" /></Button> : <LoaderCircle className="size-4 animate-spin text-sky-300" />}</div>{task.state === "uploading" ? <progress className="mt-2 h-1.5 w-full overflow-hidden rounded-full accent-sky-400" max={task.file.size} value={task.uploadedBytes} /> : null}</div>)}</div></section>;
}
