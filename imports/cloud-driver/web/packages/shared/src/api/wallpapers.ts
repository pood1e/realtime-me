import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import type { Client } from "@connectrpc/connect";
import {
  ListPublishedWallpapersRequestSchema,
  ListWallpapersRequestSchema,
  PublishWallpaperRequestSchema,
  UnpublishWallpaperRequestSchema,
  UpdateWallpaperRequestSchema,
  WallpaperAdminService,
  WallpaperPublicService,
} from "@cloud-drive/contracts";
import type { Wallpaper, WallpaperOrientation } from "@cloud-drive/contracts";

import {
  normalizeBaseUrl,
  privateTransport,
  publicTransport,
  required,
  resolveApiUrl,
} from "./core";

export type WallpaperPage = Readonly<{
  wallpapers: Wallpaper[];
  nextPageToken: string;
}>;

export class WallpaperAdminClient {
  private readonly client: Client<typeof WallpaperAdminService>;
  constructor(baseUrl: string) {
    this.client = createClient(
      WallpaperAdminService,
      privateTransport(baseUrl),
    );
  }
  async listPage(pageToken = "", signal?: AbortSignal): Promise<WallpaperPage> {
    const response = await this.client.listPublishedWallpapers(
      create(ListPublishedWallpapersRequestSchema, {
        pageSize: 60,
        pageToken,
      }),
      signal ? { signal } : undefined,
    );
    return {
      wallpapers: response.wallpapers,
      nextPageToken: response.nextPageToken,
    };
  }
  async publish(
    imageUid: string,
    title: string,
    tags: string[],
  ): Promise<Wallpaper> {
    return required(
      (
        await this.client.publishWallpaper(
          create(PublishWallpaperRequestSchema, { imageUid, title, tags }),
        )
      ).wallpaper,
      "wallpaper",
    );
  }
  async update(
    wallpaperUid: string,
    title: string,
    tags: string[],
  ): Promise<Wallpaper> {
    return required(
      (
        await this.client.updateWallpaper(
          create(UpdateWallpaperRequestSchema, { wallpaperUid, title, tags }),
        )
      ).wallpaper,
      "wallpaper",
    );
  }
  async unpublish(wallpaperUid: string): Promise<void> {
    await this.client.unpublishWallpaper(
      create(UnpublishWallpaperRequestSchema, { wallpaperUid }),
    );
  }
}

export class WallpaperPublicClient {
  readonly baseUrl: string;
  private readonly client: Client<typeof WallpaperPublicService>;
  constructor(baseUrl: string) {
    this.baseUrl = normalizeBaseUrl(baseUrl);
    this.client = createClient(
      WallpaperPublicService,
      publicTransport(baseUrl),
    );
  }
  async listPage(
    options: {
      query?: string;
      tag?: string;
      orientation?: WallpaperOrientation;
      pageToken?: string;
    } = {},
    signal?: AbortSignal,
  ): Promise<WallpaperPage> {
    const response = await this.client.listWallpapers(
      create(ListWallpapersRequestSchema, { ...options, pageSize: 60 }),
      signal ? { signal } : undefined,
    );
    return {
      wallpapers: response.wallpapers,
      nextPageToken: response.nextPageToken,
    };
  }
  assetUrl(path: string): string {
    return resolveApiUrl(this.baseUrl, path);
  }
}
