import type { PropsWithChildren } from "react";
import { useEffect, useMemo, useState } from "react";
import { RefreshCw, ShieldAlert } from "lucide-react";

import {
  SessionClient,
  authenticationUrl,
  isUnauthenticatedError,
} from "../api";
import { EmptyState, LoadingIndicator } from "./feedback";
import { Button } from "./ui/button";

type State = "checking" | "ready" | "failed";
const feedbackDelayMs = 600;
const validationTimeoutMs = 10_000;

export function AuthGuard({
  apiBase,
  authOrigin,
  children,
}: PropsWithChildren<{ apiBase: string; authOrigin: string }>) {
  const client = useMemo(() => new SessionClient(apiBase), [apiBase]);
  const [state, setState] = useState<State>("checking");
  const [showProgress, setShowProgress] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    const controller = new AbortController();
    let disposed = false;
    let timedOut = false;
    setState("checking");
    setShowProgress(false);
    setMessage("");
    const feedbackTimer = window.setTimeout(() => {
      if (!disposed) setShowProgress(true);
    }, feedbackDelayMs);
    const timeoutTimer = window.setTimeout(() => {
      if (disposed) return;
      timedOut = true;
      controller.abort();
      setMessage("会话验证超时，请重新验证。");
      setState("failed");
    }, validationTimeoutMs);
    const clearTimers = () => {
      window.clearTimeout(feedbackTimer);
      window.clearTimeout(timeoutTimer);
    };
    void client
      .getSession(controller.signal)
      .then(() => {
        if (disposed || timedOut) return;
        clearTimers();
        setState("ready");
      })
      .catch((error: unknown) => {
        if (disposed || timedOut) return;
        clearTimers();
        if (isUnauthenticatedError(error)) {
          window.location.replace(authenticationUrl(authOrigin));
          return;
        }
        setMessage(
          error instanceof Error ? error.message : "暂时无法连接服务。",
        );
        setState("failed");
      });
    return () => {
      disposed = true;
      clearTimers();
      controller.abort();
    };
  }, [authOrigin, client]);

  if (state === "checking" && !showProgress)
    return <div className="min-h-dvh bg-background" aria-hidden="true" />;
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
          action={
            <Button variant="outline" onClick={() => window.location.reload()}>
              <RefreshCw />
              重新验证
            </Button>
          }
        />
      </div>
    );
  return <>{children}</>;
}
