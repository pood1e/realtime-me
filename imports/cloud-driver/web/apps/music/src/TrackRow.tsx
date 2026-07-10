import { Heart, MoreHorizontal, Music2 } from "lucide-react";
import type { Track } from "@cloud-drive/contracts";
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  MusicClient,
  formatBytes,
} from "@cloud-drive/shared";

export function TrackRow({
  track,
  index,
  client,
  active,
  trashed,
  onPlay,
  onFavorite,
  onRemove,
  onRestore,
}: {
  track: Track;
  index: number;
  client: MusicClient;
  active: boolean;
  trashed: boolean;
  onPlay: () => void;
  onFavorite: () => void;
  onRemove: () => void;
  onRestore: () => void;
}) {
  const artwork = track.artworkUrl ? client.artworkUrl(track) : "";
  return (
    <div className={rowClassName(active)}>
      <button
        type="button"
        onClick={onPlay}
        className="grid size-9 place-items-center overflow-hidden rounded-md bg-muted text-xs text-muted-foreground"
      >
        {artwork ? (
          <img src={artwork} alt="" className="h-full w-full object-cover" />
        ) : active ? (
          <Music2 className="size-4 text-primary" />
        ) : (
          index + 1
        )}
      </button>
      <TrackSummary track={track} onPlay={onPlay} />
      <p className="hidden truncate text-sm text-muted-foreground md:block">
        {track.album || "—"}
      </p>
      <p className="hidden text-xs text-muted-foreground md:block">
        {formatBytes(track.sizeBytes)}
      </p>
      <TrackMenu
        track={track}
        trashed={trashed}
        onFavorite={onFavorite}
        onRemove={onRemove}
        onRestore={onRestore}
      />
    </div>
  );
}

function TrackSummary({ track, onPlay }: { track: Track; onPlay: () => void }) {
  return (
    <button type="button" onClick={onPlay} className="min-w-0 text-left">
      <p className="truncate text-sm font-medium">
        {track.title || track.originalFileName}
      </p>
      <p className="truncate text-xs text-muted-foreground">
        {track.artists.join("、") || "未知艺人"}
      </p>
    </button>
  );
}

function TrackMenu({
  track,
  trashed,
  onFavorite,
  onRemove,
  onRestore,
}: {
  track: Track;
  trashed: boolean;
  onFavorite: () => void;
  onRemove: () => void;
  onRestore: () => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon-sm">
          <MoreHorizontal />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {trashed ? (
          <DropdownMenuItem onSelect={onRestore}>恢复</DropdownMenuItem>
        ) : (
          <DropdownMenuItem onSelect={onFavorite}>
            <Heart
              className={track.favorite ? "fill-current text-rose-400" : ""}
            />
            {track.favorite ? "取消收藏" : "收藏"}
          </DropdownMenuItem>
        )}
        <DropdownMenuItem variant="destructive" onSelect={onRemove}>
          {trashed ? "永久删除" : "移入回收站"}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function rowClassName(active: boolean): string {
  const base =
    "grid grid-cols-[2.5rem_minmax(0,1fr)_3rem] items-center gap-3 border-b px-4 py-3 last:border-b-0 md:grid-cols-[3rem_minmax(0,1fr)_minmax(8rem,.6fr)_6rem_3rem]";
  return `${base} ${active ? "bg-primary/8" : "hover:bg-accent/35"}`;
}
