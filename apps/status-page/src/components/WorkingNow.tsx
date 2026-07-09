import { Laptop } from 'lucide-react';
import type { Agent } from '@/gen/realtime/me/v1/status_pb';
import { AgentClip, agentIcon, agentMotionLabel, agentName } from '@/components/AgentCard';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { agentDeviceLabel } from '@/lib/status';

// Working agents live here rather than on the card of the machine they run on:
// a device can host several at once, and stacking them beside its name crowds
// the title and pushes that card's readings out of line with its neighbours.
// The band disappears entirely when nothing is working, so an idle fleet costs
// no space at all.
export function WorkingNow({ agents }: { agents: Agent[] }) {
  if (agents.length === 0) return null;
  return (
    <Card>
      <CardContent className="flex flex-wrap items-end gap-x-8 gap-y-6 py-4">
        {agents.map((agent) => (
          <WorkingAgent key={agent.uid} agent={agent} />
        ))}
      </CardContent>
    </Card>
  );
}

function WorkingAgent({ agent }: { agent: Agent }) {
  const device = agentDeviceLabel(agent);
  return (
    <div className="flex min-w-0 items-end gap-3">
      <AgentClip agent={agent} className="working-agent-image" alt={agentMotionLabel(agent)} />
      <div className="grid min-w-0 gap-1.5 pb-0.5">
        <div className="flex min-w-0 items-center gap-2 text-sm font-medium">
          {agentIcon(agent.kind)}
          <span className="truncate">{agentName(agent.kind)}</span>
        </div>
        {agent.model && <p className="truncate text-xs text-muted-foreground">{agent.model}</p>}
        <div className="flex flex-wrap items-center gap-2">
          {device && (
            <Badge variant="outline" className="min-w-0 shrink" title={device}>
              <Laptop />
              <span className="truncate">{device}</span>
            </Badge>
          )}
          {agent.subagentCount > 0 && <Badge variant="secondary">{subagentText(agent.subagentCount)}</Badge>}
        </div>
      </div>
    </div>
  );
}

function subagentText(count: number): string {
  return count === 1 ? '1 sub-agent' : `${count} sub-agents`;
}
