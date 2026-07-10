import { fromJson } from "@bufbuild/protobuf";
import type { JsonValue } from "@bufbuild/protobuf";
import {
  ListSharedItemsResponseSchema,
  ResolveShareResponseSchema,
} from "@cloud-drive/contracts";
import type { DriveItem, ShareLink } from "@cloud-drive/contracts";

import { ApiError, normalizeBaseUrl, required, resolveApiUrl } from "./core";

export class PublicShareClient {
  private readonly baseUrl: string;
  constructor(baseUrl: string) {
    this.baseUrl = normalizeBaseUrl(baseUrl);
  }
  async resolveShare(
    token: string,
    signal?: AbortSignal,
  ): Promise<ResolvedShare> {
    const message = fromJson(
      ResolveShareResponseSchema,
      await this.get(`/v1/shares/${encodeURIComponent(token)}`, signal),
    );
    return {
      shareLink: required(message.shareLink, "shareLink"),
      target: required(message.target, "target"),
    };
  }
  async listSharedItems(
    token: string,
    parentUid: string,
    signal?: AbortSignal,
  ): Promise<DriveItem[]> {
    const query = new URLSearchParams({ parentUid, pageSize: "200" });
    return fromJson(
      ListSharedItemsResponseSchema,
      await this.get(
        `/v1/shares/${encodeURIComponent(token)}/items?${query}`,
        signal,
      ),
    ).items;
  }
  contentUrl(token: string, itemUid: string): string {
    return resolveApiUrl(
      this.baseUrl,
      `/v1/shares/${encodeURIComponent(token)}/items/${encodeURIComponent(itemUid)}/content`,
    );
  }
  private async get(path: string, signal?: AbortSignal): Promise<JsonValue> {
    const response = await fetch(resolveApiUrl(this.baseUrl, path), {
      credentials: "omit",
      headers: { Accept: "application/json" },
      referrerPolicy: "no-referrer",
      signal,
    });
    if (!response.ok)
      throw new ApiError(
        `Request failed (${response.status}).`,
        response.status,
      );
    return response.json() as Promise<JsonValue>;
  }
}

export type ResolvedShare = Readonly<{
  shareLink: ShareLink;
  target: DriveItem;
}>;
