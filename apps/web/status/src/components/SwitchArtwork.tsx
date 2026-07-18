import { Gamepad2 } from 'lucide-react';
import { cn } from '@/lib/utils';

export function SwitchArtwork({ imageUri, className }: { imageUri?: string; className?: string }) {
  return (
    <span className={cn('relative flex shrink-0 items-center justify-center overflow-hidden rounded bg-muted', className)}>
      <Gamepad2 className="size-[55%] text-primary" />
      {!!imageUri && (
        <img
          src={imageUri}
          alt=""
          className="absolute inset-0 size-full object-cover"
          decoding="async"
          referrerPolicy="no-referrer"
          onError={(event) => {
            event.currentTarget.hidden = true;
          }}
        />
      )}
    </span>
  );
}
