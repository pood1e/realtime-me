import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import {
  MetricsService,
  ProfileService,
  ProjectsService,
  StatusService,
} from "@realtime-me/status-contracts";

// POLL_INTERVAL_MS is the cadence for the public and internal status loops.
export const POLL_INTERVAL_MS = 10_000;

// createStatusApi binds the Status feature to one explicit application boundary.
export function createStatusApi(baseUrl: string) {
  const transport = createConnectTransport({ baseUrl: baseUrl.replace(/\/+$/, "") });
  return {
    status: createClient(StatusService, transport),
    profile: createClient(ProfileService, transport),
    projects: createClient(ProjectsService, transport),
    metrics: createClient(MetricsService, transport),
  } as const;
}

export type StatusApi = ReturnType<typeof createStatusApi>;
