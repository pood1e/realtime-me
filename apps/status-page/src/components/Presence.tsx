import { Footprints, Gamepad2, HeartPulse, Music } from 'lucide-react';
import type { PublicStatus } from '@/gen/realtime/me/v1/status_pb';
import type { MediaStatus } from '@/gen/realtime/me/v1/status_types_pb';

export function Presence({ status }: { status?: PublicStatus | null }) {
  const watch = status?.mobile?.watch;
  const heartRate = watch?.heartRate?.beatsPerMinute;
  const steps = watch?.activityTotals?.steps;
  const game = status?.mobile?.switchPresence?.gameName;
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
        <span className="flex max-w-[15rem] items-center gap-1.5" title={`Switch: ${game}`}>
          <Gamepad2 className="size-3.5 text-primary" />
          <span className="truncate">{game}</span>
        </span>
      )}
      {media && <NowPlaying media={media} />}
    </div>
  );
}

function NowPlaying({ media }: { media: MediaStatus }) {
  const text = media.artist ? `${media.title} · ${media.artist}` : media.title;
  return (
    <span className="flex max-w-[15rem] items-center gap-1.5" title={`Now playing: ${text}`}>
      <Music className="size-3.5 text-primary" />
      <span className="truncate">{text}</span>
    </span>
  );
}

function nowPlaying(status?: PublicStatus | null): MediaStatus | null {
  if (!status) return null;
  const devices = [status.server, ...(status.devices ?? [])];
  for (const device of devices) {
    if (device?.media?.title) return device.media;
  }
  return null;
}
