import { Code, ConnectError } from '@connectrpc/connect';
import { useEffect, useRef, useState } from 'react';

export type Poll<T> = {
  data: T | null;
  error: unknown;
  refresh: () => void;
};

// usePolling drives a request on an interval. Each run aborts the previous
// in-flight request before issuing the next, and results from aborted or
// unmounted requests are ignored. An intervalMs of 0 loads once without polling.
export function usePolling<T>(
  fetcher: (signal: AbortSignal) => Promise<T>,
  options: { intervalMs: number; enabled?: boolean },
): Poll<T> {
  const { intervalMs, enabled = true } = options;
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<unknown>(null);
  const refreshRef = useRef<() => void>(() => {});

  useEffect(() => {
    setData(null);
    setError(null);
    if (!enabled) {
      refreshRef.current = () => {};
      return;
    }

    let active = true;
    let controller: AbortController | null = null;

    const run = () => {
      controller?.abort();
      const current = new AbortController();
      controller = current;
      fetcher(current.signal)
        .then((result) => {
          if (active && !current.signal.aborted) {
            setData(result);
            setError(null);
          }
        })
        .catch((cause: unknown) => {
          if (active && !current.signal.aborted && !isAbort(cause)) setError(cause);
        });
    };

    refreshRef.current = run;
    run();
    const interval = intervalMs > 0 ? window.setInterval(run, intervalMs) : 0;
    return () => {
      active = false;
      controller?.abort();
      if (interval) window.clearInterval(interval);
    };
  }, [fetcher, intervalMs, enabled]);

  return { data, error, refresh: () => refreshRef.current() };
}

function isAbort(cause: unknown): boolean {
  if (cause instanceof ConnectError) return cause.code === Code.Canceled;
  return cause instanceof DOMException && cause.name === 'AbortError';
}
