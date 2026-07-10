import type { FormEvent } from "react";
import { useEffect, useMemo, useState } from "react";
import { KeyRound, LoaderCircle, LockKeyhole, ShieldCheck } from "lucide-react";
import {
  Button,
  Input,
  SessionClient,
  isUnauthenticatedError,
} from "@cloud-drive/shared";

import { API_BASE, DEFAULT_RETURN_URL, PRIVATE_APP_ORIGINS } from "./config";

function requestedReturnUrl(): string {
  const candidate =
    new URLSearchParams(window.location.search).get("return_to") ||
    DEFAULT_RETURN_URL;
  try {
    const parsed = new URL(candidate);
    return PRIVATE_APP_ORIGINS.has(parsed.origin)
      ? parsed.toString()
      : DEFAULT_RETURN_URL;
  } catch {
    return DEFAULT_RETURN_URL;
  }
}

export function App() {
  const client = useMemo(() => new SessionClient(API_BASE), []);
  const [password, setPassword] = useState("");
  const [checking, setChecking] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const returnUrl = useMemo(requestedReturnUrl, []);

  useEffect(() => {
    const controller = new AbortController();
    void client
      .getSession(controller.signal)
      .then(() => window.location.replace(returnUrl))
      .catch((sessionError: unknown) => {
        if (!controller.signal.aborted && !isUnauthenticatedError(sessionError))
          setError(
            sessionError instanceof Error
              ? sessionError.message
              : "服务暂时不可用。",
          );
        setChecking(false);
      });
    return () => controller.abort();
  }, [client, returnUrl]);

  const login = async (event: FormEvent) => {
    event.preventDefault();
    if (!password || submitting) return;
    setSubmitting(true);
    setError("");
    try {
      const response = await client.login(password, returnUrl);
      window.location.replace(response.returnUrl || returnUrl);
    } catch (loginError) {
      setError(
        loginError instanceof Error ? loginError.message : "密码不正确。",
      );
      setSubmitting(false);
    }
  };

  return (
    <main className="relative grid min-h-dvh overflow-hidden bg-background px-5 py-10 text-foreground sm:place-items-center">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_20%_15%,oklch(0.75_0.14_231.6/0.15),transparent_32%),radial-gradient(circle_at_80%_85%,oklch(0.63_0.16_290/0.1),transparent_30%)]" />
      <section className="relative w-full max-w-sm rounded-3xl border bg-card/75 p-7 shadow-2xl shadow-black/25 backdrop-blur-xl sm:p-9">
        <div className="mb-8 flex size-12 items-center justify-center rounded-2xl bg-primary/12 text-primary">
          <LockKeyhole className="size-6" />
        </div>
        <p className="text-xs font-semibold tracking-[0.2em] text-primary uppercase">
          Local Library
        </p>
        <h1 className="mt-3 text-2xl font-semibold tracking-tight">欢迎回来</h1>
        <p className="mt-2 text-sm leading-6 text-muted-foreground">
          一个密码进入你的书架、音乐盒、图床与云盘。
        </p>
        {checking ? (
          <div className="mt-8 flex items-center gap-2 text-sm text-muted-foreground">
            <LoaderCircle className="size-4 animate-spin" />
            正在检查会话
          </div>
        ) : (
          <form
            className="mt-8 space-y-4"
            onSubmit={(event) => void login(event)}
          >
            <label className="block text-sm font-medium" htmlFor="password">
              访问密码
            </label>
            <div className="relative">
              <KeyRound className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                id="password"
                autoFocus
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                className="h-11 pl-10"
              />
            </div>
            {error ? (
              <p role="alert" className="text-sm text-destructive">
                {error}
              </p>
            ) : null}
            <Button
              type="submit"
              className="h-11 w-full"
              disabled={!password || submitting}
            >
              {submitting ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : (
                <ShieldCheck className="size-4" />
              )}
              {submitting ? "正在登录" : "进入"}
            </Button>
          </form>
        )}
      </section>
    </main>
  );
}
