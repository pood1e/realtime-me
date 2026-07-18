import type { DriveItem } from "@realtime-me/library-contracts";
import {
  Breadcrumbs,
  DriveClient,
  DriveItemView,
  DriveViewModeToggle,
  driveItemIsDirectory,
  driveItemName,
  driveItemUid,
  EmptyState,
  InfiniteScrollSentinel,
  InlineError,
  LoadingIndicator,
  UploadButton,
  UploadClient,
  useCursorQuery,
  useDialog,
  useDriveViewMode,
  useToast,
} from "@realtime-me/library-web";
import { ConsolePage } from "@realtime-me/web-shell";
import { Button, Input } from "@realtime-me/web-ui";
import { Folder, FolderPlus, Search, Trash2 } from "lucide-react";
import { useDeferredValue, useMemo, useState } from "react";
import { LIBRARY_API_BASE as API_BASE } from "@/config";
import { DriveItemMenu } from "./DriveItemMenu";
import { DriveDialog, type DriveDialogState } from "./drive-dialogs";

type View = "files" | "trash";
type Trail = Readonly<{ id: string; label: string }>;
const ROOT: readonly Trail[] = [{ id: "", label: "我的文件" }];

export function DrivePage() {
  const client = useMemo(() => new DriveClient(API_BASE), []);
  const uploader = useMemo(() => new UploadClient(API_BASE), []);
  const { showToast } = useToast();
  const { confirm } = useDialog();
  const [view, setView] = useState<View>("files");
  const [trail, setTrail] = useState<readonly Trail[]>(ROOT);
  const [query, setQuery] = useState("");
  const [dialog, setDialog] = useState<DriveDialogState>(null);
  const [mode, setMode] = useDriveViewMode("realtime-me.library.drive.view-mode");
  const parentUid = trail.at(-1)?.id ?? "";
  const deferredQuery = useDeferredValue(query.trim());
  const catalog = useCursorQuery({
    queryKey: ["drive-items", view, parentUid, deferredQuery],
    loadPage: async (pageToken, signal) => {
      if (view === "trash") return client.listTrashPage(pageToken, signal);
      if (deferredQuery) return client.searchPage(deferredQuery, pageToken, signal);
      return client.listPage(parentUid, pageToken, signal);
    },
  });
  const items = catalog.items;
  const load = async () => {
    await catalog.refetch();
  };

  const upload = async (files: File[]) => {
    for (const file of files) {
      try {
        const uploadUid = await uploader.upload(file);
        await client.importUpload(uploadUid, parentUid, file.name);
        showToast(`${file.name} 已上传`);
      } catch (uploadError) {
        showToast(`${file.name}: ${message(uploadError)}`, "error");
      }
    }
    await load();
  };

  const open = (item: DriveItem) => {
    if (driveItemIsDirectory(item)) {
      setTrail((current) => [...current, { id: driveItemUid(item), label: driveItemName(item) }]);
      setQuery("");
    } else setDialog({ kind: "preview", item });
  };

  const trash = async (item: DriveItem) => {
    if (
      !(await confirm({
        title: "移入回收站",
        description: `将“${driveItemName(item)}”移入回收站？`,
        confirmLabel: "移入回收站",
      }))
    )
      return;
    try {
      await client.trash(driveItemUid(item));
      showToast("已移入回收站");
      await load();
    } catch (actionError) {
      showToast(message(actionError), "error");
    }
  };
  const restore = async (item: DriveItem) => {
    try {
      await client.restore(driveItemUid(item));
      showToast("已恢复");
      await load();
    } catch (actionError) {
      showToast(message(actionError), "error");
    }
  };
  const purge = async (item: DriveItem) => {
    if (
      !(await confirm({
        title: "永久删除文件",
        description: "此操作无法撤销。",
        confirmLabel: "永久删除",
        destructive: true,
      }))
    )
      return;
    try {
      await client.purge(driveItemUid(item));
      showToast("已永久删除");
      await load();
    } catch (actionError) {
      showToast(message(actionError), "error");
    }
  };
  const emptyTrash = async () => {
    if (
      !(await confirm({
        title: "清空回收站",
        description: "回收站中的全部内容将被永久删除，此操作无法撤销。",
        confirmLabel: "永久删除全部内容",
        destructive: true,
      }))
    )
      return;
    try {
      await client.emptyTrash();
      showToast("回收站已清空");
      await load();
    } catch (error) {
      showToast(message(error), "error");
    }
  };

  const actions = (item: DriveItem) => (
    <DriveItemMenu
      item={item}
      trashed={view === "trash"}
      onPreview={() => setDialog({ kind: "preview", item })}
      onRename={() => setDialog({ kind: "rename", item })}
      onMove={() => setDialog({ kind: "move", item })}
      onShare={() => setDialog({ kind: "share", item })}
      onTrash={() => void trash(item)}
      onRestore={() => void restore(item)}
      onPurge={() => void purge(item)}
    />
  );

  return (
    <ConsolePage
      title="云盘"
      subtitle="通用文件管理"
      actions={
        view === "files" ? (
          <UploadButton onFiles={upload} />
        ) : (
          <Button variant="destructive" onClick={() => void emptyTrash()}>
            清空回收站
          </Button>
        )
      }
    >
      <div className="mb-6 flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
        <div className="flex items-center gap-2">
          <Button
            variant={view === "files" ? "secondary" : "ghost"}
            onClick={() => {
              setView("files");
              setTrail(ROOT);
            }}
          >
            文件
          </Button>
          <Button
            variant={view === "trash" ? "secondary" : "ghost"}
            onClick={() => setView("trash")}
          >
            <Trash2 />
            回收站
          </Button>
          {view === "files" ? (
            <Button variant="outline" onClick={() => setDialog({ kind: "folder" })}>
              <FolderPlus />
              新建文件夹
            </Button>
          ) : null}
        </div>
        <div className="flex min-w-0 items-center gap-2">
          <div className="relative min-w-0 flex-1 xl:w-80">
            <Search className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="搜索文件"
              className="pl-9"
              disabled={view === "trash"}
            />
          </div>
          <DriveViewModeToggle mode={mode} onChange={setMode} />
        </div>
      </div>
      {view === "files" && !query ? (
        <div className="mb-5">
          <Breadcrumbs
            items={trail}
            onNavigate={(id) =>
              setTrail((current) =>
                current.slice(0, current.findIndex((part) => part.id === id) + 1),
              )
            }
          />
        </div>
      ) : null}
      {catalog.error ? (
        <InlineError message={message(catalog.error)} onRetry={() => void load()} />
      ) : catalog.isPending ? (
        <LoadingIndicator label="正在读取文件" />
      ) : (
        <>
          <DriveItemView
            mode={mode}
            items={items}
            onOpen={open}
            actions={actions}
            empty={
              <EmptyState
                icon={<Folder className="size-6" />}
                title={view === "trash" ? "回收站为空" : query ? "没有匹配文件" : "这里还没有文件"}
              />
            }
          />
          <InfiniteScrollSentinel
            hasMore={catalog.hasNextPage}
            loading={catalog.isFetchingNextPage}
            failed={catalog.isFetchNextPageError}
            loadingLabel="继续加载文件"
            completeLabel="已加载全部文件"
            onLoadMore={() => void catalog.fetchNextPage()}
          />
        </>
      )}
      {dialog ? (
        <DriveDialog
          state={dialog}
          client={client}
          currentItems={items}
          parentUid={parentUid}
          onClose={() => setDialog(null)}
          onChanged={async () => {
            setDialog(null);
            await load();
          }}
        />
      ) : null}
    </ConsolePage>
  );
}

function message(error: unknown) {
  return error instanceof Error ? error.message : "操作未完成。";
}
