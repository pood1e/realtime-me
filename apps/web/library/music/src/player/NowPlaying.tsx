import type { PlayableTrack, PlaybackDescriptor } from "@realtime-me/library-contracts";
import { LOCAL_PROVIDER_ID, type MusicClient } from "@realtime-me/library-web";
import { Badge } from "@realtime-me/web-ui";
import { useProviderLabel } from "../provider-catalog";

export function NowPlaying({
  track,
  descriptor,
  client,
  error,
}: {
  track: PlayableTrack;
  descriptor: PlaybackDescriptor | undefined;
  client: MusicClient;
  error: string;
}) {
  const providerLabel = useProviderLabel();
  const artwork = client.providers.artworkUrl(track);
  const sourceLabel =
    descriptor?.providerId === LOCAL_PROVIDER_ID && track.providerId !== LOCAL_PROVIDER_ID
      ? "本地缓存"
      : providerLabel(track.providerId);
  return (
    <div className="flex min-w-0 items-center gap-3">
      {artwork ? (
        <img src={artwork} alt="" className="size-10 shrink-0 rounded-md object-cover sm:size-11" />
      ) : (
        <div className="size-10 shrink-0 rounded-md bg-muted sm:size-11" />
      )}
      <div className="min-w-0">
        <div className="flex min-w-0 items-center gap-2">
          <p className="truncate text-sm font-medium">{track.title}</p>
          <Badge variant="outline" className="hidden shrink-0 lg:inline-flex">
            {sourceLabel}
          </Badge>
        </div>
        <p
          className={
            error ? "truncate text-xs text-destructive" : "truncate text-xs text-muted-foreground"
          }
        >
          {error || track.artists.join("、") || "未知艺人"}
        </p>
      </div>
    </div>
  );
}
