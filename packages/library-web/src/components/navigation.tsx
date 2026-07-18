import { Button } from "@realtime-me/web-ui/button";
import { cn } from "@realtime-me/web-ui/cn";
import { ChevronRight, LayoutGrid, List as ListIcon } from "lucide-react";
import type { ReactNode } from "react";
import { useCallback, useState } from "react";

export type Breadcrumb = Readonly<{ id: string; label: string }>;
export type DriveViewMode = "list" | "grid";

function storedDriveViewMode(storageKey: string): DriveViewMode {
  try {
    return window.localStorage.getItem(storageKey) === "grid" ? "grid" : "list";
  } catch {
    return "list";
  }
}

export function useDriveViewMode(storageKey: string) {
  const [mode, setMode] = useState<DriveViewMode>(() => storedDriveViewMode(storageKey));
  const selectMode = useCallback(
    (nextMode: DriveViewMode) => {
      setMode(nextMode);
      try {
        window.localStorage.setItem(storageKey, nextMode);
      } catch {
        // The selected mode remains active for this session.
      }
    },
    [storageKey],
  );
  return [mode, selectMode] as const;
}

export function IconButton({
  label,
  children,
  ...props
}: Omit<React.ComponentProps<typeof Button>, "size" | "children"> & {
  label: string;
  children: ReactNode;
}) {
  return (
    <Button aria-label={label} title={label} size="icon" variant="ghost" {...props}>
      {children}
    </Button>
  );
}

export function Breadcrumbs({
  items,
  onNavigate,
}: {
  items: readonly Breadcrumb[];
  onNavigate?: (id: string) => void;
}) {
  return (
    <nav
      aria-label="Breadcrumb"
      className="flex min-w-0 items-center gap-1 overflow-x-auto text-sm text-muted-foreground"
    >
      {items.map((item, index) => (
        <span className="flex shrink-0 items-center gap-1" key={item.id}>
          {index > 0 ? <ChevronRight className="size-3.5 opacity-45" aria-hidden="true" /> : null}
          <button
            type="button"
            className={cn(
              "rounded px-1 py-0.5 transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
              index === items.length - 1 && "text-foreground",
            )}
            onClick={() => onNavigate?.(item.id)}
          >
            {item.label}
          </button>
        </span>
      ))}
    </nav>
  );
}

export function DriveViewModeToggle({
  mode,
  onChange,
}: {
  mode: DriveViewMode;
  onChange: (mode: DriveViewMode) => void;
}) {
  return (
    <fieldset className="m-0 flex min-w-0 shrink-0 rounded-lg border bg-muted/30 p-1">
      <legend className="sr-only">文件视图</legend>
      <IconButton
        label="列表视图"
        aria-pressed={mode === "list"}
        onClick={() => onChange("list")}
        className={cn(mode === "list" && "bg-accent text-foreground")}
      >
        <ListIcon className="size-4" aria-hidden="true" />
      </IconButton>
      <IconButton
        label="网格视图"
        aria-pressed={mode === "grid"}
        onClick={() => onChange("grid")}
        className={cn(mode === "grid" && "bg-accent text-foreground")}
      >
        <LayoutGrid className="size-4" aria-hidden="true" />
      </IconButton>
    </fieldset>
  );
}
