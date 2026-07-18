import {
  type PlayableTrack,
  type ProviderSearchGroup,
  ProviderSearchStatus,
} from "@realtime-me/library-contracts";
import {
  EmptyState,
  LoadingIndicator,
  type MusicClient,
  useQuery,
  useQueryClient,
  useToast,
} from "@realtime-me/library-web";
import { Button, Input } from "@realtime-me/web-ui";
import { Music2, Search } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { PlayableTrackRow } from "./PlayableTrackRow";
import type { PlaybackQueueSelection } from "./playback/playback-types";
import { useProviderLabel } from "./provider-catalog";

export function OnlineSearch({
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
  const queryClient = useQueryClient();
  const [query, setQuery] = useState("");
  const [submittedQuery, setSubmittedQuery] = useState("");
  const [groups, setGroups] = useState<ProviderSearchGroup[]>([]);
  const [loadingProviders, setLoadingProviders] = useState<Set<string>>(new Set());
  const search = useQuery({
    queryKey: ["music-search", submittedQuery],
    enabled: submittedQuery !== "",
    queryFn: ({ signal }) => client.providers.search(submittedQuery, [], signal),
  });
  useEffect(() => {
    if (search.data) setGroups(search.data);
  }, [search.data]);
  useEffect(() => {
    if (search.error) showToast(message(search.error), "error");
  }, [search.error, showToast]);
  const submit = (event: FormEvent) => {
    event.preventDefault();
    const normalized = query.trim();
    if (!normalized) return;
    if (normalized === submittedQuery) void search.refetch();
    else setSubmittedQuery(normalized);
  };
  const loadMore = async (group: ProviderSearchGroup) => {
    if (!group.nextPageToken || loadingProviders.has(group.providerId)) return;
    setLoadingProviders((current) => new Set(current).add(group.providerId));
    try {
      const [page] = await queryClient.fetchQuery({
        queryKey: ["music-search-page", submittedQuery, group.providerId, group.nextPageToken],
        queryFn: ({ signal }) =>
          client.providers.search(
            submittedQuery,
            [
              {
                providerId: group.providerId,
                pageToken: group.nextPageToken,
              },
            ],
            signal,
          ),
      });
      if (!page) return;
      setGroups((currentGroups) =>
        currentGroups.map((currentGroup) =>
          currentGroup.providerId === page.providerId
            ? {
                ...page,
                tracks: appendUnique(currentGroup.tracks, page.tracks),
              }
            : currentGroup,
        ),
      );
    } catch (error) {
      showToast(message(error), "error");
    } finally {
      setLoadingProviders((current) => {
        const next = new Set(current);
        next.delete(group.providerId);
        return next;
      });
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
            placeholder="同时搜索本地与已连接的音乐来源"
            className="pl-9"
          />
        </div>
        <Button type="submit" disabled={search.isFetching || !query.trim()}>
          搜索
        </Button>
      </form>
      {search.isFetching ? (
        <LoadingIndicator label="正在查询各音乐来源" />
      ) : groups.length ? (
        groups.map((group) => (
          <SearchGroup
            key={group.providerId}
            group={group}
            query={submittedQuery}
            current={current}
            client={client}
            onPlay={onPlay}
            onLyrics={onLyrics}
            loadingMore={loadingProviders.has(group.providerId)}
            onLoadMore={() => void loadMore(group)}
          />
        ))
      ) : (
        <EmptyState
          icon={<Music2 className="size-6" />}
          title="搜索你的音乐来源"
          detail="结果按来源分组展示，并按当前结果建立播放队列。"
        />
      )}
    </div>
  );
}

function SearchGroup({
  group,
  query,
  current,
  client,
  onPlay,
  onLyrics,
  onLoadMore,
  loadingMore,
}: {
  group: ProviderSearchGroup;
  query: string;
  current: PlayableTrack | undefined;
  client: MusicClient;
  onPlay: (selection: PlaybackQueueSelection) => void;
  onLyrics: (track: PlayableTrack) => void;
  onLoadMore: () => void;
  loadingMore: boolean;
}) {
  const providerLabel = useProviderLabel();
  const detail = groupDetail(group);
  const loadPlaybackPage = async (pageToken: string, signal: AbortSignal) => {
    const pages = await client.providers.search(
      query,
      [{ providerId: group.providerId, pageToken }],
      signal,
    );
    const page = pages.find((candidate) => candidate.providerId === group.providerId);
    return {
      tracks: page?.tracks ?? [],
      nextPageToken: page?.nextPageToken ?? "",
    };
  };
  return (
    <section>
      <div className="mb-3 flex items-end justify-between gap-3">
        <div>
          <h2 className="text-sm font-semibold">{providerLabel(group.providerId)}</h2>
          {detail ? <p className="mt-1 text-xs text-muted-foreground">{detail}</p> : null}
        </div>
        <span className="text-xs text-muted-foreground">
          {group.tracks.length ? `${group.tracks.length} 首` : ""}
        </span>
      </div>
      {group.status === ProviderSearchStatus.READY && group.tracks.length ? (
        <div className="overflow-hidden rounded-xl border bg-card/35">
          {group.tracks.map((track, index) => (
            <PlayableTrackRow
              key={`${track.providerId}-${track.trackId}`}
              track={track}
              index={index + 1}
              active={sameTrack(current, track)}
              client={client}
              onPlay={() =>
                onPlay({
                  tracks: group.tracks,
                  startIndex: index,
                  nextPageToken: group.nextPageToken,
                  loadNextPage: loadPlaybackPage,
                })
              }
              onLyrics={() => onLyrics(track)}
            />
          ))}
          {group.nextPageToken ? (
            <div className="flex justify-center border-t p-3">
              <Button variant="ghost" size="sm" onClick={onLoadMore} disabled={loadingMore}>
                {loadingMore ? "正在加载" : "加载更多"}
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
  return a?.providerId === b.providerId && a.trackId === b.trackId;
}

function appendUnique(current: PlayableTrack[], next: PlayableTrack[]): PlayableTrack[] {
  const known = new Set(current.map((track) => `${track.providerId}\u0000${track.trackId}`));
  return [
    ...current,
    ...next.filter((track) => !known.has(`${track.providerId}\u0000${track.trackId}`)),
  ];
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : "搜索未完成";
}
