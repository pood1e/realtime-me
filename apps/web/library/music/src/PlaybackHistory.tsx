import { useEffect } from "react";
import { History } from "lucide-react";
import type { PlayableTrack, PlaybackEntry } from "@realtime-me/library-contracts";
import {
  EmptyState,
  InfiniteScrollSentinel,
  LoadingIndicator,
  MusicClient,
  useCursorQuery,
  useToast,
} from "@realtime-me/library-web";
import { PlayableTrackRow } from "./PlayableTrackRow";
import type { PlaybackQueueSelection } from "./playback/playback-types";

export function PlaybackHistory({
  client,
  current,
  refreshKey,
  onPlay,
  onLyrics,
}: {
  client: MusicClient;
  current: PlayableTrack | undefined;
  refreshKey: number;
  onPlay: (selection: PlaybackQueueSelection) => void;
  onLyrics: (track: PlayableTrack) => void;
}) {
  const { showToast } = useToast();
  const history = useCursorQuery({
    queryKey: ["music-history", refreshKey],
    loadPage: async (pageToken, signal) => {
      const page = await client.library.historyPage(pageToken, signal);
      return { items: page.entries, nextPageToken: page.nextPageToken };
    },
  });
  useEffect(() => {
    if (history.error) showToast(message(history.error), "error");
  }, [history.error, showToast]);
  if (history.isPending) return <LoadingIndicator label="正在读取播放历史" />;
  const entries = history.items.filter(
    (entry): entry is PlaybackEntry & { track: PlayableTrack } =>
      Boolean(entry.track),
  );
  const tracks = entries.map((entry) => entry.track);
  const loadPlaybackPage = async (pageToken: string, signal: AbortSignal) => {
    const page = await client.library.historyPage(pageToken, signal);
    return {
      tracks: playableHistoryTracks(page.entries),
      nextPageToken: page.nextPageToken,
    };
  };
  if (!tracks.length)
    return (
      <EmptyState
        icon={<History className="size-6" />}
        title="还没有播放记录"
        detail="歌曲开始播放后会记录在这里，不会向第三方平台回写。"
      />
    );
  return (
    <>
      <div className="overflow-hidden rounded-xl border bg-card/35">
        {tracks.map((track, index) => (
          <PlayableTrackRow
            key={`${entries[index]?.uid}-${track.providerId}-${track.trackId}`}
            track={track}
            index={index + 1}
            active={sameTrack(current, track)}
            client={client}
            onPlay={() =>
              onPlay({
                tracks,
                startIndex: index,
                nextPageToken: history.data?.pages.at(-1)?.nextPageToken ?? "",
                loadNextPage: loadPlaybackPage,
              })
            }
            onLyrics={() => onLyrics(track)}
          />
        ))}
      </div>
      <InfiniteScrollSentinel
        hasMore={history.hasNextPage}
        loading={history.isFetchingNextPage}
        failed={history.isFetchNextPageError}
        loadingLabel="继续加载播放历史"
        completeLabel="已加载全部播放历史"
        onLoadMore={() => void history.fetchNextPage()}
      />
    </>
  );
}

function sameTrack(a: PlayableTrack | undefined, b: PlayableTrack): boolean {
  return a?.providerId === b.providerId && a.trackId === b.trackId;
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : "播放历史读取失败";
}

function playableHistoryTracks(entries: PlaybackEntry[]): PlayableTrack[] {
  return entries.flatMap((entry) => (entry.track ? [entry.track] : []));
}
