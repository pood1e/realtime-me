import { useEffect, useRef, useState } from "react";
import { Pause, Play, Volume2 } from "lucide-react";
import {
  MusicProvider,
  PlaybackQuality,
  type PlaybackDescriptor,
  type PlayableTrack,
} from "@cloud-drive/contracts";
import { Badge, Button, MusicClient } from "@cloud-drive/shared";
import { clock, durationSeconds, providerLabel } from "./music-model";
import { SpotifyController } from "./spotify-player";

export function PlayerBar({
  track,
  client,
  onEnded,
  onRecorded,
}: {
  track: PlayableTrack;
  client: MusicClient;
  onEnded: () => void;
  onRecorded: () => void;
}) {
  const audio = useRef<HTMLAudioElement>(null);
  const spotify = useRef<SpotifyController | undefined>(undefined);
  const recorded = useRef(false);
  const ended = useRef(false);
  const [descriptor, setDescriptor] = useState<PlaybackDescriptor>();
  const [playing, setPlaying] = useState(false);
  const [time, setTime] = useState(0);
  const [duration, setDuration] = useState(durationSeconds(track));
  const [fallbackUsed, setFallbackUsed] = useState(false);
  const [error, setError] = useState("");
  useEffect(() => {
    spotify.current?.disconnect();
    spotify.current = undefined;
    recorded.current = false;
    ended.current = false;
    setDescriptor(undefined);
    setPlaying(false);
    setTime(0);
    setDuration(durationSeconds(track));
    setFallbackUsed(false);
    setError("");
    let active = true;
    void client
      .resolvePlayback(track)
      .then((playback) => {
        if (!active) return;
        setDescriptor(playback);
        if (playback.playback.case === "spotify") {
          const controller = new SpotifyController(
            client,
            (state) => {
              setPlaying(!state.paused);
              setTime(state.position);
              setDuration(state.duration);
              recordMeaningfulPlayback(state.position, state.duration);
              if (
                !ended.current &&
                state.duration > 0 &&
                state.position >= state.duration - 0.6
              ) {
                ended.current = true;
                onEnded();
              }
            },
            setError,
          );
          spotify.current = controller;
          void controller
            .play(playback.playback.value.uri)
            .catch((reason: unknown) => setError(message(reason)));
        } else {
        }
      })
      .catch((reason: unknown) => setError(message(reason)));
    return () => {
      active = false;
    };
  }, [client, track]);
  useEffect(() => () => spotify.current?.disconnect(), []);
  const direct =
    descriptor?.playback.case === "directAudio"
      ? descriptor.playback.value
      : undefined;
  const directURL = direct ? client.playbackUrl(direct.url) : "";
  useEffect(() => {
    const node = audio.current;
    if (!node || !directURL) return;
    node.load();
    void node
      .play()
      .then(() => setPlaying(true))
      .catch(() => setPlaying(false));
  }, [directURL]);
  const recordMeaningfulPlayback = (
    currentTime: number,
    mediaDuration: number,
  ) => {
    const threshold = Math.max(1, Math.min(10, mediaDuration / 2));
    if (currentTime < threshold || recorded.current) return;
    recorded.current = true;
    void client
      .recordPlayback(track)
      .then(onRecorded)
      .catch(() => undefined);
  };
  const toggle = () => {
    if (descriptor?.playback.case === "spotify") {
      void spotify.current
        ?.toggle()
        .catch((reason: unknown) => setError(message(reason)));
      return;
    }
    const node = audio.current;
    if (!node) return;
    if (node.paused) void node.play();
    else node.pause();
  };
  const seek = (value: number) => {
    if (descriptor?.playback.case === "spotify") {
      void spotify.current
        ?.seek(value)
        .catch((reason: unknown) => setError(message(reason)));
    } else if (audio.current) audio.current.currentTime = value;
    setTime(value);
  };
  const retryLowerQuality = () => {
    if (track.provider === MusicProvider.LOCAL || fallbackUsed) {
      setError("当前音频无法播放");
      return;
    }
    setFallbackUsed(true);
    void client
      .resolvePlayback(track, PlaybackQuality.HIGH)
      .then(setDescriptor)
      .catch((reason: unknown) => setError(message(reason)));
  };
  const artwork = client.playableArtworkUrl(track);
  return (
    <div className="fixed right-0 bottom-0 left-0 z-40 border-t bg-card/95 px-4 py-3 shadow-2xl backdrop-blur lg:left-60">
      {direct ? (
        <audio
          ref={audio}
          crossOrigin={
            track.provider === MusicProvider.LOCAL
              ? "use-credentials"
              : undefined
          }
          src={directURL}
          onPlay={() => setPlaying(true)}
          onPause={() => setPlaying(false)}
          onTimeUpdate={(event) => {
            const position = event.currentTarget.currentTime;
            setTime(position);
            recordMeaningfulPlayback(position, duration);
          }}
          onLoadedMetadata={(event) =>
            setDuration(event.currentTarget.duration)
          }
          onEnded={onEnded}
          onError={retryLowerQuality}
        />
      ) : null}
      <div className="mx-auto grid max-w-[92rem] grid-cols-[minmax(0,1fr)_auto] items-center gap-4 sm:grid-cols-[minmax(0,1fr)_minmax(15rem,1fr)_auto]">
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
            <div className="flex min-w-0 items-center gap-2">
              <p className="truncate text-sm font-medium">{track.title}</p>
              <Badge
                variant="outline"
                className="hidden shrink-0 md:inline-flex"
              >
                {providerLabel(track.provider)}
              </Badge>
            </div>
            <p className="truncate text-xs text-muted-foreground">
              {error || track.artists.join("、") || "未知艺人"}
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
            value={Math.min(time, duration || 0)}
            onChange={(event) => seek(Number(event.target.value))}
            className="w-full accent-primary"
          />
          <span className="w-10 text-[11px] text-muted-foreground">
            {clock(duration)}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <Volume2 className="hidden size-4 text-muted-foreground md:block" />
          <Button size="icon" onClick={toggle} disabled={!descriptor}>
            {playing ? <Pause /> : <Play />}
          </Button>
        </div>
      </div>
    </div>
  );
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : "播放失败";
}
