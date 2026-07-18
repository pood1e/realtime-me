import type { PlaybackDescriptor } from "@realtime-me/library-contracts";
import { LOCAL_PROVIDER_ID, type MusicClient } from "@realtime-me/library-web";
import { DirectAudioPlayer } from "./direct-audio-player";
import type { PlaybackAdapter, PlaybackAdapterEvents } from "./playback-types";
import { createProviderPlayer } from "./provider-player-registry";

export function createPlaybackAdapter(
  descriptor: PlaybackDescriptor,
  client: MusicClient,
  events: PlaybackAdapterEvents,
): PlaybackAdapter | undefined {
  if (descriptor.playback.case === "directAudio") {
    return new DirectAudioPlayer(
      descriptor.providerId === LOCAL_PROVIDER_ID,
      events,
    );
  }
  if (descriptor.playback.case === "providerSdk") {
    return createProviderPlayer(
      descriptor.playback.value.sdkId,
      client,
      events,
    );
  }
  return undefined;
}

export function playbackResource(
  descriptor: PlaybackDescriptor,
  client: MusicClient,
): string {
  if (descriptor.playback.case === "directAudio")
    return client.providers.playbackUrl(descriptor.playback.value.url);
  if (descriptor.playback.case === "providerSdk")
    return descriptor.playback.value.resourceUri;
  throw new Error("播放来源未提供可用资源");
}
