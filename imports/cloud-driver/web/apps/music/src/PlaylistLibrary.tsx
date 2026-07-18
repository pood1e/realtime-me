import { useEffect, useState } from "react";
import { ListMusic } from "lucide-react";
import { MusicProviderCapability } from "@cloud-drive/contracts";
import type {
  PlayableTrack,
  Playlist,
  ProviderDescriptor,
} from "@cloud-drive/contracts";
import {
  EmptyState,
  InfiniteScrollSentinel,
  LoadingIndicator,
  MusicClient,
  type ProviderId,
  useCursorQuery,
  useDialog,
  useQuery,
  useToast,
} from "@cloud-drive/shared";
import { PlaylistImportDialog } from "./PlaylistImportDialog";
import { PlaylistRow } from "./PlaylistRow";
import { PlaylistTracks } from "./PlaylistTracks";
import type { PlaybackQueueSelection } from "./playback/playback-types";

export function PlaylistLibrary({
  client,
  current,
  onPlay,
  onLyrics,
}: {
  client: MusicClient;
  current: PlayableTrack | undefined;
  onPlay: (selection: PlaybackQueueSelection) => void;
  onLyrics: (track: PlayableTrack) => void;
}) {
  const { showToast } = useToast();
  const { confirm } = useDialog();
  const [expanded, setExpanded] = useState("");
  const [busy, setBusy] = useState("");
  const catalog = useCursorQuery<Playlist>({
    queryKey: ["music-playlists"],
    pollInterval: 2_500,
    shouldPoll: (playlists) =>
      playlists.some((playlist) => playlist.pendingTrackCount > 0),
    loadPage: async (pageToken, signal) => {
      const page = await client.playlists.page(pageToken, signal);
      return { items: page.playlists, nextPageToken: page.nextPageToken };
    },
  });
  const descriptors = useQuery({
    queryKey: ["music-provider-descriptors"],
    queryFn: ({ signal }) => client.providers.descriptors(signal),
  });
  const providers = (descriptors.data ?? []).filter(
    (provider: ProviderDescriptor) =>
      provider.configured &&
      provider.capabilities.includes(MusicProviderCapability.PLAYLIST_IMPORT),
  );
  useEffect(() => {
    if (catalog.error) showToast(message(catalog.error), "error");
  }, [catalog.error, showToast]);
  useEffect(() => {
    if (descriptors.error) showToast(message(descriptors.error), "error");
  }, [descriptors.error, showToast]);
  const importPlaylist = async (providerId: ProviderId, source: string) => {
    setBusy("import");
    try {
      const playlist = await client.playlists.importPlaylist(
        providerId,
        source,
      );
      await catalog.refetch();
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
      await client.playlists.download(playlist.uid);
      await catalog.refetch();
      showToast("已开始存入本地，可离开此页面继续处理");
    } catch (error) {
      showToast(message(error), "error");
    } finally {
      setBusy("");
    }
  };
  const remove = async (playlist: Playlist) => {
    if (
      !(await confirm({
        title: "移除歌单",
        description: `移除“${playlist.displayName}”？已经存入本地的歌曲会保留。`,
        confirmLabel: "移除歌单",
        destructive: true,
      }))
    )
      return;
    setBusy(playlist.uid);
    try {
      await client.playlists.delete(playlist.uid);
      if (expanded === playlist.uid) setExpanded("");
      await catalog.refetch();
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
          providers={providers}
          onImport={importPlaylist}
        />
      </div>
      {catalog.isPending ? (
        <LoadingIndicator label="正在读取歌单" />
      ) : catalog.items.length ? (
        <>
          <div className="overflow-hidden rounded-xl border bg-card/35">
            {catalog.items.map((playlist) => (
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
          <InfiniteScrollSentinel
            hasMore={catalog.hasNextPage}
            loading={catalog.isFetchingNextPage}
            failed={catalog.isFetchNextPageError}
            loadingLabel="继续加载歌单"
            completeLabel="已加载全部歌单"
            onLoadMore={() => void catalog.fetchNextPage()}
          />
        </>
      ) : (
        <EmptyState
          icon={<ListMusic className="size-6" />}
          title="还没有导入歌单"
          detail="连接支持歌单导入的音乐来源后，可粘贴链接或歌单标识。"
        />
      )}
    </div>
  );
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : "歌单操作失败";
}
