import type { PublicStatus } from "@realtime-me/status-contracts";
import { ContactLinks } from "@realtime-me/status-web/components/ContactLinks";
import { Presence } from "@realtime-me/status-web/components/Presence";
import { usePolling } from "@realtime-me/status-web/hooks/usePolling";
import type { StatusApi } from "@realtime-me/status-web/lib/transport";
import { POLL_INTERVAL_MS } from "@realtime-me/status-web/lib/transport";
import { ThemeToggle } from "@realtime-me/web-shell/theme";
import { Badge } from "@realtime-me/web-ui/badge";
import { Button } from "@realtime-me/web-ui/button";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@realtime-me/web-ui/sheet";
import { Skeleton } from "@realtime-me/web-ui/skeleton";
import { TooltipProvider } from "@realtime-me/web-ui/tooltip";
import { CloudOff, Menu } from "lucide-react";
import type { ReactNode } from "react";
import { useCallback } from "react";
import { Link, NavLink, Outlet } from "react-router-dom";

export type ShellContext = {
  status: PublicStatus | null;
  statusFailed: boolean;
  api: StatusApi;
};

export function AppShell({ api, consoleUrl }: { api: StatusApi; consoleUrl?: string | undefined }) {
  const fetchProfile = useCallback(
    async (signal: AbortSignal) => (await api.profile.getProfile({}, { signal })).profile ?? null,
    [api],
  );
  const { data: profile } = usePolling(fetchProfile, { intervalMs: 0 });
  const fetchStatus = useCallback(
    async (signal: AbortSignal) =>
      (await api.status.getPublicStatus({}, { signal })).status ?? null,
    [api],
  );
  const { data: status, error: statusError } = usePolling(fetchStatus, {
    intervalMs: POLL_INTERVAL_MS,
  });
  const statusFailed = statusError !== null;

  return (
    <TooltipProvider>
      <div className="flex min-h-screen flex-col bg-[radial-gradient(46rem_30rem_at_50%_-6rem,color-mix(in_oklab,var(--primary)_12%,transparent),transparent)]">
        <header className="sticky top-0 z-30 border-b border-border/60 bg-background/70 backdrop-blur-md">
          <div className="mx-auto flex w-full max-w-6xl items-center justify-between gap-3 px-5 py-2.5">
            <div className="flex items-center gap-4">
              {/* The name and avatar are the profile's to give. They were once
                  hardcoded here as a fallback, which is precisely why a profile that
                  had gone missing still looked like a working page. */}
              <Link to="/" className="flex items-center gap-2.5" aria-label="Home">
                {profile ? (
                  <>
                    <img
                      src={profile.avatarUrl}
                      alt=""
                      className="size-9 rounded-full border border-border object-cover"
                      width={36}
                      height={36}
                    />
                    <span className="hidden font-heading text-lg font-semibold tracking-tight sm:inline">
                      {profile.displayName}
                    </span>
                  </>
                ) : (
                  <>
                    <Skeleton className="size-9 rounded-full" />
                    <Skeleton className="hidden h-5 w-20 sm:block" />
                  </>
                )}
              </Link>
              {statusFailed ? <OfflineBadge /> : <Presence status={status} />}
            </div>
            <div className="hidden items-center gap-4 lg:flex">
              <NavTabs consoleUrl={consoleUrl} />
              <div className="flex items-center gap-0.5">
                <ContactLinks links={profile?.links} />
                <ThemeToggle />
              </div>
            </div>
            <MobileNav consoleUrl={consoleUrl} contactLinks={profile?.links} />
          </div>
        </header>

        <main className="mx-auto w-full max-w-6xl grow px-5 py-8 md:py-10">
          <Outlet context={{ status, statusFailed, api } satisfies ShellContext} />
        </main>

        <footer className="mx-auto flex w-full max-w-6xl items-center justify-between gap-3 px-5 py-6 text-xs text-muted-foreground">
          <span>© {new Date().getFullYear()} pood1e</span>
          <span>realtime-me · self-hosted realtime status</span>
        </footer>
      </div>
    </TooltipProvider>
  );
}

function MobileNav({
  consoleUrl,
  contactLinks,
}: {
  consoleUrl?: string | undefined;
  contactLinks: Parameters<typeof ContactLinks>[0]["links"];
}) {
  return (
    <div className="lg:hidden">
      <Sheet>
        <SheetTrigger asChild>
          <Button variant="ghost" size="icon" aria-label="Open navigation">
            <Menu />
          </Button>
        </SheetTrigger>
        <SheetContent side="right" className="w-72 p-5">
          <SheetHeader>
            <SheetTitle className="text-left">pood1e</SheetTitle>
          </SheetHeader>
          <nav className="mt-7 grid gap-1" aria-label="Site">
            <MobileTab to="/">Status</MobileTab>
            <MobileTab to="/projects">Projects</MobileTab>
            <MobileTab to="/wallpapers">Wallpapers</MobileTab>
            {consoleUrl ? (
              <SheetClose asChild>
                <a
                  href={consoleUrl}
                  className="rounded-lg px-3 py-2.5 text-sm font-medium text-muted-foreground hover:bg-accent hover:text-foreground"
                >
                  Console
                </a>
              </SheetClose>
            ) : null}
          </nav>
          <div className="mt-7 flex items-center gap-1 border-t pt-4">
            <ContactLinks links={contactLinks} />
            <ThemeToggle />
          </div>
        </SheetContent>
      </Sheet>
    </div>
  );
}

function MobileTab({ to, children }: { to: string; children: ReactNode }) {
  return (
    <SheetClose asChild>
      <NavLink
        to={to}
        end={to === "/"}
        className={({ isActive }) =>
          `rounded-lg px-3 py-2.5 text-sm font-medium ${
            isActive
              ? "bg-primary/12 text-primary"
              : "text-muted-foreground hover:bg-accent hover:text-foreground"
          }`
        }
      >
        {children}
      </NavLink>
    </SheetClose>
  );
}

// An outage must not read as an empty life: say the status is unreachable rather
// than render every card's "no data" state.
function OfflineBadge() {
  return (
    <Badge variant="destructive" className="gap-1.5" title="Cannot reach the status API">
      <CloudOff className="size-3.5" />
      Status offline
    </Badge>
  );
}

function NavTabs({ consoleUrl }: { consoleUrl?: string | undefined }) {
  return (
    <nav className="flex items-center gap-5 text-sm font-medium">
      <Tab to="/">Status</Tab>
      <Tab to="/projects">Projects</Tab>
      <Tab to="/wallpapers">Wallpapers</Tab>
      {consoleUrl ? (
        <a
          href={consoleUrl}
          className="py-1 text-muted-foreground transition-colors hover:text-foreground"
        >
          Console
        </a>
      ) : null}
    </nav>
  );
}

function Tab({ to, children }: { to: string; children: ReactNode }) {
  return (
    <NavLink
      to={to}
      end={to === "/"}
      className={({ isActive }) =>
        `relative py-1 transition-colors ${isActive ? "text-foreground" : "text-muted-foreground hover:text-foreground"}`
      }
    >
      {({ isActive }) => (
        <>
          {children}
          <span
            className={`absolute -bottom-0.5 left-0 h-0.5 rounded-full bg-primary transition-all duration-200 ${isActive ? "w-full opacity-100" : "w-0 opacity-0"}`}
          />
        </>
      )}
    </NavLink>
  );
}
