import { useCallback } from 'react';
import type { ReactNode } from 'react';
import { Link, NavLink, Outlet } from 'react-router-dom';
import type { ProfilePage } from '@/gen/realtime/me/v1/profile_pb';
import type { PublicStatus } from '@/gen/realtime/me/v1/status_pb';
import { TooltipProvider } from '@/components/ui/tooltip';
import { ContactLinks } from '@/components/ContactLinks';
import { Presence } from '@/components/Presence';
import { ThemeToggle } from '@/components/theme';
import { usePolling } from '@/hooks/usePolling';
import { POLL_INTERVAL_MS, profileClient, statusClient } from '@/lib/transport';

export type ShellContext = { page?: ProfilePage | null; status?: PublicStatus | null };

export function AppShell() {
  const fetchProfile = useCallback(async (signal: AbortSignal) => (await profileClient.getProfilePage({}, { signal })).page, []);
  const { data: page } = usePolling(fetchProfile, { intervalMs: 0 });
  const fetchStatus = useCallback(async (signal: AbortSignal) => (await statusClient.getPublicStatus({}, { signal })).status, []);
  const { data: status } = usePolling(fetchStatus, { intervalMs: POLL_INTERVAL_MS });
  const profile = page?.profile;

  return (
    <TooltipProvider>
      <div className="flex min-h-screen flex-col bg-[radial-gradient(46rem_30rem_at_50%_-6rem,color-mix(in_oklab,var(--primary)_12%,transparent),transparent)]">
        <header className="sticky top-0 z-30 border-b border-border/60 bg-background/70 backdrop-blur-md">
          <div className="mx-auto flex w-full max-w-6xl items-center justify-between gap-3 px-5 py-2.5">
            <div className="flex items-center gap-4">
              <Link to="/" className="flex items-center gap-2.5" aria-label="Home">
                <img
                  src={profile?.avatarUrl || '/scallion.png'}
                  alt=""
                  className="size-9 rounded-full border border-border object-cover"
                  width={36}
                  height={36}
                />
                <span className="font-heading text-lg font-semibold tracking-tight">{profile?.displayName || 'pood1e'}</span>
              </Link>
              <Presence status={status} />
            </div>
            <div className="flex items-center gap-2 sm:gap-4">
              <NavTabs />
              <div className="flex items-center gap-0.5">
                <ContactLinks links={profile?.links} />
                <ThemeToggle />
              </div>
            </div>
          </div>
        </header>

        <main className="mx-auto w-full max-w-6xl grow px-5 py-8 md:py-10">
          <Outlet context={{ page, status } satisfies ShellContext} />
        </main>

        <footer className="mx-auto flex w-full max-w-6xl items-center justify-between gap-3 px-5 py-6 text-xs text-muted-foreground">
          <span>© {new Date().getFullYear()} pood1e</span>
          <span>realtime-me · 自托管实时状态</span>
        </footer>
      </div>
    </TooltipProvider>
  );
}

function NavTabs() {
  return (
    <nav className="flex items-center gap-5 text-sm font-medium">
      <Tab to="/">Status</Tab>
      <Tab to="/about">About</Tab>
    </nav>
  );
}

function Tab({ to, children }: { to: string; children: ReactNode }) {
  return (
    <NavLink
      to={to}
      end={to === '/'}
      className={({ isActive }) => `relative py-1 transition-colors ${isActive ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
    >
      {({ isActive }) => (
        <>
          {children}
          <span
            className={`absolute -bottom-0.5 left-0 h-0.5 rounded-full bg-primary transition-all duration-200 ${isActive ? 'w-full opacity-100' : 'w-0 opacity-0'}`}
          />
        </>
      )}
    </NavLink>
  );
}
