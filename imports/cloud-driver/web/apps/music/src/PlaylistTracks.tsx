import { useEffect, useMemo, useState } from "react";
import type { PlayableTrack, Playlist } from "@cloud-drive/contracts";
import {
  Button,
  LoadingIndicator,
  MusicClient,
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
  current?: PlayableTrack;
  onPlay: (track: PlayableTrack, queue: PlayableTrack[]) => void;
  onLyrics: (track: PlayableTrack) => void;
}) {
  const { showToast } = useToast();
  const [tracks, setTracks] = useState<PlayableTrack[]>([]);
  const [nextPageToken, setNextPageToken] = useState("");
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    let active = true;
    setLoading(true);
    void client
      .playlistTracks(playlist.uid)
      .then((page) => {
        if (!active) return;
        setTracks(
          page.tracks.flatMap((item) => (item.track ? [item.track] : [])),
        );
        setNextPageToken(page.nextPageToken);
      })
      .catch((error: unknown) => showToast(message(error), "error"))
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, [client, playlist.uid, showToast]);
  const queue = useMemo(() => tracks, [tracks]);
  const loadMore = async () => {
    if (!nextPageToken) return;
    try {
      const page = await client.playlistTracks(playlist.uid, nextPageToken);
      setTracks((currentTracks) => [
        ...currentTracks,
        ...page.tracks.flatMap((item) => (item.track ? [item.track] : [])),
      ]);
      setNextPageToken(page.nextPageToken);
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  if (loading) return <LoadingIndicator label="正在读取歌单" />;
  return (
    <div className="border-t bg-background/35 px-3 py-3 sm:px-12">
      {tracks.length ? (
        <div className="overflow-hidden rounded-lg border bg-card/35">
          {tracks.map((track, index) => (
            <PlayableTrackRow
              key={`${track.provider}-${track.trackId}-${index}`}
              track={track}
              index={index + 1}
              active={sameTrack(current, track)}
              client={client}
              onPlay={() => onPlay(track, queue)}
              onLyrics={() => onLyrics(track)}
            />
          ))}
          {nextPageToken ? (
            <div className="flex justify-center border-t p-3">
              <Button variant="ghost" size="sm" onClick={() => void loadMore()}>
                加载更多
              </Button>
            </div>
          ) : null}
        </div>
      ) : (
        <p className="py-8 text-center text-sm text-muted-foreground">
          这个歌单没有可用歌曲
        </p>
      )}
    </div>
  );
}

function sameTrack(a: PlayableTrack | undefined, b: PlayableTrack): boolean {
  return a?.provider === b.provider && a.trackId === b.trackId;
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : "歌单读取失败";
}
