import { useCallback, useMemo, useState } from "react";
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
import { LocalTrackList } from "./LocalTrackList";
import { localPlayableTrack } from "./music-model";
import { useLocalTrackCatalog } from "./useLocalTrackCatalog";
import type { LocalLibraryMode } from "./useLocalTrackCatalog";

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
  const [query, setQuery] = useState("");
  const onLoadError = useCallback(
    (error: unknown) => showToast(message(error), "error"),
    [showToast],
  );
  const catalog = useLocalTrackCatalog({
    client,
    query,
    mode,
    onError: onLoadError,
  });
  const queue = useMemo(
    () => catalog.tracks.map(localPlayableTrack),
    [catalog.tracks],
  );
  const upload = async (files: File[]) => {
    for (const file of files) {
      try {
        await client.importUpload(await uploader.upload(file));
        showToast(`${file.name} 已加入音乐库`);
      } catch (error) {
        showToast(`${file.name}: ${message(error)}`, "error");
      }
    }
    await catalog.refresh();
  };
  const favorite = async (track: Track) => {
    try {
      const updated = await client.favorite(track.uid, !track.favorite);
      if (mode === "favorites" && !updated.favorite)
        catalog.removeTrack(track.uid);
      else catalog.updateTrack(updated);
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  const remove = async (track: Track) => {
    if (mode === "trash" && !window.confirm("永久删除这首音乐？")) return;
    try {
      if (mode === "trash") await client.purge(track.uid);
      else await client.trash(track.uid);
      catalog.removeTrack(track.uid);
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  const restore = async (track: Track) => {
    try {
      await client.restore(track.uid);
      catalog.removeTrack(track.uid);
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  const emptyTrash = async () => {
    if (!window.confirm("永久删除音乐回收站中的全部文件？")) return;
    try {
      await client.emptyTrash();
      await catalog.refresh();
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
      {catalog.initialLoading ? (
        <LoadingIndicator label="正在读取本地音乐" />
      ) : catalog.tracks.length ? (
        <LocalTrackList
          tracks={catalog.tracks}
          queue={queue}
          client={client}
          current={current}
          trashed={mode === "trash"}
          hasMore={catalog.hasMore}
          loadingMore={catalog.loadingMore}
          loadMoreFailed={catalog.loadMoreFailed}
          onLoadMore={catalog.loadMore}
          onPlay={(track, nextQueue) =>
            onPlay(localPlayableTrack(track), nextQueue)
          }
          onFavorite={(track) => void favorite(track)}
          onRemove={(track) => void remove(track)}
          onRestore={(track) => void restore(track)}
        />
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
