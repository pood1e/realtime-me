import { Button } from "@realtime-me/web-ui/button";
import { RefreshCw, ShieldAlert } from "lucide-react";
import type { PropsWithChildren } from "react";
import { useEffect, useMemo, useState } from "react";
import {
  authenticationUrl,
  hasRecentSessionValidation,
  isUnauthenticatedError,
  markSessionValidated,
  SessionClient,
} from "../api";
import { EmptyState, LoadingIndicator } from "./feedback";

type State = "checking" | "ready" | "failed";
const feedbackDelayMs = 600;
const validationTimeoutMs = 10_000;

export function AuthGuard({
  apiBase,
  authOrigin,
  children,
}: PropsWithChildren<{ apiBase: string; authOrigin: string }>) {
  const client = useMemo(() => new SessionClient(apiBase), [apiBase]);
  const [state, setState] = useState<State>(() =>
    hasRecentSessionValidation(apiBase) ? "ready" : "checking",
  );
  const [showProgress, setShowProgress] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    const controller = new AbortController();
    let disposed = false;
    let timedOut = false;
    const backgroundValidation = hasRecentSessionValidation(apiBase);
    if (!backgroundValidation) setState("checking");
    setShowProgress(false);
    setMessage("");
    const feedbackTimer = window.setTimeout(() => {
      if (!disposed) setShowProgress(true);
    }, feedbackDelayMs);
    const timeoutTimer = window.setTimeout(() => {
      if (disposed) return;
      timedOut = true;
      controller.abort();
      if (backgroundValidation) return;
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
        markSessionValidated(apiBase);
        setState("ready");
      })
      .catch((error: unknown) => {
        if (disposed || timedOut) return;
        clearTimers();
        if (isUnauthenticatedError(error)) {
          window.location.replace(authenticationUrl(authOrigin));
          return;
        }
        if (backgroundValidation) return;
        setMessage(error instanceof Error ? error.message : "暂时无法连接服务。");
        setState("failed");
      });
    return () => {
      disposed = true;
      clearTimers();
      controller.abort();
    };
  }, [apiBase, authOrigin, client]);

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
