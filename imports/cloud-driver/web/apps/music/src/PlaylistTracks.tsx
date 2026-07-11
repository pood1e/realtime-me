import { useEffect } from "react";
import {
  PlaylistTrackDownloadStatus,
  type PlayableTrack,
  type Playlist,
  type PlaylistTrack,
} from "@cloud-drive/contracts";
import {
  InfiniteScrollSentinel,
  LoadingIndicator,
  MusicClient,
  useCursorQuery,
  useToast,
} from "@cloud-drive/shared";
import { PlayableTrackRow } from "./PlayableTrackRow";

export function PlaylistTracks({
  playlist,
  client,
  current,
  onPlay,
  onLyrics,
}: {
  playlist: Playlist;
  client: MusicClient;
  current: PlayableTrack | undefined;
  onPlay: (track: PlayableTrack, queue: PlayableTrack[]) => void;
  onLyrics: (track: PlayableTrack) => void;
}) {
  const { showToast } = useToast();
  const catalog = useCursorQuery<PlaylistTrack>({
    queryKey: ["music-playlist-tracks", playlist.uid],
    pollInterval: 2_500,
    shouldPoll: (tracks) =>
      tracks.some(
        (track) =>
          track.downloadStatus === PlaylistTrackDownloadStatus.PENDING ||
          track.downloadStatus === PlaylistTrackDownloadStatus.RUNNING,
      ),
    loadPage: async (pageToken, signal) => {
      const page = await client.playlists.tracks(
        playlist.uid,
        pageToken,
        signal,
      );
      return { items: page.tracks, nextPageToken: page.nextPageToken };
    },
  });
  useEffect(() => {
    if (catalog.error) showToast(message(catalog.error), "error");
  }, [catalog.error, showToast]);
  if (catalog.isPending) return <LoadingIndicator label="正在读取歌单" />;
  const entries = catalog.items.filter(
    (entry): entry is typeof entry & { track: PlayableTrack } =>
      Boolean(entry.track),
  );
  const queue = entries.map((entry) => entry.track);
  return (
    <div className="border-t bg-background/35 px-3 py-3 sm:px-12">
      {entries.length ? (
        <>
          <div className="overflow-hidden rounded-lg border bg-card/35">
            {entries.map((entry, index) => (
              <PlayableTrackRow
                key={entry.uid}
                track={entry.track}
                index={index + 1}
                active={sameTrack(current, entry.track)}
                client={client}
                onPlay={() => onPlay(entry.track, queue)}
                onLyrics={() => onLyrics(entry.track)}
              />
            ))}
          </div>
          <InfiniteScrollSentinel
            hasMore={catalog.hasNextPage}
            loading={catalog.isFetchingNextPage}
            failed={catalog.isFetchNextPageError}
            loadingLabel="继续加载歌单歌曲"
            completeLabel="已加载全部歌单歌曲"
            onLoadMore={() => void catalog.fetchNextPage()}
          />
        </>
      ) : (
        <p className="py-8 text-center text-sm text-muted-foreground">
          这个歌单没有可用歌曲
        </p>
      )}
    </div>
  );
}

function sameTrack(a: PlayableTrack | undefined, b: PlayableTrack): boolean {
  return a?.providerId === b.providerId && a.trackId === b.trackId;
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : "歌单读取失败";
}
