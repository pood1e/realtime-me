import { useCallback, useEffect, useState } from "react";
import { PLAYBACK_MODES, type PlaybackMode } from "./playback-types";

const STORAGE_KEY = "cloud-drive.music.playback-settings";
const DEFAULT_VOLUME = 0.8;

type PlaybackSettings = Readonly<{
  mode: PlaybackMode;
  volume: number;
  muted: boolean;
}>;

const DEFAULT_SETTINGS: PlaybackSettings = {
  mode: "sequential",
  volume: DEFAULT_VOLUME,
  muted: false,
};

export function usePlaybackSettings() {
  const [settings, setSettings] = useState(readSettings);

  useEffect(() => {
    try {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
    } catch {
      // Playback remains functional when browser storage is unavailable.
    }
  }, [settings]);

  const setMode = useCallback((mode: PlaybackMode) => {
    setSettings((current) => ({ ...current, mode }));
  }, []);
  const setVolume = useCallback((volume: number) => {
    const normalized = clampVolume(volume);
    setSettings((current) => ({
      ...current,
      volume: normalized,
      muted: normalized === 0,
    }));
  }, []);
  const toggleMuted = useCallback(() => {
    setSettings((current) => ({
      ...current,
      muted: !current.muted,
      volume:
        current.muted && current.volume === 0 ? DEFAULT_VOLUME : current.volume,
    }));
  }, []);

  return { ...settings, setMode, setVolume, toggleMuted };
}

function readSettings(): PlaybackSettings {
  try {
    const value: unknown = JSON.parse(
      window.localStorage.getItem(STORAGE_KEY) ?? "null",
    );
    if (!value || typeof value !== "object") return DEFAULT_SETTINGS;
    const record = value as Record<string, unknown>;
    return {
      mode: isPlaybackMode(record.mode) ? record.mode : DEFAULT_SETTINGS.mode,
      volume:
        typeof record.volume === "number"
          ? clampVolume(record.volume)
          : DEFAULT_SETTINGS.volume,
      muted:
        typeof record.muted === "boolean"
          ? record.muted
          : DEFAULT_SETTINGS.muted,
    };
  } catch {
    return DEFAULT_SETTINGS;
  }
}

function isPlaybackMode(value: unknown): value is PlaybackMode {
  return PLAYBACK_MODES.some((mode) => mode === value);
}

function clampVolume(volume: number): number {
  return Math.min(1, Math.max(0, Number.isFinite(volume) ? volume : 0));
}
