import type { ReactNode } from "react";
import {
  File,
  FileArchive,
  FileCode2,
  FileImage,
  FileText,
  Folder,
  Music2,
  Video,
} from "lucide-react";
import type { DriveItem } from "@cloud-drive/contracts";

import { fileExtension, formatBytes, formatDate } from "../format";
import { cn } from "../lib/utils";
import {
  driveItemContentType,
  driveItemIsDirectory,
  driveItemName,
  driveItemSize,
  driveItemUid,
  driveItemUpdatedAt,
} from "../message";
import type { DriveViewMode } from "./navigation";

export function FileGlyph({
  item,
  className,
}: {
  item: DriveItem;
  className?: string;
}) {
  const contentType = driveItemContentType(item);
  const extension = fileExtension(driveItemName(item));
  const iconClass = cn("size-5", className);
  if (driveItemIsDirectory(item))
    return (
      <Folder
        className={cn(iconClass, "fill-amber-300/15 text-amber-300")}
        aria-hidden="true"
      />
    );
  if (
    contentType.startsWith("image/") ||
    ["avif", "gif", "jpg", "jpeg", "png", "svg", "webp"].includes(extension)
  )
    return (
      <FileImage
        className={cn(iconClass, "text-fuchsia-300")}
        aria-hidden="true"
      />
    );
  if (contentType.startsWith("audio/"))
    return (
      <Music2 className={cn(iconClass, "text-violet-300")} aria-hidden="true" />
    );
  if (contentType.startsWith("video/"))
    return (
      <Video className={cn(iconClass, "text-rose-300")} aria-hidden="true" />
    );
  if (["zip", "gz", "rar", "7z", "tar"].includes(extension))
    return (
      <FileArchive
        className={cn(iconClass, "text-amber-200")}
        aria-hidden="true"
      />
    );
  if (
    contentType.startsWith("text/") ||
    [
      "css",
      "go",
      "html",
      "js",
      "json",
      "md",
      "py",
      "rs",
      "sql",
      "ts",
      "tsx",
      "xml",
      "yaml",
      "yml",
    ].includes(extension)
  )
    return (
      <FileCode2 className={cn(iconClass, "text-sky-300")} aria-hidden="true" />
    );
  if (extension === "pdf" || ["doc", "docx", "rtf", "txt"].includes(extension))
    return (
      <FileText
        className={cn(iconClass, "text-emerald-300")}
        aria-hidden="true"
      />
    );
  return (
    <File
      className={cn(iconClass, "text-muted-foreground")}
      aria-hidden="true"
    />
  );
}

type DriveItemCollectionProps = {
  items: readonly DriveItem[];
  selectedUid?: string;
  onOpen: (item: DriveItem) => void;
  actions?: (item: DriveItem) => ReactNode;
};

function DriveItemList({
  items,
  selectedUid,
  onOpen,
  actions,
}: DriveItemCollectionProps) {
  return (
    <div className="relative isolate rounded-xl border bg-card/60">
      <div className="sticky top-0 z-10 hidden grid-cols-[minmax(0,1fr)_7rem_10rem_3rem] gap-4 rounded-t-xl border-b bg-card/95 px-4 py-2.5 text-[11px] font-medium tracking-[0.14em] text-muted-foreground uppercase backdrop-blur sm:grid xl:grid-cols-[minmax(20rem,1fr)_9rem_12rem_3rem]">
        <span>名称</span>
        <span>大小</span>
        <span>修改时间</span>
        <span />
      </div>
      <div className="divide-y">
        {items.map((item) => (
          <DriveListRow
            key={driveItemUid(item)}
            item={item}
            selected={selectedUid === driveItemUid(item)}
            onOpen={onOpen}
            {...(actions ? { actions } : {})}
          />
        ))}
      </div>
    </div>
  );
}

function DriveListRow({
  item,
  selected,
  onOpen,
  actions,
}: {
  item: DriveItem;
  selected: boolean;
  onOpen: (item: DriveItem) => void;
  actions?: (item: DriveItem) => ReactNode;
}) {
  const directory = driveItemIsDirectory(item);
  return (
    <div
      className={cn(
        "group grid grid-cols-[minmax(0,1fr)_3rem] items-stretch transition-colors hover:bg-accent/50",
        selected && "bg-accent",
      )}
    >
      <button
        type="button"
        onClick={() => onOpen(item)}
        className="grid min-w-0 grid-cols-[minmax(0,1fr)] items-center gap-3 px-4 py-3 text-left outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring sm:grid-cols-[minmax(0,1fr)_7rem_10rem] sm:gap-4 xl:grid-cols-[minmax(20rem,1fr)_9rem_12rem]"
      >
        <div className="flex min-w-0 items-center gap-3">
          <FileGlyph item={item} />
          <div className="min-w-0">
            <p
              className="truncate text-sm font-medium"
              title={driveItemName(item)}
            >
              {driveItemName(item)}
            </p>
            <p className="mt-0.5 text-xs text-muted-foreground sm:hidden">
              {directory
                ? "文件夹"
                : `${formatBytes(driveItemSize(item))} · ${formatDate(driveItemUpdatedAt(item))}`}
            </p>
          </div>
        </div>
        <span className="hidden text-sm text-muted-foreground sm:block">
          {directory ? "—" : formatBytes(driveItemSize(item))}
        </span>
        <span className="hidden text-sm text-muted-foreground sm:block">
          {formatDate(driveItemUpdatedAt(item))}
        </span>
      </button>
      <div className="flex items-center justify-center">
        {actions ? actions(item) : null}
      </div>
    </div>
  );
}

function DriveItemGrid({
  items,
  selectedUid,
  onOpen,
  actions,
}: DriveItemCollectionProps) {
  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-[repeat(auto-fill,minmax(11rem,1fr))]">
      {items.map((item) => (
        <article
          key={driveItemUid(item)}
          className={cn(
            "group relative overflow-hidden rounded-xl border bg-card/60 transition-colors hover:bg-accent/50",
            selectedUid === driveItemUid(item) && "border-primary/50 bg-accent",
          )}
        >
          <button
            type="button"
            onClick={() => onOpen(item)}
            className="flex min-h-40 w-full flex-col p-4 text-left outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
          >
            <div className="grid size-11 place-items-center rounded-xl bg-muted">
              <FileGlyph item={item} className="size-7" />
            </div>
            <div className="mt-5 min-w-0 max-w-[calc(100%-2rem)]">
              <p
                className="truncate text-sm font-medium"
                title={driveItemName(item)}
              >
                {driveItemName(item)}
              </p>
              <p className="mt-1 text-xs text-muted-foreground">
                {driveItemIsDirectory(item)
                  ? "文件夹"
                  : formatBytes(driveItemSize(item))}
              </p>
              <p className="mt-1 truncate text-xs text-muted-foreground/70">
                {formatDate(driveItemUpdatedAt(item))}
              </p>
            </div>
          </button>
          {actions ? (
            <div className="absolute top-3 right-3">{actions(item)}</div>
          ) : null}
        </article>
      ))}
    </div>
  );
}

export function DriveItemView({
  mode,
  empty,
  ...props
}: DriveItemCollectionProps & { mode: DriveViewMode; empty: ReactNode }) {
  if (!props.items.length) return <>{empty}</>;
  return mode === "grid" ? (
    <DriveItemGrid {...props} />
  ) : (
    <DriveItemList {...props} />
  );
}
