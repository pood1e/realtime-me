import type { FormEvent } from "react";
import { useEffect, useState } from "react";
import { Download, Link2 } from "lucide-react";
import type { DriveItem } from "@realtime-me/library-contracts";
import {
  Button,
  AppDialog,
  DriveClient,
  EmptyState,
  InlineError,
  Input,
  LoadingIndicator,
  driveItemIsDirectory,
  driveItemName,
  driveItemUid,
  isImage,
  isPdf,
  isText,
  useToast,
} from "@realtime-me/library-web";

export type DriveDialogState =
  | { kind: "folder" }
  | { kind: "rename" | "move" | "share" | "preview"; item: DriveItem }
  | null;

export function DriveDialog({
  state,
  client,
  currentItems,
  parentUid,
  onClose,
  onChanged,
}: {
  state: Exclude<DriveDialogState, null>;
  client: DriveClient;
  currentItems: readonly DriveItem[];
  parentUid: string;
  onClose: () => void;
  onChanged: () => Promise<void>;
}) {
  if (state.kind === "preview")
    return (
      <PreviewDialog item={state.item} client={client} onClose={onClose} />
    );
  if (state.kind === "share")
    return <ShareDialog item={state.item} client={client} onClose={onClose} />;
  return (
    <EditDialog
      state={state}
      client={client}
      currentItems={currentItems}
      parentUid={parentUid}
      onClose={onClose}
      onChanged={onChanged}
    />
  );
}

function EditDialog({
  state,
  client,
  currentItems,
  parentUid,
  onClose,
  onChanged,
}: {
  state: Exclude<
    DriveDialogState,
    null | { kind: "preview" } | { kind: "share" }
  >;
  client: DriveClient;
  currentItems: readonly DriveItem[];
  parentUid: string;
  onClose: () => void;
  onChanged: () => Promise<void>;
}) {
  const { showToast } = useToast();
  const [value, setValue] = useState(
    state.kind === "rename" ? driveItemName(state.item) : "",
  );
  const [busy, setBusy] = useState(false);
  const submit = async (event: FormEvent) => {
    event.preventDefault();
    setBusy(true);
    try {
      if (state.kind === "folder")
        await client.createDirectory(parentUid, value.trim());
      else if (state.kind === "rename")
        await client.rename(driveItemUid(state.item), value.trim());
      else await client.move(driveItemUid(state.item), value);
      showToast("操作已完成");
      await onChanged();
    } catch (error) {
      showToast(error instanceof Error ? error.message : "操作未完成", "error");
      setBusy(false);
    }
  };
  const title =
    state.kind === "folder"
      ? "新建文件夹"
      : state.kind === "rename"
        ? "重命名"
        : "移动项目";
  return (
    <AppDialog open title={title} onClose={onClose}>
      <form className="space-y-4" onSubmit={(event) => void submit(event)}>
        {state.kind === "move" ? (
          <label className="block text-sm">
            目标目录
            <select
              value={value}
              onChange={(event) => setValue(event.target.value)}
              className="mt-2 h-10 w-full rounded-md border bg-background px-3"
            >
              <option value="">根目录</option>
              {currentItems
                .filter(
                  (item) =>
                    driveItemIsDirectory(item) &&
                    driveItemUid(item) !== driveItemUid(state.item),
                )
                .map((item) => (
                  <option key={driveItemUid(item)} value={driveItemUid(item)}>
                    {driveItemName(item)}
                  </option>
                ))}
            </select>
          </label>
        ) : (
          <Input
            autoFocus
            value={value}
            onChange={(event) => setValue(event.target.value)}
            placeholder="名称"
          />
        )}
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={onClose}>
            取消
          </Button>
          <Button
            type="submit"
            disabled={busy || (state.kind !== "move" && !value.trim())}
          >
            {busy ? "处理中" : "确定"}
          </Button>
        </div>
      </form>
    </AppDialog>
  );
}

function ShareDialog({
  item,
  client,
  onClose,
}: {
  item: DriveItem;
  client: DriveClient;
  onClose: () => void;
}) {
  const { showToast } = useToast();
  const [url, setUrl] = useState("");
  const [busy, setBusy] = useState(false);
  const create = async () => {
    setBusy(true);
    try {
      const result = await client.createShare(
        driveItemUid(item),
        new Date(Date.now() + 7 * 86_400_000),
      );
      setUrl(result.shareUrl);
    } catch (error) {
      showToast(error instanceof Error ? error.message : "创建失败", "error");
    } finally {
      setBusy(false);
    }
  };
  const copy = async () => {
    await navigator.clipboard.writeText(url);
    showToast("链接已复制");
  };
  return (
    <AppDialog
      open
      title="分享文件"
      description="链接默认 7 天后失效。"
      onClose={onClose}
    >
      {url ? (
        <div className="space-y-4">
          <p className="break-all rounded-lg border bg-muted/30 p-3 text-sm">
            {url}
          </p>
          <Button className="w-full" onClick={() => void copy()}>
            <Link2 />
            复制链接
          </Button>
        </div>
      ) : (
        <Button
          className="w-full"
          disabled={busy}
          onClick={() => void create()}
        >
          {busy ? "创建中" : "创建分享链接"}
        </Button>
      )}
    </AppDialog>
  );
}

function PreviewDialog({
  item,
  client,
  onClose,
}: {
  item: DriveItem;
  client: DriveClient;
  onClose: () => void;
}) {
  const [url, setUrl] = useState("");
  const [text, setText] = useState<string>();
  const [error, setError] = useState("");
  useEffect(() => {
    let active = true;
    void client
      .downloadUrl(driveItemUid(item))
      .then((next) => {
        if (active) setUrl(next);
      })
      .catch((reason) => {
        if (active)
          setError(reason instanceof Error ? reason.message : "无法预览");
      });
    return () => {
      active = false;
    };
  }, [client, item]);
  useEffect(() => {
    if (!url || !isText(item)) return;
    const controller = new AbortController();
    void fetch(url, { credentials: "include", signal: controller.signal })
      .then((response) => response.text())
      .then(setText)
      .catch((reason) => {
        if (!controller.signal.aborted) setError(String(reason));
      });
    return () => controller.abort();
  }, [item, url]);
  const download = url
    ? `${url}${url.includes("?") ? "&" : "?"}download=1`
    : "";
  return (
    <AppDialog
      open
      title={driveItemName(item)}
      size="preview"
      onClose={onClose}
    >
      {error ? (
        <InlineError message={error} />
      ) : !url ? (
        <LoadingIndicator label="正在准备预览" />
      ) : (
        <div className="space-y-4">
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
              <pre className="whitespace-pre-wrap p-3 text-xs">{text}</pre>
            ) : (
              <EmptyState title="不支持在线预览" />
            )}
          </div>
          <div className="flex justify-end">
            <Button asChild>
              <a href={download}>
                <Download />
                下载
              </a>
            </Button>
          </div>
        </div>
      )}
    </AppDialog>
  );
}
