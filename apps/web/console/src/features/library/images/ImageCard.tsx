import type { Image } from "@realtime-me/library-contracts";
import { formatBytes, type ImagesClient } from "@realtime-me/library-web";
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@realtime-me/web-ui";
import { ImageIcon, MoreHorizontal } from "lucide-react";

export function ImageCard({
  image,
  client,
  trashed,
  onOpen,
  onRemove,
  onRestore,
}: {
  image: Image;
  client: ImagesClient;
  trashed: boolean;
  onOpen: () => void;
  onRemove: () => void;
  onRestore: () => void;
}) {
  return (
    <article className="group relative mb-3 break-inside-avoid overflow-hidden rounded-xl border bg-card">
      <button type="button" onClick={onOpen} disabled={trashed} className="block w-full bg-muted">
        {image.previewUrl ? (
          <img
            src={client.previewUrl(image)}
            alt={image.displayName}
            loading="lazy"
            decoding="async"
            className="h-auto w-full object-cover transition-transform duration-300 group-hover:scale-[1.02]"
          />
        ) : (
          <div className="grid aspect-square place-items-center">
            <ImageIcon className="size-8 text-muted-foreground" />
          </div>
        )}
      </button>
      <div className="flex items-center gap-2 p-3">
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium">{image.displayName}</p>
          <p className="mt-0.5 text-[11px] text-muted-foreground">
            {image.width} × {image.height} · {formatBytes(image.sizeBytes)}
          </p>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon-xs" aria-label="图片操作">
              <MoreHorizontal />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {trashed ? (
              <DropdownMenuItem onSelect={onRestore}>恢复</DropdownMenuItem>
            ) : (
              <DropdownMenuItem onSelect={onOpen}>打开与分享</DropdownMenuItem>
            )}
            <DropdownMenuItem variant="destructive" onSelect={onRemove}>
              {trashed ? "永久删除" : "移入回收站"}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </article>
  );
}
