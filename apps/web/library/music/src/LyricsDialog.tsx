import type { Lyric, PlayableTrack } from "@realtime-me/library-contracts";
import { AppDialog, LoadingIndicator, type MusicClient } from "@realtime-me/library-web";
import { useEffect, useState } from "react";

export function LyricsDialog({
  track,
  client,
  onClose,
}: {
  track: PlayableTrack | undefined;
  client: MusicClient;
  onClose: () => void;
}) {
  const [lyric, setLyric] = useState<Lyric>();
  const [error, setError] = useState("");
  useEffect(() => {
    if (!track) return;
    const controller = new AbortController();
    setLyric(undefined);
    setError("");
    void client.providers
      .lyrics(track, controller.signal)
      .then((value) => !controller.signal.aborted && setLyric(value))
      .catch((reason: unknown) => {
        if (!controller.signal.aborted) setError(message(reason));
      });
    return () => {
      controller.abort();
    };
  }, [client, track]);
  return (
    <AppDialog
      open={Boolean(track)}
      title={track?.title || "歌词"}
      description={track?.artists.join("、") || "来自当前音乐来源"}
      size="standard"
      onClose={onClose}
    >
      <div className="max-h-[62vh] overflow-y-auto rounded-lg border bg-card/35 p-5">
        {error ? (
          <p className="text-sm text-destructive">{error}</p>
        ) : !lyric ? (
          <LoadingIndicator label="正在读取歌词" />
        ) : (
          <LyricText lyric={lyric} />
        )}
      </div>
    </AppDialog>
  );
}

function LyricText({ lyric }: { lyric: Lyric }) {
  const original = lyric.syncedText || lyric.plainText;
  if (!original && !lyric.translatedText)
    return <p className="text-sm text-muted-foreground">暂无歌词</p>;
  return (
    <div className="grid gap-6 md:grid-cols-2">
      <pre className="whitespace-pre-wrap font-sans text-sm leading-7">
        {original || "暂无原文"}
      </pre>
      {lyric.translatedText ? (
        <pre className="whitespace-pre-wrap font-sans text-sm leading-7 text-muted-foreground">
          {lyric.translatedText}
        </pre>
      ) : null}
    </div>
  );
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : "歌词读取失败";
}
