import { useCallback, useEffect, useMemo, useState } from "react";
import { Heart, History, Music2, Search, Trash2 } from "lucide-react";
import type { Track } from "@cloud-drive/contracts";
import {
  Button,
  EmptyState,
  Input,
  LoadingIndicator,
  MusicClient,
  PrivateAppShell,
  Tabs,
  TabsList,
  TabsTrigger,
  UploadButton,
  UploadClient,
  useToast,
} from "@cloud-drive/shared";
import { PlayerBar } from "./PlayerBar";
import { TrackRow } from "./TrackRow";
import { API_BASE, APP_LINKS } from "./config";

type View = "all" | "favorites" | "history" | "trash";

export function MusicPage() {
  const client = useMemo(() => new MusicClient(API_BASE), []);
  const uploader = useMemo(() => new UploadClient(API_BASE), []);
  const { showToast } = useToast();
  const [tracks, setTracks] = useState<Track[]>([]);
  const [query, setQuery] = useState("");
  const [view, setView] = useState<View>("all");
  const [loading, setLoading] = useState(true);
  const [current, setCurrent] = useState<Track>();
  const load = useCallback(async () => {
    setLoading(true);
    try {
      if (view === "history") {
        const history = await client.history();
        setTracks(
          history.flatMap((entry) => (entry.track ? [entry.track] : [])),
        );
      } else
        setTracks(
          await client.tracks({
            query,
            favorites: view === "favorites",
            trashed: view === "trash",
          }),
        );
    } catch (error) {
      showToast(message(error), "error");
    } finally {
      setLoading(false);
    }
  }, [client, query, showToast, view]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 180);
    return () => window.clearTimeout(timer);
  }, [load]);
  const upload = async (files: File[]) => {
    for (const file of files)
      try {
        const uid = await uploader.upload(file);
        await client.importUpload(uid);
        showToast(`${file.name} 已加入音乐库`);
      } catch (error) {
        showToast(`${file.name}: ${message(error)}`, "error");
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
    try {
      if (view === "trash") {
        if (!window.confirm("永久删除这首音乐？")) return;
        await client.purge(track.uid);
      } else await client.trash(track.uid);
      await load();
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  const restore = async (track: Track) => {
    await client.restore(track.uid);
    await load();
  };
  return (
    <PrivateAppShell
      app="music"
      title="音乐盒"
      subtitle="本地原音播放"
      apiBase={API_BASE}
      links={APP_LINKS}
      actions={
        view === "trash" ? (
          <Button
            variant="destructive"
            onClick={() => void emptyTrash(client, load, showToast)}
          >
            清空
          </Button>
        ) : (
          <UploadButton
            accept="audio/*,.flac,.m4a,.ogg,.opus,.wav"
            onFiles={upload}
            label="导入音乐"
          />
        )
      }
    >
      <div className="mb-6 flex flex-col gap-4 xl:flex-row xl:items-center">
        <Tabs value={view} onValueChange={(value) => setView(value as View)}>
          <TabsList>
            <TabsTrigger value="all">全部</TabsTrigger>
            <TabsTrigger value="favorites">
              <Heart />
              收藏
            </TabsTrigger>
            <TabsTrigger value="history">
              <History />
              最近播放
            </TabsTrigger>
            <TabsTrigger value="trash">
              <Trash2 />
              回收站
            </TabsTrigger>
          </TabsList>
        </Tabs>
        <div className="relative ml-auto w-full xl:w-80">
          <Search className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="搜索音乐、艺人或专辑"
            className="pl-9"
            disabled={view === "history"}
          />
        </div>
      </div>
      <div className={current ? "pb-28" : ""}>
        {loading ? (
          <LoadingIndicator label="正在读取音乐库" />
        ) : tracks.length ? (
          <div className="overflow-hidden rounded-xl border bg-card/50">
            <div className="hidden grid-cols-[3rem_minmax(0,1fr)_minmax(8rem,.6fr)_6rem_3rem] gap-3 border-b px-4 py-3 text-xs text-muted-foreground md:grid">
              <span>#</span>
              <span>标题</span>
              <span>专辑</span>
              <span>大小</span>
              <span />
            </div>
            {tracks.map((track, index) => (
              <TrackRow
                key={`${track.uid}-${index}`}
                track={track}
                index={index}
                client={client}
                active={current?.uid === track.uid}
                trashed={view === "trash"}
                onPlay={() => setCurrent(track)}
                onFavorite={() => void favorite(track)}
                onRemove={() => void remove(track)}
                onRestore={() => void restore(track)}
              />
            ))}
          </div>
        ) : (
          <EmptyState
            icon={<Music2 className="size-6" />}
            title="音乐库还是空的"
            detail="导入音频后会自动读取标题、艺人、专辑与封面。"
          />
        )}
      </div>
      {current ? (
        <PlayerBar
          track={current}
          client={client}
          onEnded={() => {
            const index = tracks.findIndex(
              (track) => track.uid === current.uid,
            );
            setCurrent(tracks[index + 1]);
          }}
        />
      ) : null}
    </PrivateAppShell>
  );
}

function message(error: unknown) {
  return error instanceof Error ? error.message : "操作未完成";
}
async function emptyTrash(
  client: MusicClient,
  reload: () => Promise<void>,
  toast: (message: string, variant?: "default" | "error") => void,
) {
  if (!window.confirm("永久删除音乐回收站？")) return;
  try {
    await client.emptyTrash();
    await reload();
    toast("回收站已清空");
  } catch (error) {
    toast(message(error), "error");
  }
}
