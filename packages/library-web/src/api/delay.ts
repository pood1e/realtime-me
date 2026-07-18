export function abortableDelay(milliseconds: number, signal?: AbortSignal): Promise<void> {
  if (signal?.aborted) return Promise.reject(abortReason(signal));

  return new Promise((resolve, reject) => {
    const complete = () => {
      signal?.removeEventListener("abort", abort);
      resolve();
    };
    const abort = () => {
      window.clearTimeout(timeout);
      reject(signal ? abortReason(signal) : new DOMException("Aborted", "AbortError"));
    };
    const timeout = window.setTimeout(complete, milliseconds);
    signal?.addEventListener("abort", abort, { once: true });
  });
}

function abortReason(signal: AbortSignal): unknown {
  return signal.reason ?? new DOMException("Aborted", "AbortError");
}
