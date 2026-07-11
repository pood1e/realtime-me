import { useCallback, useEffect, useState } from "react";
import { ListMusic } from "lucide-react";
import type {
  MusicProvider,
  PlayableTrack,
  Playlist,
} from "@cloud-drive/contracts";
import {
  EmptyState,
  LoadingIndicator,
  MusicClient,
  useToast,
} from "@cloud-drive/shared";
import { PlaylistImportDialog } from "./PlaylistImportDialog";
import { PlaylistRow } from "./PlaylistRow";
import { PlaylistTracks } from "./PlaylistTracks";

export function PlaylistLibrary({
  client,
  current,
  onPlay,
  onLyrics,
}: {
  client: MusicClient;
  current?: PlayableTrack;
  onPlay: (track: PlayableTrack, queue: PlayableTrack[]) => void;
  onLyrics: (track: PlayableTrack) => void;
}) {
  const { showToast } = useToast();
  const [playlists, setPlaylists] = useState<Playlist[]>([]);
  const [expanded, setExpanded] = useState("");
  const [busy, setBusy] = useState("");
  const [loading, setLoading] = useState(true);
  const load = useCallback(
    async (showLoading = false) => {
      if (showLoading) setLoading(true);
      try {
        setPlaylists(await client.playlists());
      } catch (error) {
        showToast(message(error), "error");
      } finally {
        if (showLoading) setLoading(false);
      }
    },
    [client, showToast],
  );
  useEffect(() => {
    void load(true);
  }, [load]);
  const pending = playlists.some((playlist) => playlist.pendingTrackCount > 0);
  useEffect(() => {
    if (!pending) return;
    const timer = window.setInterval(() => void load(), 2500);
    return () => window.clearInterval(timer);
  }, [load, pending]);
  const importPlaylist = async (provider: MusicProvider, source: string) => {
    setBusy("import");
    try {
      const playlist = await client.importPlaylist(provider, source);
      await load();
      setExpanded(playlist.uid);
      showToast("歌单已导入");
      return true;
    } catch (error) {
      showToast(message(error), "error");
      return false;
    } finally {
      setBusy("");
    }
  };
  const download = async (playlist: Playlist) => {
    setBusy(playlist.uid);
    try {
      await client.downloadPlaylist(playlist.uid);
      await load();
      showToast("已开始存入本地，可离开此页面继续处理");
    } catch (error) {
      showToast(message(error), "error");
    } finally {
      setBusy("");
    }
  };
  const remove = async (playlist: Playlist) => {
    if (!window.confirm("移除这个歌单？已经存入本地的歌曲会保留。")) return;
    setBusy(playlist.uid);
    try {
      await client.deletePlaylist(playlist.uid);
      if (expanded === playlist.uid) setExpanded("");
      await load();
    } catch (error) {
      showToast(message(error), "error");
    } finally {
      setBusy("");
    }
  };
  return (
    <div className="space-y-5">
      <div className="flex justify-end">
        <PlaylistImportDialog
          importing={busy === "import"}
          onImport={importPlaylist}
        />
      </div>
      {loading ? (
        <LoadingIndicator label="正在读取歌单" />
      ) : playlists.length ? (
        <div className="overflow-hidden rounded-xl border bg-card/35">
          {playlists.map((playlist) => (
            <div key={playlist.uid} className="border-b last:border-b-0">
              <PlaylistRow
                playlist={playlist}
                expanded={expanded === playlist.uid}
                busy={busy === playlist.uid}
                onToggle={() =>
                  setExpanded((currentUID) =>
                    currentUID === playlist.uid ? "" : playlist.uid,
                  )
                }
                onDownload={() => void download(playlist)}
                onDelete={() => void remove(playlist)}
              />
              {expanded === playlist.uid ? (
                <PlaylistTracks
                  playlist={playlist}
                  client={client}
                  current={current}
                  onPlay={onPlay}
                  onLyrics={onLyrics}
                />
              ) : null}
            </div>
          ))}
        </div>
      ) : (
        <EmptyState
          icon={<ListMusic className="size-6" />}
          title="还没有导入歌单"
          detail="支持 QQ 音乐、网易云音乐和 Spotify 歌单链接。"
        />
      )}
    </div>
  );
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : "歌单操作失败";
}
