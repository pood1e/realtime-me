import type { ButtonHTMLAttributes, KeyboardEvent, PropsWithChildren, ReactNode } from "react";
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import {
  AlertCircle,
  Check,
  ChevronRight,
  File,
  FileArchive,
  FileCode2,
  FileImage,
  FileText,
  Folder,
  LoaderCircle,
  Music2,
  Video,
  X,
} from "lucide-react";
import type { DriveItem } from "@cloud-drive/contracts";

import { fileExtension, formatBytes, formatDate } from "./format";
import {
  driveItemContentType,
  driveItemIsDirectory,
  driveItemName,
  driveItemSize,
  driveItemUid,
  driveItemUpdatedAt,
} from "./message";

export function cn(...values: Array<string | false | null | undefined>): string {
  return values.filter(Boolean).join(" ");
}

type ButtonVariant = "default" | "outline" | "ghost" | "destructive";
type ButtonSize = "default" | "sm" | "icon";

const buttonVariants: Record<ButtonVariant, string> = {
  default: "bg-sky-500 text-slate-950 hover:bg-sky-300 focus-visible:ring-sky-300",
  outline: "border border-white/12 bg-white/[0.025] text-slate-100 hover:border-white/25 hover:bg-white/[0.075] focus-visible:ring-white/50",
  ghost: "text-slate-300 hover:bg-white/[0.075] hover:text-white focus-visible:ring-white/50",
  destructive: "bg-rose-500/90 text-white hover:bg-rose-400 focus-visible:ring-rose-300",
};

const buttonSizes: Record<ButtonSize, string> = {
  default: "h-10 px-4 text-sm",
  sm: "h-8 px-3 text-xs",
  icon: "size-9",
};

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
  size?: ButtonSize;
};

export function Button({ className, variant = "default", size = "default", type = "button", ...props }: ButtonProps) {
  return (
    <button
      type={type}
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-colors outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-slate-950 disabled:pointer-events-none disabled:opacity-45",
        buttonVariants[variant],
        buttonSizes[size],
        className,
      )}
      {...props}
    />
  );
}

export function IconButton({ label, children, className, ...props }: Omit<ButtonProps, "size" | "children"> & { label: string; children: ReactNode }) {
  return (
    <Button aria-label={label} title={label} size="icon" variant="ghost" className={className} {...props}>
      {children}
    </Button>
  );
}

export type Breadcrumb = Readonly<{
  id: string;
  label: string;
}>;

export type DialogSize = "compact" | "standard" | "preview";

const dialogSizes: Record<DialogSize, string> = {
  compact: "max-w-md",
  standard: "max-w-2xl",
  preview: "h-[calc(100dvh-1rem)] max-w-none sm:h-[min(90dvh,64rem)] sm:w-[calc(100vw-3rem)] xl:w-[calc(100vw-5rem)]",
};

export function Breadcrumbs({ items, onNavigate }: { items: readonly Breadcrumb[]; onNavigate?: (id: string) => void }) {
  return (
    <nav aria-label="Breadcrumb" className="flex min-w-0 items-center gap-1 overflow-x-auto text-sm text-slate-400">
      {items.map((item, index) => (
        <span className="flex shrink-0 items-center gap-1" key={item.id}>
          {index > 0 ? <ChevronRight className="size-3.5 text-slate-600" aria-hidden="true" /> : null}
          <button
            type="button"
            className={cn("rounded px-1 py-0.5 transition-colors hover:text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sky-300", index === items.length - 1 && "text-slate-100")}
            onClick={() => onNavigate?.(item.id)}
          >
            {item.label}
          </button>
        </span>
      ))}
    </nav>
  );
}

export function Dialog({
  open,
  title,
  description,
  size = "compact",
  onClose,
  children,
}: PropsWithChildren<{
  open: boolean;
  title: string;
  description?: string;
  size?: DialogSize;
  onClose: () => void;
}>) {
  useEffect(() => {
    if (!open) {
      return undefined;
    }
    const onKeyDown = (event: globalThis.KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [onClose, open]);

  if (!open) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 grid place-items-center p-2 sm:px-6 sm:py-8" role="presentation">
      <button aria-label="Close dialog" className="absolute inset-0 bg-slate-950/75 backdrop-blur-sm" onClick={onClose} />
      <section
        role="dialog"
        aria-modal="true"
        aria-labelledby="dialog-title"
        className={cn(
          "relative z-10 max-h-full w-full overflow-auto rounded-2xl border border-white/10 bg-slate-900 p-5 shadow-2xl shadow-black/40",
          dialogSizes[size],
        )}
      >
        <div className="mb-5 flex items-start justify-between gap-4">
          <div>
            <h2 id="dialog-title" className="text-base font-semibold text-white">
              {title}
            </h2>
            {description ? <p className="mt-1 text-sm leading-5 text-slate-400">{description}</p> : null}
          </div>
          <IconButton label="Close dialog" onClick={onClose} className="-mr-2 -mt-2 shrink-0">
            <X className="size-4" />
          </IconButton>
        </div>
        {children}
      </section>
    </div>
  );
}

export function LoadingIndicator({ label = "Loading" }: { label?: string }) {
  return (
    <div className="flex items-center justify-center gap-2 py-14 text-sm text-slate-400">
      <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />
      <span>{label}</span>
    </div>
  );
}

export function EmptyState({ icon, title, detail, action }: { icon?: ReactNode; title: string; detail?: string; action?: ReactNode }) {
  return (
    <div className="flex min-h-64 flex-col items-center justify-center px-6 py-12 text-center">
      <div className="mb-4 grid size-12 place-items-center rounded-2xl bg-white/[0.055] text-sky-300">{icon ?? <Folder className="size-6" />}</div>
      <h2 className="text-base font-medium text-white">{title}</h2>
      {detail ? <p className="mt-1 max-w-sm text-sm leading-6 text-slate-400">{detail}</p> : null}
      {action ? <div className="mt-5">{action}</div> : null}
    </div>
  );
}

export function InlineError({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <div className="m-4 flex items-center justify-between gap-4 rounded-xl border border-rose-400/20 bg-rose-400/10 px-4 py-3 text-sm text-rose-100">
      <span className="flex min-w-0 items-center gap-2"><AlertCircle className="size-4 shrink-0" />{message}</span>
      {onRetry ? <Button variant="ghost" size="sm" onClick={onRetry}>Retry</Button> : null}
    </div>
  );
}

export function FileGlyph({ item, className }: { item: DriveItem; className?: string }) {
  const contentType = driveItemContentType(item);
  const extension = fileExtension(driveItemName(item));
  const props = { className: cn("size-5", className), "aria-hidden": true };

  if (driveItemIsDirectory(item)) {
    return <Folder {...props} className={cn("size-5 fill-amber-300/15 text-amber-300", className)} />;
  }
  if (contentType.startsWith("image/") || ["avif", "gif", "jpg", "jpeg", "png", "svg", "webp"].includes(extension)) {
    return <FileImage {...props} className={cn("size-5 text-fuchsia-300", className)} />;
  }
  if (contentType.startsWith("audio/")) {
    return <Music2 {...props} className={cn("size-5 text-violet-300", className)} />;
  }
  if (contentType.startsWith("video/")) {
    return <Video {...props} className={cn("size-5 text-rose-300", className)} />;
  }
  if (["zip", "gz", "rar", "7z", "tar"].includes(extension)) {
    return <FileArchive {...props} className={cn("size-5 text-amber-200", className)} />;
  }
  if (contentType.startsWith("text/") || ["css", "go", "html", "js", "json", "md", "py", "rs", "sql", "ts", "tsx", "xml", "yaml", "yml"].includes(extension)) {
    return <FileCode2 {...props} className={cn("size-5 text-sky-300", className)} />;
  }
  if (extension === "pdf" || ["doc", "docx", "rtf", "txt"].includes(extension)) {
    return <FileText {...props} className={cn("size-5 text-emerald-300", className)} />;
  }
  return <File {...props} className={cn("size-5 text-slate-400", className)} />;
}

export function DriveItemList({
  items,
  empty,
  selectedUid,
  onOpen,
  actions,
}: {
  items: readonly DriveItem[];
  empty: ReactNode;
  selectedUid?: string;
  onOpen: (item: DriveItem) => void;
  actions?: (item: DriveItem) => ReactNode;
}) {
  if (!items.length) {
    return <>{empty}</>;
  }

  const onKeyOpen = (event: KeyboardEvent<HTMLDivElement>, item: DriveItem) => {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      onOpen(item);
    }
  };

  return (
    <div className="relative isolate rounded-xl border border-white/[0.08] bg-slate-900/55">
      <div className="sticky top-0 z-10 hidden grid-cols-[minmax(0,1fr)_7rem_10rem_3rem] gap-4 rounded-t-xl border-b border-white/[0.08] bg-slate-900/95 px-4 py-2.5 text-[11px] font-medium uppercase tracking-[0.14em] text-slate-500 backdrop-blur sm:grid xl:grid-cols-[minmax(20rem,1fr)_9rem_12rem_3rem]">
        <span>Name</span><span>Size</span><span>Modified</span><span aria-label="Actions" />
      </div>
      <div className="divide-y divide-white/[0.06]">
        {items.map((item) => {
          const uid = driveItemUid(item);
          const isDirectory = driveItemIsDirectory(item);
          return (
            <div
              role="button"
              tabIndex={0}
              key={uid}
              onClick={() => onOpen(item)}
              onKeyDown={(event) => onKeyOpen(event, item)}
              className={cn(
                "group grid cursor-pointer grid-cols-[minmax(0,1fr)_auto] items-center gap-3 px-4 py-3 outline-none transition-colors first:rounded-t-xl last:rounded-b-xl hover:bg-white/[0.045] focus-visible:bg-white/[0.06] sm:grid-cols-[minmax(0,1fr)_7rem_10rem_3rem] sm:gap-4 sm:first:rounded-t-none xl:grid-cols-[minmax(20rem,1fr)_9rem_12rem_3rem]",
                selectedUid === uid && "bg-sky-400/10",
              )}
            >
              <div className="flex min-w-0 items-center gap-3">
                <FileGlyph item={item} />
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium text-slate-100" title={driveItemName(item)}>{driveItemName(item)}</p>
                  <p className="mt-0.5 text-xs text-slate-500 sm:hidden">{isDirectory ? "Folder" : `${formatBytes(driveItemSize(item))} · ${formatDate(driveItemUpdatedAt(item))}`}</p>
                </div>
              </div>
              <span className="hidden text-sm text-slate-400 sm:block">{isDirectory ? "—" : formatBytes(driveItemSize(item))}</span>
              <span className="hidden text-sm text-slate-400 sm:block">{formatDate(driveItemUpdatedAt(item))}</span>
              <div className="flex justify-end" onClick={(event) => event.stopPropagation()}>{actions?.(item)}</div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

type ToastVariant = "default" | "error";
type Toast = Readonly<{ id: number; message: string; variant: ToastVariant }>;
type ToastContextValue = Readonly<{ showToast: (message: string, variant?: ToastVariant) => void }>;

const ToastContext = createContext<ToastContextValue | undefined>(undefined);

export function ToastProvider({ children }: PropsWithChildren) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const showToast = useCallback((message: string, variant: ToastVariant = "default") => {
    const id = Date.now() + Math.round(Math.random() * 10_000);
    setToasts((current) => [...current, { id, message, variant }]);
    window.setTimeout(() => setToasts((current) => current.filter((toast) => toast.id !== id)), 4_000);
  }, []);
  const context = useMemo(() => ({ showToast }), [showToast]);

  return (
    <ToastContext.Provider value={context}>
      {children}
      <div aria-live="polite" className="pointer-events-none fixed bottom-5 right-5 z-[60] flex w-[min(24rem,calc(100vw-2.5rem))] flex-col gap-2">
        {toasts.map((toast) => (
          <div key={toast.id} className={cn("pointer-events-auto flex items-center gap-2 rounded-xl border px-4 py-3 text-sm shadow-xl", toast.variant === "error" ? "border-rose-300/20 bg-rose-950 text-rose-100" : "border-sky-300/20 bg-slate-800 text-slate-100")}>
            {toast.variant === "error" ? <AlertCircle className="size-4 shrink-0" /> : <Check className="size-4 shrink-0 text-sky-300" />}
            <span>{toast.message}</span>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast(): ToastContextValue {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error("useToast must be used inside ToastProvider.");
  }
  return context;
}
