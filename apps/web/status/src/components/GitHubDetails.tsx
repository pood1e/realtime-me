import type { GithubSyncDetail } from "@realtime-me/status-contracts";
import { Badge } from "@realtime-me/web-ui/badge";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@realtime-me/web-ui/card";
import { AlertTriangle } from "lucide-react";
import { siGithub } from "simple-icons/icons";
import { BrandIcon } from "@/components/brand";
import { githubBadgeVariant, githubIcon, githubStatusTitle } from "@/components/github";
import { InlineTime } from "@/components/layout";
import { formatDateTime } from "@/lib/format";

export function GitHubDetails({ github }: { github: GithubSyncDetail }) {
  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <BrandIcon icon={siGithub} />
            GitHub status sync
          </CardTitle>
          <CardAction>
            <Badge
              variant={githubBadgeVariant(github.state)}
              aria-label={githubStatusTitle(github.state)}
            >
              {githubIcon(github.state)}
            </Badge>
          </CardAction>
        </CardHeader>
        <CardContent className="grid gap-3 text-sm">
          <DetailRow label="State" value={githubStatusTitle(github.state)} />
          <DetailRow label="Last success" value={formatDateTime(github.lastSuccessTime)} />
          <DetailRow label="Last attempt" value={formatDateTime(github.lastAttemptTime)} />
          <DetailRow label="Emoji" value={github.emoji || "—"} />
          <DetailRow label="Message" value={github.message || "—"} />
        </CardContent>
      </Card>
      {github.lastError && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-destructive">
              <AlertTriangle className="size-4" />
              Last error
            </CardTitle>
            <CardAction>
              <InlineTime value={github.lastErrorTime} />
            </CardAction>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">{github.lastError}</CardContent>
        </Card>
      )}
    </div>
  );
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3 border-b border-border/50 pb-2 last:border-0 last:pb-0">
      <span className="text-muted-foreground">{label}</span>
      <span className="min-w-0 truncate text-right font-medium">{value}</span>
    </div>
  );
}
