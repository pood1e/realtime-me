import { useEffect, useRef } from "react";

const SEEK_STEP_SECONDS = 5;

export function usePlaybackShortcuts({
  enabled,
  position,
  onToggle,
  onPrevious,
  onNext,
  onSeek,
  onToggleMuted,
}: {
  enabled: boolean;
  position: number;
  onToggle: () => void;
  onPrevious: () => void;
  onNext: () => void;
  onSeek: (seconds: number) => void;
  onToggleMuted: () => void;
}) {
  const state = useRef({
    position,
    onToggle,
    onPrevious,
    onNext,
    onSeek,
    onToggleMuted,
  });
  state.current = {
    position,
    onToggle,
    onPrevious,
    onNext,
    onSeek,
    onToggleMuted,
  };
  useEffect(() => {
    if (!enabled) return;
    const handleKeyDown = (event: KeyboardEvent) => {
      if (
        event.repeat ||
        event.ctrlKey ||
        event.metaKey ||
        event.altKey ||
        isInteractive(event.target)
      )
        return;
      if (event.key === " ") {
        event.preventDefault();
        state.current.onToggle();
      } else if (event.key.toLowerCase() === "m") {
        state.current.onToggleMuted();
      } else if (event.key === "ArrowLeft" && event.shiftKey) {
        event.preventDefault();
        state.current.onPrevious();
      } else if (event.key === "ArrowRight" && event.shiftKey) {
        event.preventDefault();
        state.current.onNext();
      } else if (event.key === "ArrowLeft") {
        event.preventDefault();
        state.current.onSeek(state.current.position - SEEK_STEP_SECONDS);
      } else if (event.key === "ArrowRight") {
        event.preventDefault();
        state.current.onSeek(state.current.position + SEEK_STEP_SECONDS);
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [enabled]);
}

function isInteractive(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) return false;
  return Boolean(
    target.closest(
      "button,a[href],input,textarea,select,[contenteditable=true],[role=tab],[role=slider],[role=menuitem]",
    ),
  );
}
