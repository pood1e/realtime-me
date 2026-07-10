import { create, fromJson } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import type { Client } from "@connectrpc/connect";
import {
  AbandonUploadRequestSchema,
  ContentUploadService,
  GetUploadRequestSchema,
  StartUploadRequestSchema,
  WriteUploadChunkResponseSchema,
} from "@cloud-drive/contracts";
import type { Upload } from "@cloud-drive/contracts";

import {
  uploadChunkSize,
  uploadRanges,
  uploadReceivedBytes,
  uploadUid,
} from "../message";
import {
  ApiError,
  normalizeBaseUrl,
  privateTransport,
  required,
  resolveApiUrl,
} from "./core";

const MAX_UPLOAD_ATTEMPTS = 3;

export type UploadProgress = Readonly<{
  uploadedBytes: number;
  totalBytes: number;
}>;
export type UploadOptions = Readonly<{
  signal?: AbortSignal;
  resumeUploadUid?: string;
  onProgress?: (progress: UploadProgress) => void;
  onSession?: (uploadUid: string) => void;
}>;

export class UploadClient {
  private readonly baseUrl: string;
  private readonly client: Client<typeof ContentUploadService>;

  constructor(baseUrl: string) {
    this.baseUrl = normalizeBaseUrl(baseUrl);
    this.client = createClient(ContentUploadService, privateTransport(baseUrl));
  }

  async upload(file: File, options: UploadOptions = {}): Promise<string> {
    const state = options.resumeUploadUid
      ? await this.resume(options.resumeUploadUid, options.signal)
      : await this.start(file, options.signal);
    let upload = state.upload;
    const uid = uploadUid(upload);
    options.onSession?.(uid);
    options.onProgress?.({
      uploadedBytes: acknowledgedBytes(upload, file.size),
      totalBytes: file.size,
    });

    const chunkSize = Math.max(1, uploadChunkSize(upload));
    for (let start = 0; start < file.size; start += chunkSize) {
      const end = Math.min(file.size, start + chunkSize);
      if (rangeContains(uploadRanges(upload), start, end)) continue;
      upload = await this.writeChunkWithRetry(
        upload,
        state.chunkUrl,
        file.slice(start, end),
        start,
        file.size,
        options.signal,
      );
      options.onProgress?.({
        uploadedBytes: acknowledgedBytes(upload, file.size),
        totalBytes: file.size,
      });
    }
    return uid;
  }

  async abandon(uploadUidValue: string, signal?: AbortSignal): Promise<void> {
    await this.client.abandonUpload(
      create(AbandonUploadRequestSchema, { uploadUid: uploadUidValue }),
      { signal },
    );
  }

  private async start(file: File, signal?: AbortSignal) {
    const response = await this.client.startUpload(
      create(StartUploadRequestSchema, {
        fileName: file.name,
        contentType: file.type,
        totalSizeBytes: BigInt(file.size),
      }),
      { signal },
    );
    return {
      upload: required(response.upload, "upload"),
      chunkUrl: response.chunkUrl,
    };
  }

  private async resume(uid: string, signal?: AbortSignal) {
    const response = await this.client.getUpload(
      create(GetUploadRequestSchema, { uploadUid: uid }),
      { signal },
    );
    return {
      upload: required(response.upload, "upload"),
      chunkUrl: `/v1/uploads/${encodeURIComponent(uid)}/chunks`,
    };
  }

  private async writeChunkWithRetry(
    current: Upload,
    chunkUrl: string,
    data: Blob,
    start: number,
    total: number,
    signal?: AbortSignal,
  ): Promise<Upload> {
    const uid = uploadUid(current);
    const end = start + data.size;
    let lastError: unknown;
    for (let attempt = 1; attempt <= MAX_UPLOAD_ATTEMPTS; attempt += 1) {
      try {
        const response = await fetch(resolveApiUrl(this.baseUrl, chunkUrl), {
          method: "PUT",
          credentials: "include",
          headers: {
            Accept: "application/json",
            "Content-Type": "application/octet-stream",
            "Content-Range": `bytes ${start}-${end - 1}/${total}`,
          },
          body: data,
          signal,
        });
        if (!response.ok)
          throw new ApiError(
            `Upload failed (${response.status}).`,
            response.status,
          );
        return required(
          fromJson(WriteUploadChunkResponseSchema, await response.json())
            .upload,
          "upload",
        );
      } catch (error) {
        lastError = error;
        if (signal?.aborted || attempt === MAX_UPLOAD_ATTEMPTS) break;
        await sleep(350 * 2 ** (attempt - 1), signal);
        const recovered = required(
          (
            await this.client.getUpload(
              create(GetUploadRequestSchema, { uploadUid: uid }),
              { signal },
            )
          ).upload,
          "upload",
        );
        if (rangeContains(uploadRanges(recovered), start, end))
          return recovered;
      }
    }
    throw lastError instanceof Error
      ? lastError
      : new ApiError("Unable to upload this chunk.");
  }
}

function rangeContains(
  ranges: ReadonlyArray<readonly [number, number]>,
  start: number,
  end: number,
): boolean {
  return ranges.some(
    ([rangeStart, rangeEnd]) => rangeStart <= start && rangeEnd >= end,
  );
}

function acknowledgedBytes(upload: Upload, totalBytes: number): number {
  const ranges = [...uploadRanges(upload)].sort(
    ([left], [right]) => left - right,
  );
  let covered = 0;
  for (const [start, end] of ranges) {
    if (start > covered) break;
    covered = Math.max(covered, end);
  }
  return Math.min(totalBytes, Math.max(uploadReceivedBytes(upload), covered));
}

function sleep(milliseconds: number, signal?: AbortSignal): Promise<void> {
  return new Promise((resolve, reject) => {
    const timeout = window.setTimeout(resolve, milliseconds);
    signal?.addEventListener(
      "abort",
      () => {
        window.clearTimeout(timeout);
        reject(signal.reason);
      },
      { once: true },
    );
  });
}
