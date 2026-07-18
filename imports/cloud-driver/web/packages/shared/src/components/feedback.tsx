import type { PropsWithChildren, ReactNode } from "react";
import { useCallback, useMemo } from "react";
import { AlertCircle, Folder, LoaderCircle } from "lucide-react";
import { Toaster, toast } from "sonner";

import { Button } from "./ui/button";

export function LoadingIndicator({ label = "加载中" }: { label?: string }) {
  return (
    <div className="flex items-center justify-center gap-2 py-14 text-sm text-muted-foreground">
      <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />
      <span>{label}</span>
    </div>
  );
}

export function EmptyState({
  icon,
  title,
  detail,
  action,
}: {
  icon?: ReactNode;
  title: string;
  detail?: string;
  action?: ReactNode;
}) {
  return (
    <div className="flex min-h-64 flex-col items-center justify-center px-6 py-12 text-center">
      <div className="mb-4 grid size-12 place-items-center rounded-2xl bg-muted text-primary">
        {icon ?? <Folder className="size-6" />}
      </div>
      <h2 className="text-base font-medium text-foreground">{title}</h2>
      {detail ? (
        <p className="mt-1 max-w-sm text-sm leading-6 text-muted-foreground">
          {detail}
        </p>
      ) : null}
      {action ? <div className="mt-5">{action}</div> : null}
    </div>
  );
}

export function InlineError({
  message,
  onRetry,
}: {
  message: string;
  onRetry?: () => void;
}) {
  return (
    <div className="m-4 flex items-center justify-between gap-4 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-foreground">
      <span className="flex min-w-0 items-center gap-2">
        <AlertCircle className="size-4 shrink-0 text-destructive" />
        {message}
      </span>
      {onRetry ? (
        <Button variant="ghost" size="sm" onClick={onRetry}>
          重试
        </Button>
      ) : null}
    </div>
  );
}

type ToastVariant = "default" | "error";
type ToastContextValue = Readonly<{
  showToast: (message: string, variant?: ToastVariant) => void;
}>;

export function ToastProvider({ children }: PropsWithChildren) {
  return (
    <>
      {children}
      <Toaster richColors position="bottom-right" />
    </>
  );
}

export function useToast(): ToastContextValue {
  const showToast = useCallback(
    (message: string, variant: ToastVariant = "default") => {
      if (variant === "error") toast.error(message);
      else toast.success(message);
    },
    [],
  );
  return useMemo(() => ({ showToast }), [showToast]);
}
