import { create } from "@bufbuild/protobuf";
import { createClient, type Client, type Transport } from "@connectrpc/connect";
import {
  DeleteTrackRequestSchema,
  EmptyTrackTrashRequestSchema,
  GetTrackRequestSchema,
  ImportTrackRequestSchema,
  ListPlaybackHistoryRequestSchema,
  ListTracksRequestSchema,
  MusicLibraryService,
  PurgeTrackRequestSchema,
  RecordPlaybackRequestSchema,
  RestoreTrackRequestSchema,
  SetTrackFavoriteRequestSchema,
} from "@cloud-drive/contracts";
import type {
  PlayableTrack,
  PlaybackEntry,
  Track,
} from "@cloud-drive/contracts";

import {
  normalizeBaseUrl,
  privateTransport,
  required,
  resolveApiUrl,
} from "./core";

export type TrackListOptions = Readonly<{
  query?: string;
  album?: string;
  artist?: string;
  favorites?: boolean;
  trashed?: boolean;
  pageSize?: number;
  pageToken?: string;
}>;

export type TrackListPage = Readonly<{
  tracks: Track[];
  nextPageToken: string;
}>;

export type PlaybackHistoryPage = Readonly<{
  entries: PlaybackEntry[];
  nextPageToken: string;
}>;

export class MusicLibraryClient {
  private readonly baseUrl: string;
  private readonly client: Client<typeof MusicLibraryService>;

  constructor(
    baseUrl: string,
    transport: Transport = privateTransport(baseUrl),
  ) {
    this.baseUrl = normalizeBaseUrl(baseUrl);
    this.client = createClient(MusicLibraryService, transport);
  }

  async trackPage(
    options: TrackListOptions = {},
    signal?: AbortSignal,
  ): Promise<TrackListPage> {
    const response = await this.client.listTracks(
      create(ListTracksRequestSchema, options),
      signal ? { signal } : undefined,
    );
    return { tracks: response.tracks, nextPageToken: response.nextPageToken };
  }

  async get(trackUid: string, signal?: AbortSignal): Promise<Track> {
    return required(
      (
        await this.client.getTrack(
          create(GetTrackRequestSchema, { trackUid }),
          signal ? { signal } : undefined,
        )
      ).track,
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

  async recordPlayback(track: PlayableTrack): Promise<void> {
    await this.client.recordPlayback(
      create(RecordPlaybackRequestSchema, { track }),
    );
  }

  async historyPage(
    pageToken = "",
    signal?: AbortSignal,
  ): Promise<PlaybackHistoryPage> {
    const response = await this.client.listPlaybackHistory(
      create(ListPlaybackHistoryRequestSchema, { pageSize: 50, pageToken }),
      signal ? { signal } : undefined,
    );
    return {
      entries: response.playbackEntries,
      nextPageToken: response.nextPageToken,
    };
  }

  contentUrl(track: Track): string {
    return resolveApiUrl(this.baseUrl, track.contentUrl);
  }

  artworkUrl(track: Track): string {
    return resolveApiUrl(this.baseUrl, track.artworkUrl);
  }
}
