import type { Timestamp } from '@bufbuild/protobuf/wkt';
import { AlertTriangle, CheckCircle2, Clock, LoaderCircle, RefreshCw } from 'lucide-react';
import type { ReactElement, ReactNode } from 'react';
import blueberryLogoUrl from '@/assets/blueberry.svg';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { ThemeToggle } from '@/components/theme';
import { formatTime } from '@/lib/format';

export function PageFrame({ maxWidth = 'max-w-6xl', children }: { maxWidth?: string; children: ReactNode }) {
  return (
    <TooltipProvider>
      <main className="min-h-screen bg-[radial-gradient(46rem_30rem_at_50%_-4rem,color-mix(in_oklab,var(--primary)_13%,transparent),transparent)]">
        <div className={`mx-auto flex min-h-screen w-full ${maxWidth} flex-col gap-8 px-5 py-9`}>{children}</div>
      </main>
    </TooltipProvider>
  );
}

export function SiteLogo() {
  return (
    <a href="/" className="flex items-center gap-2.5" aria-label="Home">
      <img src={blueberryLogoUrl} alt="" className="size-11 rounded-2xl drop-shadow-sm" width={44} height={44} />
      <span className="font-heading text-xl font-semibold tracking-tight">pood1e</span>
    </a>
  );
}

export function NavLinks() {
  const onAbout = window.location.pathname.startsWith('/about');
  return (
    <nav className="flex items-center gap-1 text-sm">
      <a href="/" className={navLinkClass(!onAbout)}>Status</a>
      <a href="/about" className={navLinkClass(onAbout)}>About</a>
    </nav>
  );
}

function navLinkClass(active: boolean): string {
  return `rounded-md px-2.5 py-1 font-medium transition-colors ${active ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'}`;
}

export function HeaderActions({ failed, refresh }: { failed: boolean; refresh: () => void }) {
  return (
    <div className="flex items-center gap-2">
      <Badge variant={failed ? 'destructive' : 'default'} aria-label={failed ? 'API offline' : 'Online'} title={failed ? 'API offline' : 'Online'}>
        {failed ? <AlertTriangle /> : <CheckCircle2 />}
      </Badge>
      <ThemeToggle />
      <Tooltip>
        <TooltipTrigger asChild>
          <Button variant="secondary" size="icon" aria-label="Refresh" title="Refresh" onClick={refresh}>
            <RefreshCw />
          </Button>
        </TooltipTrigger>
        <TooltipContent>Refresh</TooltipContent>
      </Tooltip>
    </div>
  );
}

export function StatusSection({ title, icon, columns = 'md:grid-cols-2 xl:grid-cols-4', children }: {
  title: string;
  icon: ReactElement;
  columns?: string;
  children: ReactNode;
}) {
  return (
    <section className="grid gap-3">
      <div className="flex items-end justify-between gap-3">
        <h2 className="flex items-center gap-2 text-xl font-semibold tracking-tight">{icon}{title}</h2>
      </div>
      <div className={`grid gap-4 ${columns}`}>{children}</div>
    </section>
  );
}

export function SummaryCard({ icon, title, value, detail }: { icon: ReactElement; title: string; value: string; detail: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-sm text-muted-foreground">{icon}{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-3xl font-semibold tracking-tight">{value}</div>
        <div className="mt-1 text-xs text-muted-foreground">{detail}</div>
      </CardContent>
    </Card>
  );
}

export function EmptyCard({ text }: { text: string }) {
  return (
    <Card>
      <CardContent>
        <CardDescription>{text}</CardDescription>
      </CardContent>
    </Card>
  );
}

export function LoadingCard() {
  return (
    <Card>
      <CardContent className="flex items-center gap-2 text-sm text-muted-foreground">
        <LoaderCircle className="size-4 animate-spin" />Loading
      </CardContent>
    </Card>
  );
}

export function InlineTime({ value }: { value?: Timestamp }) {
  return <span className="text-xs text-muted-foreground">{formatTime(value)}</span>;
}

export function PageFooter({ updatedAt }: { updatedAt?: Timestamp }) {
  return (
    <footer className="flex items-center gap-2 text-xs text-muted-foreground">
      <Clock className="size-3.5" />
      <span>{updatedAt ? `Updated ${formatTime(updatedAt)}` : 'Waiting for first status'}</span>
    </footer>
  );
}
