import { create, fromJson } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import type { Client } from "@connectrpc/connect";
import {
  AbandonUploadRequestSchema,
  ContentUploadService,
  FinalizeUploadRequestSchema,
  GetUploadRequestSchema,
  StartUploadRequestSchema,
  UploadStatus,
  WriteUploadChunkResponseSchema,
} from "@realtime-me/library-contracts";
import type { Upload } from "@realtime-me/library-contracts";

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
import { abortableDelay } from "./delay";

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

    if (upload.status === UploadStatus.ACTIVE) {
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
      upload = required(
        (
          await this.client.finalizeUpload(
            create(FinalizeUploadRequestSchema, { uploadUid: uid }),
            options.signal ? { signal: options.signal } : undefined,
          )
        ).upload,
        "upload",
      );
    }
    await this.waitUntilSealed(upload, options.signal);
    return uid;
  }

  async abandon(uploadUidValue: string, signal?: AbortSignal): Promise<void> {
    await this.client.abandonUpload(
      create(AbandonUploadRequestSchema, { uploadUid: uploadUidValue }),
      signal ? { signal } : undefined,
    );
  }

  private async start(file: File, signal?: AbortSignal) {
    const response = await this.client.startUpload(
      create(StartUploadRequestSchema, {
        fileName: file.name,
        contentType: file.type,
        totalSizeBytes: BigInt(file.size),
      }),
      signal ? { signal } : undefined,
    );
    return {
      upload: required(response.upload, "upload"),
      chunkUrl: response.chunkUrl,
    };
  }

  private async resume(uid: string, signal?: AbortSignal) {
    const response = await this.client.getUpload(
      create(GetUploadRequestSchema, { uploadUid: uid }),
      signal ? { signal } : undefined,
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
          signal: signal ?? null,
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
        await abortableDelay(350 * 2 ** (attempt - 1), signal);
        const recovered = required(
          (
            await this.client.getUpload(
              create(GetUploadRequestSchema, { uploadUid: uid }),
              signal ? { signal } : undefined,
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

  private async waitUntilSealed(
    initial: Upload,
    signal?: AbortSignal,
  ): Promise<void> {
    let upload = initial;
    for (;;) {
      if (
        upload.status === UploadStatus.SEALED ||
        upload.status === UploadStatus.CLAIMED
      )
        return;
      if (upload.status === UploadStatus.FAILED)
        throw new ApiError(
          upload.failureCode
            ? `Upload finalization failed (${upload.failureCode}).`
            : "Upload finalization failed.",
        );
      if (upload.status !== UploadStatus.FINALIZING)
        throw new ApiError("Upload is no longer available.");
      await abortableDelay(500, signal);
      upload = required(
        (
          await this.client.getUpload(
            create(GetUploadRequestSchema, { uploadUid: upload.uid }),
            signal ? { signal } : undefined,
          )
        ).upload,
        "upload",
      );
    }
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
