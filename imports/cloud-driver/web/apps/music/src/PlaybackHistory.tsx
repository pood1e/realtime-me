import { useEffect, useState } from "react";
import { History } from "lucide-react";
import type { PlayableTrack, PlaybackEntry } from "@cloud-drive/contracts";
import {
  EmptyState,
  LoadingIndicator,
  MusicClient,
  useToast,
} from "@cloud-drive/shared";
import { PlayableTrackRow } from "./PlayableTrackRow";

export function PlaybackHistory({
  client,
  current,
  refreshKey,
  onPlay,
  onLyrics,
}: {
  client: MusicClient;
  current?: PlayableTrack;
  refreshKey: number;
  onPlay: (track: PlayableTrack, queue: PlayableTrack[]) => void;
  onLyrics: (track: PlayableTrack) => void;
}) {
  const { showToast } = useToast();
  const [entries, setEntries] = useState<PlaybackEntry[]>([]);
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    let active = true;
    setLoading(true);
    void client
      .history()
      .then((history) => active && setEntries(history))
      .catch((error: unknown) => showToast(message(error), "error"))
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, [client, refreshKey, showToast]);
  if (loading) return <LoadingIndicator label="正在读取播放历史" />;
  const tracks = entries.flatMap((entry) => (entry.track ? [entry.track] : []));
  if (!tracks.length)
    return (
      <EmptyState
        icon={<History className="size-6" />}
        title="还没有播放记录"
        detail="歌曲开始播放后会记录在这里，不会向第三方平台回写。"
      />
    );
  return (
    <div className="overflow-hidden rounded-xl border bg-card/35">
      {tracks.map((track, index) => (
        <PlayableTrackRow
          key={`${entries[index]?.uid}-${track.provider}-${track.trackId}`}
          track={track}
          index={index + 1}
          active={sameTrack(current, track)}
          client={client}
          onPlay={() =>
            onPlay(
              track,
              tracks.filter(
                (candidate) => candidate.provider === track.provider,
              ),
            )
          }
          onLyrics={() => onLyrics(track)}
        />
      ))}
    </div>
  );
}

function sameTrack(a: PlayableTrack | undefined, b: PlayableTrack): boolean {
  return a?.provider === b.provider && a.trackId === b.trackId;
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : "播放历史读取失败";
}
