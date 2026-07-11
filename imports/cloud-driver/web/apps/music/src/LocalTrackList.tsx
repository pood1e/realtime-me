import { useEffect, useRef } from "react";
import type { PlayableTrack, Track } from "@cloud-drive/contracts";
import { Button, LoadingIndicator, MusicClient } from "@cloud-drive/shared";
import { TrackRow } from "./TrackRow";

type LocalTrackListProps = Readonly<{
  tracks: Track[];
  queue: PlayableTrack[];
  client: MusicClient;
  current?: PlayableTrack;
  trashed: boolean;
  hasMore: boolean;
  loadingMore: boolean;
  loadMoreFailed: boolean;
  onLoadMore: () => Promise<void>;
  onPlay: (track: Track, queue: PlayableTrack[]) => void;
  onFavorite: (track: Track) => void;
  onRemove: (track: Track) => void;
  onRestore: (track: Track) => void;
}>;

export function LocalTrackList({
  tracks,
  queue,
  client,
  current,
  trashed,
  hasMore,
  loadingMore,
  loadMoreFailed,
  onLoadMore,
  onPlay,
  onFavorite,
  onRemove,
  onRestore,
}: LocalTrackListProps) {
  const sentinel = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const target = sentinel.current;
    if (
      !target ||
      !hasMore ||
      loadingMore ||
      loadMoreFailed ||
      !("IntersectionObserver" in window)
    )
      return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry?.isIntersecting) void onLoadMore();
      },
      { rootMargin: "480px 0px" },
    );
    observer.observe(target);
    return () => observer.disconnect();
  }, [hasMore, loadMoreFailed, loadingMore, onLoadMore]);

  return (
    <>
      <div className="overflow-hidden rounded-xl border bg-card/35">
        {tracks.map((track, index) => (
          <TrackRow
            key={track.uid}
            track={track}
            index={index}
            client={client}
            active={current?.trackId === track.uid}
            trashed={trashed}
            onPlay={() => onPlay(track, queue)}
            onFavorite={() => onFavorite(track)}
            onRemove={() => onRemove(track)}
            onRestore={() => onRestore(track)}
          />
        ))}
      </div>
      {hasMore ? (
        <div
          ref={sentinel}
          className="flex min-h-24 items-center justify-center"
          aria-live="polite"
        >
          {loadingMore ? (
            <LoadingIndicator label="继续加载本地音乐" />
          ) : (
            <Button variant="outline" onClick={() => void onLoadMore()}>
              {loadMoreFailed ? "重试加载" : "加载更多"}
            </Button>
          )}
        </div>
      ) : (
        <p className="sr-only" aria-live="polite">
          已加载全部本地音乐
        </p>
      )}
    </>
  );
}
