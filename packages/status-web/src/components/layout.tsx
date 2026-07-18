import type { Timestamp } from "@bufbuild/protobuf/wkt";
import { formatTime } from "@realtime-me/status-web/lib/format";
import { ThemeToggle } from "@realtime-me/web-shell/theme";
import { Badge } from "@realtime-me/web-ui/badge";
import { Button } from "@realtime-me/web-ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@realtime-me/web-ui/card";
import { Skeleton } from "@realtime-me/web-ui/skeleton";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@realtime-me/web-ui/tooltip";
import {
  AlertTriangle,
  CheckCircle2,
  Clock,
  CloudOff,
  LoaderCircle,
  RefreshCw,
} from "lucide-react";
import type { ReactElement, ReactNode } from "react";
import { Link } from "react-router-dom";

export function PageFrame({
  maxWidth = "max-w-6xl",
  children,
}: {
  maxWidth?: string;
  children: ReactNode;
}) {
  return (
    <TooltipProvider>
      <main className="min-h-screen bg-[radial-gradient(46rem_30rem_at_50%_-4rem,color-mix(in_oklab,var(--primary)_13%,transparent),transparent)]">
        <div className={`mx-auto flex min-h-screen w-full ${maxWidth} flex-col gap-8 px-5 py-9`}>
          {children}
        </div>
      </main>
    </TooltipProvider>
  );
}

export function SiteLogo() {
  return (
    <Link to="/" className="flex items-center gap-2.5" aria-label="Home">
      <img
        src="/scallion.png"
        alt=""
        className="size-11 rounded-2xl drop-shadow-sm"
        width={44}
        height={44}
      />
      <span className="font-heading text-xl font-semibold tracking-tight">pood1e</span>
    </Link>
  );
}

export function HeaderActions({ failed, refresh }: { failed: boolean; refresh: () => void }) {
  return (
    <div className="flex items-center gap-2">
      <Badge
        variant={failed ? "destructive" : "default"}
        aria-label={failed ? "API offline" : "Online"}
        title={failed ? "API offline" : "Online"}
      >
        {failed ? <AlertTriangle /> : <CheckCircle2 />}
      </Badge>
      <ThemeToggle />
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="secondary"
            size="icon"
            aria-label="Refresh"
            title="Refresh"
            onClick={refresh}
          >
            <RefreshCw />
          </Button>
        </TooltipTrigger>
        <TooltipContent>Refresh</TooltipContent>
      </Tooltip>
    </div>
  );
}

export function StatusSection({
  title,
  icon,
  columns = "sm:grid-cols-2 lg:grid-cols-3",
  children,
}: {
  title: string;
  icon: ReactElement;
  columns?: string;
  children: ReactNode;
}) {
  return (
    <section className="grid gap-3.5">
      <h2 className="flex items-center gap-2 text-xl font-semibold tracking-tight text-muted-foreground">
        <span className="text-primary">{icon}</span>
        {title}
      </h2>
      <div className={`grid items-start gap-4 ${columns}`}>{children}</div>
    </section>
  );
}

export function SummaryCard({
  icon,
  title,
  value,
  detail,
}: {
  icon: ReactElement;
  title: string;
  value: string;
  detail: string;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-sm text-muted-foreground">
          {icon}
          {title}
        </CardTitle>
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
        <LoaderCircle className="size-4 animate-spin" />
        Loading
      </CardContent>
    </Card>
  );
}

// SkeletonCard holds a card's shape while the first poll is in flight, so the
// page does not paint every empty state and then re-lay-out around real data.
export function SkeletonCard() {
  return (
    <Card aria-hidden>
      <CardHeader>
        <CardTitle className="w-full">
          <Skeleton className="h-4 w-28" />
        </CardTitle>
      </CardHeader>
      <CardContent className="grid gap-3">
        <Skeleton className="h-9 w-full" />
        <Skeleton className="h-3 w-2/3" />
      </CardContent>
    </Card>
  );
}

// ErrorCard says the backend is unreachable. Rendering an empty state instead
// tells the visitor there is nothing here, which is a different and false claim.
export function ErrorCard({ text, retry }: { text: string; retry?: () => void }) {
  return (
    <Card>
      <CardContent className="flex flex-wrap items-center gap-3">
        <CloudOff className="size-4 shrink-0 text-destructive" />
        <CardDescription className="grow">{text}</CardDescription>
        {retry && (
          <Button variant="secondary" size="sm" onClick={retry}>
            <RefreshCw />
            Retry
          </Button>
        )}
      </CardContent>
    </Card>
  );
}

export function InlineTime({ value }: { value?: Timestamp | undefined }) {
  return <span className="text-xs text-muted-foreground">{formatTime(value)}</span>;
}

export function PageFooter({ updatedAt }: { updatedAt?: Timestamp | undefined }) {
  return (
    <footer className="flex items-center gap-2 text-xs text-muted-foreground">
      <Clock className="size-3.5" />
      <span>{updatedAt ? `Updated ${formatTime(updatedAt)}` : "Waiting for first status"}</span>
    </footer>
  );
}
