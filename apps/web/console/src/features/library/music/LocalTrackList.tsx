import type { PlayableTrack, Track } from "@realtime-me/library-contracts";
import { InfiniteScrollSentinel, type MusicClient } from "@realtime-me/library-web";
import { TrackRow } from "./TrackRow";

type LocalTrackListProps = Readonly<{
  tracks: Track[];
  client: MusicClient;
  current: PlayableTrack | undefined;
  trashed: boolean;
  hasMore: boolean;
  loadingMore: boolean;
  loadMoreFailed: boolean;
  onLoadMore: () => Promise<void>;
  onPlay: (index: number) => void;
  onFavorite: (track: Track) => void;
  onRemove: (track: Track) => void;
  onRestore: (track: Track) => void;
}>;

export function LocalTrackList({
  tracks,
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
            onPlay={() => onPlay(index)}
            onFavorite={() => onFavorite(track)}
            onRemove={() => onRemove(track)}
            onRestore={() => onRestore(track)}
          />
        ))}
      </div>
      <InfiniteScrollSentinel
        hasMore={hasMore}
        loading={loadingMore}
        failed={loadMoreFailed}
        loadingLabel="继续加载本地音乐"
        completeLabel="已加载全部本地音乐"
        onLoadMore={() => void onLoadMore()}
      />
    </>
  );
}
