import { cn } from "@realtime-me/web-ui/cn";
import {
  DialogContent,
  DialogDescription,
  DialogHeader,
  Dialog as DialogRoot,
  DialogTitle,
} from "@realtime-me/web-ui/dialog";
import type { PropsWithChildren } from "react";

export type DialogSize = "compact" | "standard" | "preview";

const dialogSizes: Record<DialogSize, string> = {
  compact: "sm:max-w-md",
  standard: "sm:max-w-2xl",
  preview: "sm:max-w-5xl",
};

export function AppDialog({
  open,
  title,
  description,
  size = "compact",
  onClose,
  children,
}: PropsWithChildren<{
  open: boolean;
  title: string;
  description?: string;
  size?: DialogSize;
  onClose: () => void;
}>) {
  return (
    <DialogRoot
      open={open}
      onOpenChange={(nextOpen) => {
        if (!nextOpen) onClose();
      }}
    >
      <DialogContent className={cn("max-h-[calc(100dvh-1rem)] overflow-y-auto", dialogSizes[size])}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description ? <DialogDescription>{description}</DialogDescription> : null}
        </DialogHeader>
        {children}
      </DialogContent>
    </DialogRoot>
  );
}
