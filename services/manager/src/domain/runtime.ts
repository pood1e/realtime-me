import type { AGUIEvent, AgentCapabilities } from "@ag-ui/core";

export const CODEX_RUNTIME_UID = "11111111-1111-4111-8111-111111111111";
export const CLAUDE_RUNTIME_UID = "22222222-2222-4222-8222-222222222222";

export type ProviderKind = "codex" | "claude";
export type RuntimeAvailability =
  | "AVAILABLE"
  | "NOT_INSTALLED"
  | "NOT_AUTHENTICATED"
  | "INCOMPATIBLE"
  | "UNHEALTHY";

export interface RuntimeStatus {
  readonly uid: string;
  readonly kind: ProviderKind;
  readonly displayName: string;
  readonly version: string;
  readonly availability: RuntimeAvailability;
  readonly diagnostic: string;
  readonly capabilities: AgentCapabilities;
  readonly updateTime: Date;
}

export interface ProviderRun {
  readonly executionUid: string;
  readonly threadUid: string;
  readonly workspacePath: string;
  readonly providerSessionId: string | null;
  readonly prompt: string;
}

export interface StructuredQuestion {
  readonly id: string;
  readonly header: string;
  readonly question: string;
  readonly options: readonly string[];
  readonly multiple: boolean;
  readonly secret: boolean;
  readonly allowOther: boolean;
}

export interface InputRequest {
  readonly providerRequestId: string;
  readonly toolCallId: string;
  readonly questions: readonly StructuredQuestion[];
  readonly providerInput: unknown;
}

export interface InputAnswer {
  readonly answers: Readonly<Record<string, readonly string[]>>;
}

export interface ProviderSink {
  emit(event: AGUIEvent): Promise<void>;
  setProviderSession(providerSessionId: string): Promise<void>;
  setNativeTurn(nativeTurnId: string): Promise<void>;
  requestInput(request: InputRequest): Promise<InputAnswer>;
}

export interface RuntimeAdapter {
  readonly kind: ProviderKind;
  probe(): Promise<RuntimeStatus>;
  execute(run: ProviderRun, sink: ProviderSink): Promise<void>;
  cancel(executionUid: string): Promise<void>;
  steer(executionUid: string, instruction: string): Promise<void>;
  close(): Promise<void>;
}
