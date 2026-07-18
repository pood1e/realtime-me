import { lazy, Suspense, useCallback, useMemo, useState } from "react";
import {
  Heart,
  History,
  Library,
  ListMusic,
  Radio,
  Trash2,
  UserRoundCog,
} from "lucide-react";
import type { PlayableTrack } from "@realtime-me/library-contracts";
import {
  LoadingIndicator,
  MusicClient,
  PrivateAppShell,
  Tabs,
  TabsList,
  TabsTrigger,
} from "@realtime-me/library-web";
import { API_BASE, APP_LINKS } from "./config";
import { LocalLibrary } from "./LocalLibrary";
import { LyricsDialog } from "./LyricsDialog";
import { OnlineSearch } from "./OnlineSearch";
import { PlaybackHistory } from "./PlaybackHistory";
import { useMediaSession } from "./playback/media-session";
import { usePlaybackQueue } from "./playback/playback-queue";
import { usePlaybackSettings } from "./playback/playback-storage";
import { usePlaybackShortcuts } from "./playback/playback-shortcuts";
import { usePlaybackSession } from "./playback/use-playback-session";
import { PlayerBar } from "./player/PlayerBar";
import { MusicProviderCatalog } from "./provider-catalog";
import { registerSpotifyBrowserPlayer } from "./spotify-player";
import type { LocalLibraryMode } from "./useLocalTrackCatalog";

const ProviderAccounts = lazy(() =>
  import("./ProviderAccounts").then((module) => ({
    default: module.ProviderAccounts,
  })),
);

const PlaylistLibrary = lazy(() =>
  import("./PlaylistLibrary").then((module) => ({
    default: module.PlaylistLibrary,
  })),
);

type View = LocalLibraryMode | "online" | "playlists" | "history" | "accounts";

registerSpotifyBrowserPlayer();

export function MusicPage() {
  const client = useMemo(() => new MusicClient(API_BASE), []);
  const [view, setView] = useState<View>("all");
  const [lyricsTrack, setLyricsTrack] = useState<PlayableTrack>();
  const [historyVersion, setHistoryVersion] = useState(0);
  const settings = usePlaybackSettings();
  const queue = usePlaybackQueue({
    initialMode: settings.mode,
    onModeChange: settings.setMode,
    random: Math.random,
  });
  const recordHistory = useCallback(
    () => setHistoryVersion((value) => value + 1),
    [],
  );
  const session = usePlaybackSession({
    track: queue.currentTrack,
    playbackSequence: queue.playbackSequence,
    client,
    volume: settings.volume,
    muted: settings.muted,
    onEnded: () => void queue.next("ended"),
    onRecorded: recordHistory,
  });
  const previous = useCallback(
    () => queue.previous(session.position),
    [queue.previous, session.position],
  );
  const next = useCallback(() => void queue.next(), [queue.next]);
  const current = queue.currentTrack;
  const artwork = current ? client.providers.artworkUrl(current) : "";

  useMediaSession({
    track: current,
    artwork,
    playing: session.playing,
    position: session.position,
    duration: session.duration,
    canNext: queue.canNext,
    onPlay: session.play,
    onPause: session.pause,
    onPrevious: previous,
    onNext: next,
    onSeek: session.seek,
  });
  usePlaybackShortcuts({
    enabled: Boolean(current),
    position: session.position,
    onToggle: session.toggle,
    onPrevious: previous,
    onNext: next,
    onSeek: session.seek,
    onToggleMuted: settings.toggleMuted,
  });

  return (
    <MusicProviderCatalog client={client}>
      <PrivateAppShell
        app="music"
        title="音乐盒"
        subtitle="本地收藏与独立会员音乐来源"
        apiBase={API_BASE}
        links={APP_LINKS}
      >
        <div className={current ? "pb-40 md:pb-28" : ""}>
          <MusicNavigation view={view} onViewChange={setView} />
          {view === "online" ? (
            <OnlineSearch
              client={client}
              current={current}
              onPlay={queue.start}
              onLyrics={setLyricsTrack}
            />
          ) : view === "playlists" ? (
            <Suspense fallback={<LoadingIndicator label="正在载入歌单" />}>
              <PlaylistLibrary
                client={client}
                current={current}
                onPlay={queue.start}
                onLyrics={setLyricsTrack}
              />
            </Suspense>
          ) : view === "history" ? (
            <PlaybackHistory
              client={client}
              current={current}
              refreshKey={historyVersion}
              onPlay={queue.start}
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
              onPlay={queue.start}
            />
          )}
        </div>
        <PlayerBar
          client={client}
          queue={queue}
          session={session}
          volume={settings.volume}
          muted={settings.muted}
          onVolumeChange={settings.setVolume}
          onToggleMuted={settings.toggleMuted}
        />
        <LyricsDialog
          track={lyricsTrack}
          client={client}
          onClose={() => setLyricsTrack(undefined)}
        />
      </PrivateAppShell>
    </MusicProviderCatalog>
  );
}

function MusicNavigation({
  view,
  onViewChange,
}: {
  view: View;
  onViewChange: (view: View) => void;
}) {
  return (
    <div className="mb-6 overflow-x-auto pb-1">
      <Tabs value={view} onValueChange={(value) => onViewChange(value as View)}>
        <TabsList>
          <TabsTrigger value="all">
            <Library />
            本地音乐
          </TabsTrigger>
          <TabsTrigger value="online">
            <Radio />
            在线搜索
          </TabsTrigger>
          <TabsTrigger value="playlists">
            <ListMusic />
            歌单
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
  );
}
