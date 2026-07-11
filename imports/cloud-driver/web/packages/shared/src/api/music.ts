import { privateTransport } from "./core";
import { MusicLibraryClient } from "./music-library";
import { MusicPlaylistClient } from "./music-playlists";
import { MusicProviderClient } from "./music-providers";

export * from "./music-library";
export * from "./music-playlists";
export * from "./music-providers";

// MusicClient is the composition root for the three independently addressable services.
export class MusicClient {
  readonly library: MusicLibraryClient;
  readonly providers: MusicProviderClient;
  readonly playlists: MusicPlaylistClient;

  constructor(baseUrl: string) {
    const transport = privateTransport(baseUrl);
    this.library = new MusicLibraryClient(baseUrl, transport);
    this.providers = new MusicProviderClient(baseUrl, transport);
    this.playlists = new MusicPlaylistClient(baseUrl, transport);
  }
}
