import { ExternalLink, FileText, Music2, Play } from "lucide-react";
import type { PlayableTrack } from "@realtime-me/library-contracts";
import { Badge, Button, MusicClient } from "@realtime-me/library-web";
import { clock, durationSeconds } from "./music-model";
import { useProviderLabel } from "./provider-catalog";

export function PlayableTrackRow({
  track,
  index,
  active,
  client,
  onPlay,
  onLyrics,
}: {
  track: PlayableTrack;
  index?: number;
  active: boolean;
  client: MusicClient;
  onPlay: () => void;
  onLyrics?: () => void;
}) {
  const providerLabel = useProviderLabel();
  const artwork = client.providers.artworkUrl(track);
  return (
    <div className={rowClassName(active)}>
      <button
        type="button"
        onClick={onPlay}
        disabled={!track.playable}
        className="group grid size-10 place-items-center overflow-hidden rounded-md bg-muted text-xs text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50"
      >
        {artwork ? (
          <span className="relative h-full w-full">
            <img
              src={artwork}
              alt=""
              loading="lazy"
              decoding="async"
              className="h-full w-full object-cover"
            />
            <span className="absolute inset-0 hidden place-items-center bg-black/45 text-white group-hover:grid">
              <Play className="size-4 fill-current" />
            </span>
          </span>
        ) : active ? (
          <Music2 className="size-4 text-primary" />
        ) : (
          (index ?? <Play className="size-4" />)
        )}
      </button>
      <button
        type="button"
        onClick={onPlay}
        disabled={!track.playable}
        className="min-w-0 text-left disabled:cursor-not-allowed"
      >
        <p className="truncate text-sm font-medium">{track.title}</p>
        <p className="truncate text-xs text-muted-foreground">
          {track.artists.join("、") || "未知艺人"}
        </p>
      </button>
      <p className="hidden truncate text-sm text-muted-foreground md:block">
        {track.album || "—"}
      </p>
      <Badge variant="outline" className="hidden lg:inline-flex">
        {providerLabel(track.providerId)}
      </Badge>
      <span className="hidden text-xs text-muted-foreground sm:block">
        {clock(durationSeconds(track))}
      </span>
      <div className="flex justify-end gap-1">
        {onLyrics && track.lyricsAvailable ? (
          <Button variant="ghost" size="icon-sm" onClick={onLyrics}>
            <FileText />
            <span className="sr-only">查看歌词</span>
          </Button>
        ) : null}
        {track.providerUrl ? (
          <Button variant="ghost" size="icon-sm" asChild>
            <a href={track.providerUrl} target="_blank" rel="noreferrer">
              <ExternalLink />
              <span className="sr-only">在来源中打开</span>
            </a>
          </Button>
        ) : null}
      </div>
    </div>
  );
}

function rowClassName(active: boolean): string {
  const base =
    "grid grid-cols-[2.5rem_minmax(0,1fr)_auto] items-center gap-3 border-b px-4 py-3 last:border-b-0 sm:grid-cols-[2.5rem_minmax(0,1fr)_3.5rem_auto] md:grid-cols-[2.5rem_minmax(0,1fr)_minmax(8rem,.65fr)_3.5rem_auto] lg:grid-cols-[2.5rem_minmax(0,1fr)_minmax(8rem,.65fr)_7rem_3.5rem_auto]";
  return `${base} ${active ? "bg-primary/8" : "hover:bg-accent/35"}`;
}
