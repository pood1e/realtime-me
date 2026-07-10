import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import type { Client } from "@connectrpc/connect";
import {
  DeleteTrackRequestSchema,
  EmptyTrackTrashRequestSchema,
  GetTrackRequestSchema,
  ImportTrackRequestSchema,
  ListAlbumsRequestSchema,
  ListArtistsRequestSchema,
  ListPlaybackHistoryRequestSchema,
  ListTracksRequestSchema,
  MusicService,
  PurgeTrackRequestSchema,
  RecordPlaybackRequestSchema,
  RestoreTrackRequestSchema,
  SetTrackFavoriteRequestSchema,
} from "@cloud-drive/contracts";
import type {
  Album,
  Artist,
  PlaybackEntry,
  Track,
} from "@cloud-drive/contracts";

import {
  normalizeBaseUrl,
  privateTransport,
  required,
  resolveApiUrl,
} from "./core";

export class MusicClient {
  readonly baseUrl: string;
  private readonly client: Client<typeof MusicService>;
  constructor(baseUrl: string) {
    this.baseUrl = normalizeBaseUrl(baseUrl);
    this.client = createClient(MusicService, privateTransport(baseUrl));
  }
  async tracks(
    options: {
      query?: string;
      album?: string;
      artist?: string;
      favorites?: boolean;
      trashed?: boolean;
    } = {},
  ): Promise<Track[]> {
    return (
      await this.client.listTracks(
        create(ListTracksRequestSchema, { ...options, pageSize: 200 }),
      )
    ).tracks;
  }
  async get(trackUid: string): Promise<Track> {
    return required(
      (await this.client.getTrack(create(GetTrackRequestSchema, { trackUid })))
        .track,
      "track",
    );
  }
  async importUpload(uploadUid: string): Promise<Track> {
    return required(
      (
        await this.client.importTrack(
          create(ImportTrackRequestSchema, { uploadUid }),
        )
      ).track,
      "track",
    );
  }
  async favorite(trackUid: string, favorite: boolean): Promise<Track> {
    return required(
      (
        await this.client.setTrackFavorite(
          create(SetTrackFavoriteRequestSchema, { trackUid, favorite }),
        )
      ).track,
      "track",
    );
  }
  async trash(trackUid: string): Promise<void> {
    await this.client.deleteTrack(
      create(DeleteTrackRequestSchema, { trackUid }),
    );
  }
  async restore(trackUid: string): Promise<void> {
    await this.client.restoreTrack(
      create(RestoreTrackRequestSchema, { trackUid }),
    );
  }
  async purge(trackUid: string): Promise<void> {
    await this.client.purgeTrack(create(PurgeTrackRequestSchema, { trackUid }));
  }
  async emptyTrash(): Promise<void> {
    await this.client.emptyTrackTrash(create(EmptyTrackTrashRequestSchema));
  }
  async albums(query = ""): Promise<Album[]> {
    return (
      await this.client.listAlbums(create(ListAlbumsRequestSchema, { query }))
    ).albums;
  }
  async artists(query = ""): Promise<Artist[]> {
    return (
      await this.client.listArtists(create(ListArtistsRequestSchema, { query }))
    ).artists;
  }
  async recordPlayback(trackUid: string): Promise<void> {
    await this.client.recordPlayback(
      create(RecordPlaybackRequestSchema, { trackUid }),
    );
  }
  async history(): Promise<PlaybackEntry[]> {
    return (
      await this.client.listPlaybackHistory(
        create(ListPlaybackHistoryRequestSchema, { pageSize: 100 }),
      )
    ).playbackEntries;
  }
  contentUrl(track: Track): string {
    return resolveApiUrl(this.baseUrl, track.contentUrl);
  }
  artworkUrl(track: Track): string {
    return resolveApiUrl(this.baseUrl, track.artworkUrl);
  }
}
