import { useEffect, useRef, useState } from "react";
import { Pause, Play, Volume2 } from "lucide-react";
import type { Track } from "@cloud-drive/contracts";
import { Button, MusicClient } from "@cloud-drive/shared";

export function PlayerBar({
  track,
  client,
  onEnded,
}: {
  track: Track;
  client: MusicClient;
  onEnded: () => void;
}) {
  const audio = useRef<HTMLAudioElement>(null);
  const [playing, setPlaying] = useState(false);
  const [time, setTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const recorded = useRef("");
  useEffect(() => {
    const node = audio.current;
    if (!node) return;
    node.load();
    void node
      .play()
      .then(() => setPlaying(true))
      .catch(() => setPlaying(false));
  }, [track]);
  const toggle = () => {
    const node = audio.current;
    if (!node) return;
    if (node.paused) void node.play();
    else node.pause();
  };
  const recordMeaningfulPlayback = (currentTime: number) => {
    const threshold = Math.max(1, Math.min(10, duration / 2));
    if (currentTime < threshold || recorded.current === track.uid) return;
    recorded.current = track.uid;
    void client.recordPlayback(track.uid);
  };
  const artwork = track.artworkUrl ? client.artworkUrl(track) : "";
  return (
    <div className="fixed right-0 bottom-0 left-0 z-40 border-t bg-card/95 px-4 py-3 shadow-2xl backdrop-blur lg:left-60">
      <audio
        ref={audio}
        crossOrigin="use-credentials"
        src={client.contentUrl(track)}
        onPlay={() => setPlaying(true)}
        onPause={() => setPlaying(false)}
        onTimeUpdate={(event) => {
          const currentTime = event.currentTarget.currentTime;
          setTime(currentTime);
          recordMeaningfulPlayback(currentTime);
        }}
        onLoadedMetadata={(event) => setDuration(event.currentTarget.duration)}
        onEnded={onEnded}
      />
      <div className="mx-auto grid max-w-6xl grid-cols-[minmax(0,1fr)_auto] items-center gap-4 sm:grid-cols-[minmax(0,1fr)_minmax(15rem,1fr)_auto]">
        <div className="flex min-w-0 items-center gap-3">
          {artwork ? (
            <img
              src={artwork}
              alt=""
              className="size-11 rounded-md object-cover"
            />
          ) : (
            <div className="size-11 rounded-md bg-muted" />
          )}
          <div className="min-w-0">
            <p className="truncate text-sm font-medium">{track.title}</p>
            <p className="truncate text-xs text-muted-foreground">
              {track.artists.join("、")}
            </p>
          </div>
        </div>
        <div className="hidden items-center gap-3 sm:flex">
          <span className="w-10 text-right text-[11px] text-muted-foreground">
            {clock(time)}
          </span>
          <input
            aria-label="播放进度"
            type="range"
            min={0}
            max={duration || 0}
            value={time}
            onChange={(event) => {
              if (audio.current)
                audio.current.currentTime = Number(event.target.value);
            }}
            className="w-full accent-primary"
          />
          <span className="w-10 text-[11px] text-muted-foreground">
            {clock(duration)}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <Volume2 className="hidden size-4 text-muted-foreground md:block" />
          <Button size="icon" onClick={toggle}>
            {playing ? <Pause /> : <Play />}
          </Button>
        </div>
      </div>
    </div>
  );
}
function clock(value: number) {
  if (!Number.isFinite(value)) return "0:00";
  const minutes = Math.floor(value / 60);
  return `${minutes}:${Math.floor(value % 60)
    .toString()
    .padStart(2, "0")}`;
}
