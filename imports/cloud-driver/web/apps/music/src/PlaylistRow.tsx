import {
  ChevronDown,
  ChevronRight,
  CloudDownload,
  ExternalLink,
  Trash2,
} from "lucide-react";
import type { Playlist } from "@cloud-drive/contracts";
import { Badge, Button, Progress } from "@cloud-drive/shared";
import { providerLabel } from "./music-model";

export function PlaylistRow({
  playlist,
  expanded,
  busy,
  onToggle,
  onDownload,
  onDelete,
}: {
  playlist: Playlist;
  expanded: boolean;
  busy: boolean;
  onToggle: () => void;
  onDownload: () => void;
  onDelete: () => void;
}) {
  const progress = playlist.downloadableTrackCount
    ? (playlist.completedTrackCount / playlist.downloadableTrackCount) * 100
    : 0;
  const fullyDownloaded =
    playlist.downloadableTrackCount === 0 ||
    playlist.completedTrackCount >= playlist.downloadableTrackCount;
  return (
    <div>
      <div className="flex items-center gap-3 px-4 py-4">
        <button
          type="button"
          onClick={onToggle}
          className="grid size-8 shrink-0 place-items-center rounded-md hover:bg-accent"
          aria-label={expanded ? "收起歌单" : "展开歌单"}
        >
          {expanded ? <ChevronDown /> : <ChevronRight />}
        </button>
        {playlist.artworkUrl ? (
          <img
            src={playlist.artworkUrl}
            alt=""
            className="size-14 shrink-0 rounded-md object-cover"
          />
        ) : (
          <div className="size-14 shrink-0 rounded-md bg-muted" />
        )}
        <button
          type="button"
          onClick={onToggle}
          className="min-w-0 flex-1 text-left"
        >
          <div className="flex min-w-0 items-center gap-2">
            <p className="truncate text-sm font-medium">
              {playlist.displayName}
            </p>
            <Badge variant="outline" className="shrink-0">
              {providerLabel(playlist.provider)}
            </Badge>
          </div>
          <p className="mt-1 text-xs text-muted-foreground">
            {downloadSummary(playlist)}
          </p>
          {playlist.downloadSupported && playlist.trackCount ? (
            <Progress value={progress} className="mt-2 h-1.5 max-w-md" />
          ) : null}
        </button>
        <div className="flex shrink-0 items-center gap-1">
          {playlist.providerUrl ? (
            <Button variant="ghost" size="icon-sm" asChild>
              <a
                href={playlist.providerUrl}
                target="_blank"
                rel="noreferrer"
                title="在来源中打开"
              >
                <ExternalLink />
                <span className="sr-only">在来源中打开</span>
              </a>
            </Button>
          ) : null}
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={onDownload}
            disabled={busy || !playlist.downloadSupported || fullyDownloaded}
            title={
              playlist.downloadSupported
                ? "将未下载歌曲存入本地"
                : "该来源只支持在线播放"
            }
          >
            <CloudDownload />
            <span className="sr-only">存入本地</span>
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={onDelete}
            disabled={busy}
            title="移除歌单"
          >
            <Trash2 />
            <span className="sr-only">移除歌单</span>
          </Button>
        </div>
      </div>
    </div>
  );
}

function downloadSummary(playlist: Playlist): string {
  if (!playlist.downloadSupported) {
    return `${playlist.trackCount} 首 · 仅支持在线播放`;
  }
  const details = [
    `${playlist.completedTrackCount}/${playlist.downloadableTrackCount} 可下载歌曲已存入本地`,
  ];
  if (playlist.pendingTrackCount)
    details.push(`${playlist.pendingTrackCount} 首处理中`);
  if (playlist.failedTrackCount)
    details.push(`${playlist.failedTrackCount} 首失败`);
  return details.join(" · ");
}
