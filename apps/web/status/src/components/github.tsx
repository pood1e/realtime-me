import { AlertTriangle, CheckCircle2, CircleOff, LoaderCircle } from 'lucide-react';
import type { ReactElement } from 'react';
import { siGithub } from 'simple-icons/icons';
import { GithubSyncState } from '@realtime-me/status-contracts';
import { Badge } from '@/components/ui/badge';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { BrandIcon } from '@/components/brand';

type BadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline';

export function githubStatusTitle(state: GithubSyncState | undefined): string {
  if (state === GithubSyncState.OK) return 'Connected';
  if (state === GithubSyncState.PENDING) return 'Connecting';
  if (state === GithubSyncState.ERROR) return 'Sync failed';
  return 'Disconnected';
}

export function githubBadgeVariant(state: GithubSyncState | undefined): BadgeVariant {
  if (state === GithubSyncState.ERROR) return 'destructive';
  if (state === GithubSyncState.OK) return 'default';
  if (state === GithubSyncState.PENDING) return 'secondary';
  return 'outline';
}

export function githubIcon(state: GithubSyncState | undefined): ReactElement {
  if (state === GithubSyncState.ERROR) return <AlertTriangle />;
  if (state === GithubSyncState.PENDING) return <LoaderCircle className="animate-spin" />;
  if (state === GithubSyncState.OK) return <CheckCircle2 />;
  return <CircleOff />;
}

export function GitHubStatusBadge({ state }: { state?: GithubSyncState }) {
  const title = githubStatusTitle(state);
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Badge variant={githubBadgeVariant(state)} title={title} aria-label={`GitHub ${title}`}>
          <BrandIcon icon={siGithub} />
          {githubIcon(state)}
        </Badge>
      </TooltipTrigger>
      <TooltipContent>{title}</TooltipContent>
    </Tooltip>
  );
}
