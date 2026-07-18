import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import type {} from "@fastify/websocket";
import type { FastifyInstance } from "fastify";
import { type RawData, WebSocket } from "ws";
import type { TerminalAttachment, TerminalManager } from "../application/terminal-manager.js";
import {
  TerminalClientFrameSchema,
  type TerminalServerFrame,
  TerminalServerFrameSchema,
} from "../gen/super_manager/terminal/v1/terminal_pb.js";

const ATTACH_TIMEOUT_MS = 5_000;
const MAX_BUFFERED_BYTES = 4 * 1024 * 1024;
const MAX_OUTPUT_FRAME_BYTES = 1024 * 1024;

export function registerTerminalWebSocket(
  server: FastifyInstance,
  terminals: TerminalManager,
): void {
  server.get("/v1/terminal", { websocket: true }, (socket) => {
    let attachment: TerminalAttachment | null = null;
    let terminalUid: string | null = null;
    let closed = false;
    let processing = Promise.resolve();
    const attachTimeout = setTimeout(() => {
      sendError(socket, "ATTACH_TIMEOUT", "The first frame must attach a terminal session");
      socket.close(1008, "attach timeout");
    }, ATTACH_TIMEOUT_MS);

    socket.on("message", (data, isBinary) => {
      processing = processing
        .then(async () => {
          if (!isBinary) {
            throw new ProtocolError("BINARY_REQUIRED", "Terminal frames must be binary Protobuf");
          }
          const bytes = toUint8Array(data);
          if (bytes.byteLength > MAX_OUTPUT_FRAME_BYTES + 1024) {
            throw new ProtocolError("FRAME_TOO_LARGE", "Terminal frame exceeded the size limit");
          }
          const frame = fromBinary(TerminalClientFrameSchema, bytes);
          const payload = frame.payload;
          if (!attachment) {
            if (payload.case !== "attach") {
              throw new ProtocolError("ATTACH_REQUIRED", "The first frame must be attach");
            }
            clearTimeout(attachTimeout);
            terminalUid = payload.value.terminalSessionUid;
            const session = await terminals.get(terminalUid);
            if (session?.state !== "RUNNING") {
              throw new ProtocolError("NOT_FOUND", "Terminal session is not running");
            }
            const nextAttachment = await terminals.attach(terminalUid, {
              output: (output) => {
                for (let offset = 0; offset < output.byteLength; offset += MAX_OUTPUT_FRAME_BYTES) {
                  send(
                    socket,
                    create(TerminalServerFrameSchema, {
                      payload: {
                        case: "output",
                        value: { data: output.slice(offset, offset + MAX_OUTPUT_FRAME_BYTES) },
                      },
                    }),
                    () => attachment?.detach(),
                  );
                }
              },
              exited: (exitCode, signal) => {
                send(
                  socket,
                  create(TerminalServerFrameSchema, {
                    payload: {
                      case: "exited",
                      value: {
                        ...(exitCode === null ? {} : { exitCode }),
                        signal: signal === null ? "" : String(signal),
                      },
                    },
                  }),
                );
                socket.close(1000, "terminal attachment exited");
              },
            });
            if (closed || socket.readyState !== WebSocket.OPEN) {
              nextAttachment.detach();
              return;
            }
            attachment = nextAttachment;
            send(
              socket,
              create(TerminalServerFrameSchema, {
                payload: {
                  case: "attached",
                  value: {
                    terminalSessionUid: terminalUid,
                    columns: session.columns,
                    rows: session.rows,
                  },
                },
              }),
            );
            return;
          }
          switch (payload.case) {
            case "input":
              if (payload.value.data.byteLength === 0 || payload.value.data.byteLength > 65_536) {
                throw new ProtocolError("INVALID_INPUT", "Terminal input size is invalid");
              }
              attachment.write(payload.value.data);
              break;
            case "resize":
              attachment.resize(payload.value.columns, payload.value.rows);
              break;
            case "detach":
              attachment.detach();
              attachment = null;
              socket.close(1000, "detached");
              break;
            case "close":
              if (terminalUid) {
                await terminals.delete(terminalUid);
              }
              attachment = null;
              socket.close(1000, "closed");
              break;
            case "ping":
              send(
                socket,
                create(TerminalServerFrameSchema, {
                  payload: { case: "pong", value: { nonce: payload.value.nonce } },
                }),
              );
              break;
            case "attach":
              throw new ProtocolError("ALREADY_ATTACHED", "The socket is already attached");
            default:
              throw new ProtocolError("EMPTY_FRAME", "Terminal frame has no payload");
          }
        })
        .catch((error: unknown) => {
          const protocolError =
            error instanceof ProtocolError
              ? error
              : new ProtocolError("PROTOCOL_ERROR", boundedMessage(error));
          sendError(socket, protocolError.code, protocolError.message);
          socket.close(1008, protocolError.code);
        });
    });

    socket.on("close", () => {
      if (closed) {
        return;
      }
      closed = true;
      clearTimeout(attachTimeout);
      attachment?.detach();
      attachment = null;
    });
  });
}

function send(socket: WebSocket, frame: TerminalServerFrame, onBackpressure?: () => void): void {
  if (socket.readyState !== WebSocket.OPEN) {
    return;
  }
  if (socket.bufferedAmount > MAX_BUFFERED_BYTES) {
    onBackpressure?.();
    socket.close(1013, "terminal client is too slow");
    return;
  }
  socket.send(toBinary(TerminalServerFrameSchema, frame), { binary: true });
}

function sendError(socket: WebSocket, code: string, message: string): void {
  send(
    socket,
    create(TerminalServerFrameSchema, {
      payload: {
        case: "error",
        value: { code: code.slice(0, 64), message: message.slice(0, 512) },
      },
    }),
  );
}

function toUint8Array(data: RawData): Uint8Array {
  if (data instanceof ArrayBuffer) {
    return new Uint8Array(data);
  }
  if (Buffer.isBuffer(data)) {
    return data;
  }
  return Buffer.concat([...data]);
}

class ProtocolError extends Error {
  constructor(
    readonly code: string,
    message: string,
  ) {
    super(message);
  }
}

function boundedMessage(error: unknown): string {
  return (error instanceof Error ? error.message : String(error)).slice(0, 512);
}
