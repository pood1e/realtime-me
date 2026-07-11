import { useCallback, useEffect, useMemo, useState } from "react";
import { Music2, Search } from "lucide-react";
import type { PlayableTrack, Track } from "@cloud-drive/contracts";
import {
  Button,
  EmptyState,
  Input,
  LoadingIndicator,
  MusicClient,
  UploadButton,
  UploadClient,
  useToast,
} from "@cloud-drive/shared";
import { TrackRow } from "./TrackRow";
import { localPlayableTrack } from "./music-model";

export type LocalLibraryMode = "all" | "favorites" | "trash";

export function LocalLibrary({
  mode,
  apiBase,
  client,
  current,
  onPlay,
}: {
  mode: LocalLibraryMode;
  apiBase: string;
  client: MusicClient;
  current?: PlayableTrack;
  onPlay: (track: PlayableTrack, queue: PlayableTrack[]) => void;
}) {
  const uploader = useMemo(() => new UploadClient(apiBase), [apiBase]);
  const { showToast } = useToast();
  const [tracks, setTracks] = useState<Track[]>([]);
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const load = useCallback(async () => {
    setLoading(true);
    try {
      setTracks(
        await client.tracks({
          query,
          favorites: mode === "favorites",
          trashed: mode === "trash",
        }),
      );
    } catch (error) {
      showToast(message(error), "error");
    } finally {
      setLoading(false);
    }
  }, [client, mode, query, showToast]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 180);
    return () => window.clearTimeout(timer);
  }, [load]);
  const queue = useMemo(() => tracks.map(localPlayableTrack), [tracks]);
  const upload = async (files: File[]) => {
    for (const file of files) {
      try {
        await client.importUpload(await uploader.upload(file));
        showToast(`${file.name} 已加入音乐库`);
      } catch (error) {
        showToast(`${file.name}: ${message(error)}`, "error");
      }
    }
    await load();
  };
  const favorite = async (track: Track) => {
    try {
      await client.favorite(track.uid, !track.favorite);
      await load();
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  const remove = async (track: Track) => {
    if (mode === "trash" && !window.confirm("永久删除这首音乐？")) return;
    try {
      if (mode === "trash") await client.purge(track.uid);
      else await client.trash(track.uid);
      await load();
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  const restore = async (track: Track) => {
    try {
      await client.restore(track.uid);
      await load();
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  const emptyTrash = async () => {
    if (!window.confirm("永久删除音乐回收站中的全部文件？")) return;
    try {
      await client.emptyTrash();
      await load();
      showToast("回收站已清空");
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  return (
    <div>
      <div className="mb-5 flex flex-col gap-3 sm:flex-row sm:items-center">
        <div className="relative min-w-0 flex-1 sm:max-w-md">
          <Search className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="搜索本地音乐、艺人或专辑"
            className="pl-9"
          />
        </div>
        <div className="sm:ml-auto">
          {mode === "trash" ? (
            <Button variant="destructive" onClick={() => void emptyTrash()}>
              清空回收站
            </Button>
          ) : (
            <UploadButton
              accept="audio/*,.flac,.m4a,.ogg,.opus,.wav"
              onFiles={upload}
              label="导入音乐"
            />
          )}
        </div>
      </div>
      {loading ? (
        <LoadingIndicator label="正在读取本地音乐" />
      ) : tracks.length ? (
        <div className="overflow-hidden rounded-xl border bg-card/35">
          {tracks.map((track, index) => (
            <TrackRow
              key={track.uid}
              track={track}
              index={index}
              client={client}
              active={current?.trackId === track.uid}
              trashed={mode === "trash"}
              onPlay={() => onPlay(localPlayableTrack(track), queue)}
              onFavorite={() => void favorite(track)}
              onRemove={() => void remove(track)}
              onRestore={() => void restore(track)}
            />
          ))}
        </div>
      ) : (
        <EmptyState
          icon={<Music2 className="size-6" />}
          title={mode === "trash" ? "回收站是空的" : "本地音乐库还是空的"}
          detail={
            mode === "trash"
              ? "删除的音乐会先保留在这里。"
              : "导入音频后会自动读取标题、艺人、专辑与封面。"
          }
        />
      )}
    </div>
  );
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : "操作未完成";
}
