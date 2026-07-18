import type { MediaStatus, PublicStatus, SwitchPresence } from "@realtime-me/status-contracts";
import { OnlineState } from "@realtime-me/status-contracts";
import { Footprints, HeartPulse, Music } from "lucide-react";
import type { ReactNode } from "react";
import { SwitchArtwork } from "@/components/SwitchArtwork";

export function Presence({ status }: { status?: PublicStatus | null }) {
  const watch = status?.mobiles.find((mobile) => mobile.watch)?.watch;
  const heartRate = watch?.heartRate?.beatsPerMinute;
  const steps = watch?.activityTotals?.steps;
  const game = switchGame(status);
  const media = nowPlaying(status);

  if (!heartRate && !steps && !game && !media) return null;

  return (
    <div className="hidden items-center gap-3.5 text-xs text-muted-foreground md:flex">
      {!!heartRate && (
        <span className="flex items-center gap-1" title="Heart rate">
          <HeartPulse className="size-3.5 text-primary" />
          {heartRate}
        </span>
      )}
      {!!steps && (
        <span className="flex items-center gap-1" title="Steps today">
          <Footprints className="size-3.5" />
          {steps.toLocaleString()}
        </span>
      )}
      {!!game && (
        <PlayingIndicator
          icon={<SwitchArtwork imageUri={game.imageUri} className="size-5" />}
          text={game.gameName}
          title={`Playing on Switch: ${game.gameName}`}
        />
      )}
      {media && <NowPlaying media={media} />}
    </div>
  );
}

function NowPlaying({ media }: { media: MediaStatus }) {
  const text = media.artist ? `${media.title} · ${media.artist}` : media.title;
  return (
    <PlayingIndicator
      icon={<Music className="size-3.5 text-primary" />}
      text={text}
      title={`Now playing: ${text}`}
    />
  );
}

function PlayingIndicator({ icon, text, title }: { icon: ReactNode; text: string; title: string }) {
  return (
    <span className="flex max-w-[15rem] items-center gap-1.5" title={title}>
      {icon}
      <span className="truncate">{text}</span>
    </span>
  );
}

function switchGame(status?: PublicStatus | null): SwitchPresence | null {
  for (const mobile of status?.mobiles ?? []) {
    const presence = mobile.switchPresence;
    if (presence?.state === OnlineState.ONLINE && presence.gameName) return presence;
  }
  return null;
}

function nowPlaying(status?: PublicStatus | null): MediaStatus | null {
  if (!status) return null;
  const devices = [status.server, ...(status.devices ?? [])];
  for (const device of devices) {
    if (device?.media?.title) return device.media;
  }
  return null;
}
