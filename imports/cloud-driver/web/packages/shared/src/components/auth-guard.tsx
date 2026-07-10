import type { PropsWithChildren } from "react";
import { useEffect, useMemo, useState } from "react";
import { ShieldAlert } from "lucide-react";

import {
  SessionClient,
  authenticationUrl,
  isUnauthenticatedError,
} from "../api";
import { EmptyState, LoadingIndicator } from "./feedback";

type State = "checking" | "ready" | "failed";

export function AuthGuard({
  apiBase,
  authOrigin,
  children,
}: PropsWithChildren<{ apiBase: string; authOrigin: string }>) {
  const client = useMemo(() => new SessionClient(apiBase), [apiBase]);
  const [state, setState] = useState<State>("checking");
  const [message, setMessage] = useState("");

  useEffect(() => {
    const controller = new AbortController();
    void client
      .getSession(controller.signal)
      .then(() => setState("ready"))
      .catch((error: unknown) => {
        if (controller.signal.aborted) return;
        if (isUnauthenticatedError(error)) {
          window.location.replace(authenticationUrl(authOrigin));
          return;
        }
        setMessage(
          error instanceof Error ? error.message : "暂时无法连接服务。",
        );
        setState("failed");
      });
    return () => controller.abort();
  }, [authOrigin, client]);

  if (state === "checking")
    return (
      <div className="grid min-h-dvh place-items-center bg-background">
        <LoadingIndicator label="正在验证会话" />
      </div>
    );
  if (state === "failed")
    return (
      <div className="grid min-h-dvh place-items-center bg-background">
        <EmptyState
          icon={<ShieldAlert className="size-6" />}
          title="无法验证会话"
          detail={message}
        />
      </div>
    );
  return <>{children}</>;
}
