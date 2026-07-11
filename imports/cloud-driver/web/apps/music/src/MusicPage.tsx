import { lazy, Suspense, useCallback, useMemo, useState } from "react";
import {
  Heart,
  History,
  Library,
  Radio,
  Trash2,
  UserRoundCog,
} from "lucide-react";
import type { PlayableTrack } from "@cloud-drive/contracts";
import {
  MusicClient,
  LoadingIndicator,
  PrivateAppShell,
  Tabs,
  TabsList,
  TabsTrigger,
} from "@cloud-drive/shared";
import { LocalLibrary, type LocalLibraryMode } from "./LocalLibrary";
import { LyricsDialog } from "./LyricsDialog";
import { OnlineSearch } from "./OnlineSearch";
import { PlaybackHistory } from "./PlaybackHistory";
import { PlayerBar } from "./PlayerBar";
import { API_BASE, APP_LINKS } from "./config";

const ProviderAccounts = lazy(() =>
  import("./ProviderAccounts").then((module) => ({
    default: module.ProviderAccounts,
  })),
);

type View = LocalLibraryMode | "online" | "history" | "accounts";

export function MusicPage() {
  const client = useMemo(() => new MusicClient(API_BASE), []);
  const [view, setView] = useState<View>("all");
  const [current, setCurrent] = useState<PlayableTrack>();
  const [queue, setQueue] = useState<PlayableTrack[]>([]);
  const [lyricsTrack, setLyricsTrack] = useState<PlayableTrack>();
  const [historyVersion, setHistoryVersion] = useState(0);
  const play = useCallback(
    (track: PlayableTrack, nextQueue: PlayableTrack[]) => {
      setCurrent(track);
      setQueue(
        nextQueue.filter((candidate) => candidate.provider === track.provider),
      );
    },
    [],
  );
  const playNext = useCallback(() => {
    if (!current) return;
    const index = queue.findIndex(
      (track) =>
        track.provider === current.provider &&
        track.trackId === current.trackId,
    );
    setCurrent(index >= 0 ? queue[index + 1] : undefined);
  }, [current, queue]);
  return (
    <PrivateAppShell
      app="music"
      title="音乐盒"
      subtitle="本地收藏与独立会员音乐来源"
      apiBase={API_BASE}
      links={APP_LINKS}
    >
      <div className={current ? "pb-28" : ""}>
        <div className="mb-6 overflow-x-auto pb-1">
          <Tabs value={view} onValueChange={(value) => setView(value as View)}>
            <TabsList>
              <TabsTrigger value="all">
                <Library />
                本地音乐
              </TabsTrigger>
              <TabsTrigger value="online">
                <Radio />
                在线搜索
              </TabsTrigger>
              <TabsTrigger value="favorites">
                <Heart />
                收藏
              </TabsTrigger>
              <TabsTrigger value="history">
                <History />
                最近播放
              </TabsTrigger>
              <TabsTrigger value="accounts">
                <UserRoundCog />
                账号
              </TabsTrigger>
              <TabsTrigger value="trash">
                <Trash2 />
                回收站
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>
        {view === "online" ? (
          <OnlineSearch
            client={client}
            current={current}
            onPlay={play}
            onLyrics={setLyricsTrack}
          />
        ) : view === "history" ? (
          <PlaybackHistory
            client={client}
            current={current}
            refreshKey={historyVersion}
            onPlay={play}
            onLyrics={setLyricsTrack}
          />
        ) : view === "accounts" ? (
          <Suspense fallback={<LoadingIndicator label="正在载入账号管理" />}>
            <ProviderAccounts client={client} />
          </Suspense>
        ) : (
          <LocalLibrary
            mode={view}
            apiBase={API_BASE}
            client={client}
            current={current}
            onPlay={play}
          />
        )}
      </div>
      {current ? (
        <PlayerBar
          track={current}
          client={client}
          onEnded={playNext}
          onRecorded={() => setHistoryVersion((value) => value + 1)}
        />
      ) : null}
      <LyricsDialog
        track={lyricsTrack}
        client={client}
        onClose={() => setLyricsTrack(undefined)}
      />
    </PrivateAppShell>
  );
}
