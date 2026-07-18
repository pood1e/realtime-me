import { useEffect, useRef } from "react";

import { Button } from "./ui/button";
import { LoadingIndicator } from "./feedback";

export function InfiniteScrollSentinel({
  hasMore,
  loading,
  failed,
  loadingLabel,
  completeLabel,
  onLoadMore,
}: {
  hasMore: boolean;
  loading: boolean;
  failed: boolean;
  loadingLabel: string;
  completeLabel: string;
  onLoadMore: () => void;
}) {
  const sentinel = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const target = sentinel.current;
    if (
      !target ||
      !hasMore ||
      loading ||
      failed ||
      !("IntersectionObserver" in window)
    )
      return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry?.isIntersecting) onLoadMore();
      },
      { rootMargin: "480px 0px" },
    );
    observer.observe(target);
    return () => observer.disconnect();
  }, [failed, hasMore, loading, onLoadMore]);

  if (!hasMore)
    return (
      <p className="sr-only" aria-live="polite">
        {completeLabel}
      </p>
    );
  return (
    <div
      ref={sentinel}
      className="flex min-h-24 items-center justify-center"
      aria-live="polite"
    >
      {loading ? (
        <LoadingIndicator label={loadingLabel} />
      ) : (
        <Button variant="outline" onClick={onLoadMore}>
          {failed ? "重试加载" : "加载更多"}
        </Button>
      )}
    </div>
  );
}
