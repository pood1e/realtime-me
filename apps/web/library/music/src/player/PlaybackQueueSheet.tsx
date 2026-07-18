import { useState } from "react";
import { ListMusic } from "lucide-react";
import {
  Button,
  InfiniteScrollSentinel,
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
  Tooltip,
  TooltipContent,
  TooltipTrigger,
  type MusicClient,
} from "@realtime-me/library-web";
import type { PlaybackQueueController } from "../playback/playback-queue";
import { useProviderLabel } from "../provider-catalog";

export function PlaybackQueueSheet({
  client,
  queue,
}: {
  client: MusicClient;
  queue: PlaybackQueueController;
}) {
  const [open, setOpen] = useState(false);
  const providerLabel = useProviderLabel();
  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <Tooltip>
        <TooltipTrigger asChild>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon-sm" aria-label="播放队列">
              <ListMusic />
            </Button>
          </SheetTrigger>
        </TooltipTrigger>
        <TooltipContent side="top" sideOffset={8}>
          播放队列
        </TooltipContent>
      </Tooltip>
      <SheetContent className="w-[min(92vw,28rem)] sm:max-w-md">
        <SheetHeader>
          <SheetTitle>播放队列</SheetTitle>
          <SheetDescription>{queue.tracks.length} 首歌曲</SheetDescription>
        </SheetHeader>
        <div className="min-h-0 flex-1 overflow-y-auto px-3 pb-4">
          <div className="overflow-hidden rounded-lg border">
            {queue.tracks.map((track, index) => {
              const artwork = client.providers.artworkUrl(track);
              const active = index === queue.currentIndex;
              return (
                <button
                  key={`${track.providerId}-${track.trackId}-${index}`}
                  type="button"
                  aria-current={active ? "true" : undefined}
                  className={`flex w-full items-center gap-3 border-b px-3 py-2.5 text-left last:border-b-0 ${
                    active ? "bg-primary/10" : "hover:bg-accent/45"
                  }`}
                  onClick={() => {
                    queue.playIndex(index);
                    setOpen(false);
                  }}
                >
                  {artwork ? (
                    <img
                      src={artwork}
                      alt=""
                      loading="lazy"
                      decoding="async"
                      className="size-9 shrink-0 rounded object-cover"
                    />
                  ) : (
                    <div className="size-9 shrink-0 rounded bg-muted" />
                  )}
                  <span className="min-w-0 flex-1">
                    <span className="block truncate text-sm font-medium">
                      {track.title}
                    </span>
                    <span className="block truncate text-xs text-muted-foreground">
                      {track.artists.join("、") || "未知艺人"} ·{" "}
                      {providerLabel(track.providerId)}
                    </span>
                  </span>
                  <span className="text-xs text-muted-foreground">
                    {index + 1}
                  </span>
                </button>
              );
            })}
          </div>
          <InfiniteScrollSentinel
            hasMore={queue.hasMore}
            loading={queue.loadingMore}
            failed={Boolean(queue.loadError)}
            loadingLabel="继续加载播放队列"
            completeLabel="已加载全部播放队列"
            onLoadMore={() => void queue.loadMore()}
          />
          {queue.loadError ? (
            <p className="px-3 pb-3 text-center text-xs text-destructive">
              {queue.loadError}
            </p>
          ) : null}
        </div>
      </SheetContent>
    </Sheet>
  );
}
