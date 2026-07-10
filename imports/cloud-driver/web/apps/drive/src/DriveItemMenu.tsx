import {
  ArchiveRestore,
  Download,
  MoreHorizontal,
  Share2,
  Trash2,
} from "lucide-react";
import type { DriveItem } from "@cloud-drive/contracts";
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  driveItemIsDirectory,
} from "@cloud-drive/shared";

export function DriveItemMenu({
  item,
  trashed,
  onPreview,
  onRename,
  onMove,
  onShare,
  onTrash,
  onRestore,
  onPurge,
}: {
  item: DriveItem;
  trashed: boolean;
  onPreview: () => void;
  onRename: () => void;
  onMove: () => void;
  onShare: () => void;
  onTrash: () => void;
  onRestore: () => void;
  onPurge: () => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon-sm">
          <MoreHorizontal />
          <span className="sr-only">文件操作</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {trashed ? (
          <TrashedActions onRestore={onRestore} onPurge={onPurge} />
        ) : (
          <ActiveActions
            directory={driveItemIsDirectory(item)}
            onPreview={onPreview}
            onRename={onRename}
            onMove={onMove}
            onShare={onShare}
            onTrash={onTrash}
          />
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function TrashedActions({
  onRestore,
  onPurge,
}: {
  onRestore: () => void;
  onPurge: () => void;
}) {
  return (
    <>
      <DropdownMenuItem onSelect={onRestore}>
        <ArchiveRestore />
        恢复
      </DropdownMenuItem>
      <DropdownMenuItem variant="destructive" onSelect={onPurge}>
        <Trash2 />
        永久删除
      </DropdownMenuItem>
    </>
  );
}

function ActiveActions({
  directory,
  onPreview,
  onRename,
  onMove,
  onShare,
  onTrash,
}: {
  directory: boolean;
  onPreview: () => void;
  onRename: () => void;
  onMove: () => void;
  onShare: () => void;
  onTrash: () => void;
}) {
  return (
    <>
      {!directory ? (
        <DropdownMenuItem onSelect={onPreview}>
          <Download />
          预览与下载
        </DropdownMenuItem>
      ) : null}
      <DropdownMenuItem onSelect={onRename}>重命名</DropdownMenuItem>
      <DropdownMenuItem onSelect={onMove}>移动</DropdownMenuItem>
      <DropdownMenuItem onSelect={onShare}>
        <Share2 />
        分享
      </DropdownMenuItem>
      <DropdownMenuSeparator />
      <DropdownMenuItem variant="destructive" onSelect={onTrash}>
        <Trash2 />
        移入回收站
      </DropdownMenuItem>
    </>
  );
}
