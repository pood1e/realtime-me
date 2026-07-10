import { create, fromJson, toJson } from "@bufbuild/protobuf";
import type { JsonValue } from "@bufbuild/protobuf";
import { Code, ConnectError, createClient } from "@connectrpc/connect";
import type { Client } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import {
  CompleteUploadResponseSchema,
  CreateDirectoryRequestSchema,
  CreateShareLinkRequestSchema,
  DeleteDriveItemRequestSchema,
  DriveService,
  EmptyTrashRequestSchema,
  GetDownloadRequestSchema,
  GetDriveItemRequestSchema,
  GetSessionRequestSchema,
  GetUploadResponseSchema,
  ListDriveItemsRequestSchema,
  ListShareLinksRequestSchema,
  ListSharedItemsResponseSchema,
  ListTrashedItemsRequestSchema,
  MoveDriveItemRequestSchema,
  PurgeDriveItemRequestSchema,
  RenameDriveItemRequestSchema,
  ResolveShareResponseSchema,
  RestoreDriveItemRequestSchema,
  RevokeShareLinkRequestSchema,
  SearchDriveItemsRequestSchema,
  LoginRequestSchema,
  LogoutRequestSchema,
  SessionService,
  ShareService,
  StartUploadRequestSchema,
  StartUploadResponseSchema,
  WriteUploadChunkResponseSchema,
} from "@cloud-drive/contracts";
import type { DriveItem, ShareLink, Upload } from "@cloud-drive/contracts";

import { uploadChunkSize, uploadRanges, uploadReceivedBytes, uploadUid } from "./message";

const DEFAULT_TIMEOUT_MS = 30_000;
const MAX_UPLOAD_ATTEMPTS = 3;

export class ApiError extends Error {
  readonly status: number;

  constructor(message: string, status = 0) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

export type UploadProgress = Readonly<{
  uploadedBytes: number;
  totalBytes: number;
}>;

export type UploadOptions = Readonly<{
  parentUid: string;
  signal?: AbortSignal;
  resumeUploadUid?: string;
  onProgress?: (progress: UploadProgress) => void;
  onSession?: (uploadUid: string) => void;
}>;

export type CreatedShareLink = Readonly<{
  shareLink: ShareLink;
  shareUrl: string;
}>;

export type ResolvedShare = Readonly<{
  shareLink: ShareLink;
  target: DriveItem;
}>;

function normalizeBaseUrl(baseUrl: string): string {
  return baseUrl.replace(/\/+$/, "");
}

function resolveUrl(baseUrl: string, path: string): string {
  return new URL(path, `${normalizeBaseUrl(baseUrl)}/`).toString();
}

function required<T>(value: T | undefined, field: string): T {
  if (value === undefined) {
    throw new ApiError(`The API response is missing ${field}.`);
  }
  return value;
}

function isJsonObject(value: JsonValue): value is Record<string, JsonValue> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function errorMessage(value: JsonValue, fallback: string): string {
  if (!isJsonObject(value)) {
    return fallback;
  }

  const message = value["message"];
  if (typeof message === "string" && message) {
    return message;
  }
  const error = value["error"];
  return typeof error === "string" && error ? error : fallback;
}

function sleep(milliseconds: number, signal?: AbortSignal): Promise<void> {
  return new Promise((resolve, reject) => {
    const timeout = window.setTimeout(resolve, milliseconds);
    signal?.addEventListener(
      "abort",
      () => {
        window.clearTimeout(timeout);
        reject(signal.reason ?? new DOMException("The request was aborted.", "AbortError"));
      },
      { once: true },
    );
  });
}

function withTimeout(signal: AbortSignal | undefined, timeoutMs: number): { signal: AbortSignal; dispose: () => void } {
  const controller = new AbortController();
  const timeout = window.setTimeout(() => controller.abort(new DOMException("The request timed out.", "TimeoutError")), timeoutMs);
  const abort = () => controller.abort(signal?.reason ?? new DOMException("The request was aborted.", "AbortError"));

  if (signal?.aborted) {
    abort();
  } else {
    signal?.addEventListener("abort", abort, { once: true });
  }

  return {
    signal: controller.signal,
    dispose: () => {
      window.clearTimeout(timeout);
      signal?.removeEventListener("abort", abort);
    },
  };
}

async function parseJson(response: Response): Promise<JsonValue> {
  const body = await response.text();
  if (!body) {
    return {};
  }

  try {
    const parsed: JsonValue = JSON.parse(body);
    return parsed;
  } catch {
    return { message: body };
  }
}

async function requestJson(url: string, init: RequestInit, signal?: AbortSignal): Promise<JsonValue> {
  const timeout = withTimeout(signal, DEFAULT_TIMEOUT_MS);
  try {
    const response = await fetch(url, { ...init, signal: timeout.signal });
    const body = await parseJson(response);
    if (!response.ok) {
      throw new ApiError(errorMessage(body, `Request failed (${response.status}).`), response.status);
    }
    return body;
  } catch (error) {
    if (error instanceof ApiError || error instanceof DOMException) {
      throw error;
    }
    throw new ApiError(error instanceof Error ? error.message : "The network request failed.");
  } finally {
    timeout.dispose();
  }
}

function jsonRequest(body: JsonValue, credentials: RequestCredentials): RequestInit {
  return {
    method: "POST",
    credentials,
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  };
}

function rangeContains(ranges: ReadonlyArray<readonly [number, number]>, start: number, end: number): boolean {
  return ranges.some(([rangeStart, rangeEnd]) => rangeStart <= start && rangeEnd >= end);
}

function acknowledgedBytes(upload: Upload, totalBytes: number): number {
  const ranges = [...uploadRanges(upload)].sort(([leftStart], [rightStart]) => leftStart - rightStart);
  let coveredUntil = 0;
  for (const [start, end] of ranges) {
    if (start > coveredUntil) {
      break;
    }
    coveredUntil = Math.max(coveredUntil, end);
  }
  return Math.min(totalBytes, Math.max(uploadReceivedBytes(upload), coveredUntil));
}

function timestampFromDate(value: Date): { seconds: bigint; nanos: number } {
  const milliseconds = value.getTime();
  return {
    seconds: BigInt(Math.floor(milliseconds / 1_000)),
    nanos: (milliseconds % 1_000) * 1_000_000,
  };
}

export class PrivateDriveClient {
  private readonly baseUrl: string;
  private readonly drive: Client<typeof DriveService>;
  private readonly session: Client<typeof SessionService>;
  private readonly share: Client<typeof ShareService>;

  constructor(baseUrl: string) {
    this.baseUrl = normalizeBaseUrl(baseUrl);
    const transport = createConnectTransport({
      baseUrl: this.baseUrl,
      fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
      useBinaryFormat: false,
      defaultTimeoutMs: DEFAULT_TIMEOUT_MS,
    });
    this.drive = createClient(DriveService, transport);
    this.session = createClient(SessionService, transport);
    this.share = createClient(ShareService, transport);
  }

  async getSession(signal?: AbortSignal): Promise<void> {
    await this.session.getSession(create(GetSessionRequestSchema), { signal });
  }

  async login(password: string, signal?: AbortSignal): Promise<void> {
    await this.session.login(create(LoginRequestSchema, { password }), { signal });
  }

  async logout(signal?: AbortSignal): Promise<void> {
    await this.session.logout(create(LogoutRequestSchema), { signal });
  }

  async getDriveItem(itemUid: string, signal?: AbortSignal): Promise<DriveItem> {
    const response = await this.drive.getDriveItem(create(GetDriveItemRequestSchema, { itemUid }), { signal });
    return required(response.item, "item");
  }

  async listDriveItems(parentUid: string, includeTrashed: boolean, signal?: AbortSignal): Promise<DriveItem[]> {
    const response = await this.drive.listDriveItems(
      create(ListDriveItemsRequestSchema, { parentUid, includeTrashed, pageSize: 200, pageToken: "" }),
      { signal },
    );
    return response.items;
  }

  async listTrashedItems(signal?: AbortSignal): Promise<DriveItem[]> {
    const response = await this.drive.listTrashedItems(
      create(ListTrashedItemsRequestSchema, { pageSize: 200, pageToken: "" }),
      { signal },
    );
    return response.items;
  }

  async searchDriveItems(query: string, signal?: AbortSignal): Promise<DriveItem[]> {
    const response = await this.drive.searchDriveItems(
      create(SearchDriveItemsRequestSchema, { query, pageSize: 200, pageToken: "" }),
      { signal },
    );
    return response.items;
  }

  async createDirectory(parentUid: string, name: string, signal?: AbortSignal): Promise<DriveItem> {
    const response = await this.drive.createDirectory(create(CreateDirectoryRequestSchema, { parentUid, name }), { signal });
    return required(response.item, "item");
  }

  async renameDriveItem(itemUid: string, name: string, signal?: AbortSignal): Promise<DriveItem> {
    const response = await this.drive.renameDriveItem(create(RenameDriveItemRequestSchema, { itemUid, name }), { signal });
    return required(response.item, "item");
  }

  async moveDriveItem(itemUid: string, parentUid: string, signal?: AbortSignal): Promise<DriveItem> {
    const response = await this.drive.moveDriveItem(create(MoveDriveItemRequestSchema, { itemUid, parentUid }), { signal });
    return required(response.item, "item");
  }

  async deleteDriveItem(itemUid: string, signal?: AbortSignal): Promise<DriveItem> {
    const response = await this.drive.deleteDriveItem(create(DeleteDriveItemRequestSchema, { itemUid }), { signal });
    return required(response.item, "item");
  }

  async restoreDriveItem(itemUid: string, signal?: AbortSignal): Promise<DriveItem> {
    const response = await this.drive.restoreDriveItem(create(RestoreDriveItemRequestSchema, { itemUid }), { signal });
    return required(response.item, "item");
  }

  async purgeDriveItem(itemUid: string, signal?: AbortSignal): Promise<void> {
    await this.drive.purgeDriveItem(create(PurgeDriveItemRequestSchema, { itemUid }), { signal });
  }

  async emptyTrash(signal?: AbortSignal): Promise<void> {
    await this.drive.emptyTrash(create(EmptyTrashRequestSchema), { signal });
  }

  async createShareLink(targetUid: string, expireTime: Date, signal?: AbortSignal): Promise<CreatedShareLink> {
    const response = await this.share.createShareLink(
      create(CreateShareLinkRequestSchema, { targetUid, expireTime: timestampFromDate(expireTime) }),
      { signal },
    );
    return {
      shareLink: required(response.shareLink, "shareLink"),
      shareUrl: response.shareUrl,
    };
  }

  async listShareLinks(targetUid: string, signal?: AbortSignal): Promise<ShareLink[]> {
    const response = await this.share.listShareLinks(
      create(ListShareLinksRequestSchema, { targetUid, pageSize: 200, pageToken: "" }),
      { signal },
    );
    return response.shareLinks;
  }

  async revokeShareLink(shareUid: string, signal?: AbortSignal): Promise<ShareLink> {
    const response = await this.share.revokeShareLink(create(RevokeShareLinkRequestSchema, { shareUid }), { signal });
    return required(response.shareLink, "shareLink");
  }

  async getDownloadUrl(itemUid: string, signal?: AbortSignal): Promise<string> {
    const response = await this.drive.getDownload(create(GetDownloadRequestSchema, { itemUid }), { signal });
    return response.downloadUrl ? resolveUrl(this.baseUrl, response.downloadUrl) : this.contentUrl(itemUid);
  }

  contentUrl(itemUid: string): string {
    return resolveUrl(this.baseUrl, `/v1/items/${encodeURIComponent(itemUid)}/content`);
  }

  async readText(url: string, signal?: AbortSignal): Promise<string> {
    const timeout = withTimeout(signal, DEFAULT_TIMEOUT_MS);
    try {
      const response = await fetch(url, {
        credentials: "include",
        headers: { Range: "bytes=0-1048575" },
        signal: timeout.signal,
      });
      if (!response.ok) {
        throw new ApiError(`Unable to load preview (${response.status}).`, response.status);
      }
      return response.text();
    } finally {
      timeout.dispose();
    }
  }

  async uploadFile(file: File, options: UploadOptions): Promise<DriveItem> {
    let upload: Upload;
    let chunkUrl: string;

    if (options.resumeUploadUid) {
      upload = await this.getUpload(options.resumeUploadUid, options.signal);
      chunkUrl = `/v1/uploads/${encodeURIComponent(options.resumeUploadUid)}/chunks`;
    } else {
      const request = create(StartUploadRequestSchema, {
        parentUid: options.parentUid,
        fileName: file.name,
        contentType: file.type,
        totalSizeBytes: BigInt(file.size),
      });
      const response = fromJson(
        StartUploadResponseSchema,
        await requestJson(
          resolveUrl(this.baseUrl, "/v1/uploads"),
          jsonRequest(toJson(StartUploadRequestSchema, request), "include"),
          options.signal,
        ),
      );
      upload = required(response.upload, "upload");
      chunkUrl = response.chunkUrl || `/v1/uploads/${encodeURIComponent(uploadUid(upload))}/chunks`;
    }

    const uid = uploadUid(upload);
    if (!uid) {
      throw new ApiError("The upload session is missing its identifier.");
    }

    options.onSession?.(uid);
    const totalBytes = file.size;
    const chunkSize = Math.max(1, uploadChunkSize(upload));
    options.onProgress?.({ uploadedBytes: acknowledgedBytes(upload, totalBytes), totalBytes });

    for (let start = 0; start < totalBytes; start += chunkSize) {
      const end = Math.min(totalBytes, start + chunkSize);
      if (rangeContains(uploadRanges(upload), start, end)) {
        continue;
      }

      upload = await this.writeChunkWithRetry(upload, chunkUrl, file.slice(start, end), start, totalBytes, options.signal);
      options.onProgress?.({ uploadedBytes: acknowledgedBytes(upload, totalBytes), totalBytes });
    }

    const response = fromJson(
      CompleteUploadResponseSchema,
      await requestJson(
        resolveUrl(this.baseUrl, `/v1/uploads/${encodeURIComponent(uid)}/complete`),
        jsonRequest({}, "include"),
        options.signal,
      ),
    );
    return required(response.item, "item");
  }

  private async getUpload(uploadUidValue: string, signal?: AbortSignal): Promise<Upload> {
    const response = fromJson(
      GetUploadResponseSchema,
      await requestJson(
        resolveUrl(this.baseUrl, `/v1/uploads/${encodeURIComponent(uploadUidValue)}`),
        { method: "GET", credentials: "include", headers: { Accept: "application/json" } },
        signal,
      ),
    );
    return required(response.upload, "upload");
  }

  private async writeChunkWithRetry(
    currentUpload: Upload,
    chunkUrl: string,
    data: Blob,
    startOffset: number,
    totalSizeBytes: number,
    signal?: AbortSignal,
  ): Promise<Upload> {
    const uploadUidValue = uploadUid(currentUpload);
    const endOffset = startOffset + data.size;
    let lastError: unknown;

    for (let attempt = 1; attempt <= MAX_UPLOAD_ATTEMPTS; attempt += 1) {
      try {
        const response = fromJson(
          WriteUploadChunkResponseSchema,
          await requestJson(
            resolveUrl(this.baseUrl, chunkUrl),
            {
              method: "PUT",
              credentials: "include",
              headers: {
                Accept: "application/json",
                "Content-Type": "application/octet-stream",
                "Content-Range": `bytes ${startOffset}-${endOffset - 1}/${totalSizeBytes}`,
              },
              body: data,
            },
            signal,
          ),
        );
        return required(response.upload, "upload");
      } catch (error) {
        lastError = error;
        if (signal?.aborted || attempt === MAX_UPLOAD_ATTEMPTS) {
          break;
        }
        await sleep(350 * 2 ** (attempt - 1), signal);
        const recovered = await this.getUpload(uploadUidValue, signal);
        if (rangeContains(uploadRanges(recovered), startOffset, endOffset)) {
          return recovered;
        }
      }
    }

    throw lastError instanceof Error ? lastError : new ApiError("Unable to upload this chunk.");
  }
}

export class PublicShareClient {
  private readonly baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = normalizeBaseUrl(baseUrl);
  }

  async resolveShare(shareToken: string, signal?: AbortSignal): Promise<ResolvedShare> {
    const response = fromJson(
      ResolveShareResponseSchema,
      await this.get(`/v1/shares/${encodeURIComponent(shareToken)}`, signal),
    );
    return {
      shareLink: required(response.shareLink, "shareLink"),
      target: required(response.target, "target"),
    };
  }

  async listSharedItems(shareToken: string, parentUid: string, signal?: AbortSignal): Promise<DriveItem[]> {
    const query = new URLSearchParams({ parentUid, pageSize: "200" });
    const response = fromJson(
      ListSharedItemsResponseSchema,
      await this.get(`/v1/shares/${encodeURIComponent(shareToken)}/items?${query.toString()}`, signal),
    );
    return response.items;
  }

  contentUrl(shareToken: string, itemUid: string): string {
    return resolveUrl(
      this.baseUrl,
      `/v1/shares/${encodeURIComponent(shareToken)}/items/${encodeURIComponent(itemUid)}/content`,
    );
  }

  async readText(url: string, signal?: AbortSignal): Promise<string> {
    const timeout = withTimeout(signal, DEFAULT_TIMEOUT_MS);
    try {
      const response = await fetch(url, {
        credentials: "omit",
        headers: { Range: "bytes=0-1048575" },
        referrerPolicy: "no-referrer",
        signal: timeout.signal,
      });
      if (!response.ok) {
        throw new ApiError(`Unable to load preview (${response.status}).`, response.status);
      }
      return response.text();
    } finally {
      timeout.dispose();
    }
  }

  private get(path: string, signal?: AbortSignal): Promise<JsonValue> {
    return requestJson(
      resolveUrl(this.baseUrl, path),
      {
        method: "GET",
        credentials: "omit",
        headers: { Accept: "application/json" },
        referrerPolicy: "no-referrer",
      },
      signal,
    );
  }
}

export function apiBaseUrl(value: string | undefined, fallback: string): string {
  const configured = value?.trim();
  return normalizeBaseUrl(configured || fallback);
}

export function isUnavailableShareError(error: unknown): boolean {
  return error instanceof ApiError && [401, 403, 404, 410].includes(error.status);
}

export function isUnauthenticatedError(error: unknown): boolean {
  return (error instanceof ApiError && error.status === 401)
    || (error instanceof ConnectError && error.code === Code.Unauthenticated);
}
