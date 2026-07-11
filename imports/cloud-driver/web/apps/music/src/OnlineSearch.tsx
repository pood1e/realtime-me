import { useState, type FormEvent } from "react";
import { Music2, Search } from "lucide-react";
import {
  ProviderSearchStatus,
  type PlayableTrack,
  type ProviderSearchGroup,
} from "@cloud-drive/contracts";
import {
  Button,
  EmptyState,
  Input,
  LoadingIndicator,
  MusicClient,
  useToast,
} from "@cloud-drive/shared";
import { PlayableTrackRow } from "./PlayableTrackRow";
import { providerLabel } from "./music-model";

export function OnlineSearch({
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
  const [query, setQuery] = useState("");
  const [submittedQuery, setSubmittedQuery] = useState("");
  const [groups, setGroups] = useState<ProviderSearchGroup[]>([]);
  const [loading, setLoading] = useState(false);
  const submit = async (event: FormEvent) => {
    event.preventDefault();
    const normalized = query.trim();
    if (!normalized) return;
    setLoading(true);
    setSubmittedQuery(normalized);
    try {
      setGroups(await client.searchMusic(normalized));
    } catch (error) {
      showToast(message(error), "error");
    } finally {
      setLoading(false);
    }
  };
  const loadMore = async (group: ProviderSearchGroup) => {
    if (!group.nextPageToken) return;
    try {
      const [page] = await client.searchMusic(submittedQuery, [
        { provider: group.provider, pageToken: group.nextPageToken },
      ]);
      if (!page) return;
      setGroups((currentGroups) =>
        currentGroups.map((currentGroup) =>
          currentGroup.provider === page.provider
            ? {
                ...page,
                tracks: [...currentGroup.tracks, ...page.tracks],
              }
            : currentGroup,
        ),
      );
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  return (
    <div className="space-y-7">
      <form onSubmit={(event) => void submit(event)} className="flex gap-2">
        <div className="relative min-w-0 flex-1">
          <Search className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="同时搜索本地、QQ、网易云和 Spotify"
            className="pl-9"
          />
        </div>
        <Button type="submit" disabled={loading || !query.trim()}>
          搜索
        </Button>
      </form>
      {loading ? (
        <LoadingIndicator label="正在查询各音乐来源" />
      ) : groups.length ? (
        groups.map((group) => (
          <SearchGroup
            key={group.provider}
            group={group}
            current={current}
            client={client}
            onPlay={onPlay}
            onLyrics={onLyrics}
            onLoadMore={() => void loadMore(group)}
          />
        ))
      ) : (
        <EmptyState
          icon={<Music2 className="size-6" />}
          title="搜索你的音乐来源"
          detail="结果按来源分组展示，Spotify 不会与其他平台混合播放。"
        />
      )}
    </div>
  );
}

function SearchGroup({
  group,
  current,
  client,
  onPlay,
  onLyrics,
  onLoadMore,
}: {
  group: ProviderSearchGroup;
  current?: PlayableTrack;
  client: MusicClient;
  onPlay: (track: PlayableTrack, queue: PlayableTrack[]) => void;
  onLyrics: (track: PlayableTrack) => void;
  onLoadMore: () => void;
}) {
  const detail = groupDetail(group);
  return (
    <section>
      <div className="mb-3 flex items-end justify-between gap-3">
        <div>
          <h2 className="text-sm font-semibold">
            {providerLabel(group.provider)}
          </h2>
          {detail ? (
            <p className="mt-1 text-xs text-muted-foreground">{detail}</p>
          ) : null}
        </div>
        <span className="text-xs text-muted-foreground">
          {group.tracks.length ? `${group.tracks.length} 首` : ""}
        </span>
      </div>
      {group.status === ProviderSearchStatus.READY && group.tracks.length ? (
        <div className="overflow-hidden rounded-xl border bg-card/35">
          {group.tracks.map((track, index) => (
            <PlayableTrackRow
              key={`${track.provider}-${track.trackId}`}
              track={track}
              index={index + 1}
              active={sameTrack(current, track)}
              client={client}
              onPlay={() => onPlay(track, group.tracks)}
              onLyrics={() => onLyrics(track)}
            />
          ))}
          {group.nextPageToken ? (
            <div className="flex justify-center border-t p-3">
              <Button variant="ghost" size="sm" onClick={onLoadMore}>
                加载更多
              </Button>
            </div>
          ) : null}
        </div>
      ) : (
        <div className="rounded-xl border border-dashed px-4 py-8 text-center text-sm text-muted-foreground">
          {detail || "没有找到歌曲"}
        </div>
      )}
    </section>
  );
}

function groupDetail(group: ProviderSearchGroup): string {
  switch (group.status) {
    case ProviderSearchStatus.NOT_CONNECTED:
      return "尚未连接该来源，请先前往账号页面连接。";
    case ProviderSearchStatus.RECONNECT_REQUIRED:
      return "登录已经失效，请重新连接账号。";
    case ProviderSearchStatus.UNAVAILABLE:
      return "该来源暂时不可用，其他来源不受影响。";
    default:
      return group.tracks.length ? "" : "没有找到歌曲";
  }
}

function sameTrack(a: PlayableTrack | undefined, b: PlayableTrack): boolean {
  return a?.provider === b.provider && a.trackId === b.trackId;
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : "搜索未完成";
}
