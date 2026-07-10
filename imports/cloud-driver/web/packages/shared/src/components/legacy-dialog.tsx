import type { PropsWithChildren } from "react";

import { cn } from "../lib/utils";
import {
  Dialog as DialogRoot,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "./ui/dialog";

export type DialogSize = "compact" | "standard" | "preview";

const dialogSizes: Record<DialogSize, string> = {
  compact: "sm:max-w-md",
  standard: "sm:max-w-2xl",
  preview: "sm:max-w-5xl",
};

export function Dialog({
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
      <DialogContent
        className={cn(
          "max-h-[calc(100dvh-1rem)] overflow-y-auto",
          dialogSizes[size],
        )}
      >
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description ? (
            <DialogDescription>{description}</DialogDescription>
          ) : null}
        </DialogHeader>
        {children}
      </DialogContent>
    </DialogRoot>
  );
}
