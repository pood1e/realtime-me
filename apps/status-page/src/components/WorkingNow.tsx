import type { Agent, Subagent } from '@/gen/realtime/me/v1/status_pb';
import { AgentClip, agentIcon, agentName, subagentText } from '@/components/AgentCard';
import { Badge } from '@/components/ui/badge';
import { agentDeviceLabel } from '@/lib/status';

// Working agents live here rather than on the card of the machine they run on: a
// device can host several at once, and stacking them beside its name crowds the
// title and pushes that card's readings out of line with its neighbours. The
// strip carries no card of its own — the mascots are the surface — and it
// disappears entirely when nothing is working.
export function WorkingNow({ agents }: { agents: Agent[] }) {
  if (agents.length === 0) return null;
  return (
    <div className="flex flex-wrap items-end gap-x-10 gap-y-4">
      {agents.map((agent) => (
        <WorkingAgent key={agent.uid} agent={agent} />
      ))}
    </div>
  );
}

// One clip for the agent, one for each sub-agent it has out, so the crowd is the
// count. A sub-agent need not run the model that spawned it, so the models are
// named on the page rather than left to a tooltip nobody hovers.
function WorkingAgent({ agent }: { agent: Agent }) {
  const device = agentDeviceLabel(agent);
  return (
    <div className="flex min-w-0 flex-col gap-1.5">
      <div className="flex items-end gap-1">
        <AgentClip kind={agent.kind} seed={agent.uid} className="working-agent-image" alt={agentTitle(agent)} title={agentTitle(agent)} />
        {agent.subagents.map((subagent, index) => (
          <AgentClip
            key={index}
            kind={agent.kind}
            seed={`${agent.uid}:${index}`}
            className="working-subagent-image"
            alt={subagentTitle(agent, subagent)}
            title={subagentTitle(agent, subagent)}
          />
        ))}
      </div>
      <div className="flex min-w-0 items-center gap-1.5 text-xs text-muted-foreground">
        <span className="[&_svg]:size-3.5">{agentIcon(agent.kind)}</span>
        <span className="truncate">{[agent.model || agentName(agent.kind), device].filter(Boolean).join(' · ')}</span>
      </div>
      {agent.subagents.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {subagentModelCounts(agent.subagents).map(({ model, count }) => (
            <Badge key={model} variant="secondary" className="font-normal">
              {model ? `${count} × ${model}` : subagentText(count)}
            </Badge>
          ))}
        </div>
      )}
    </div>
  );
}

// The sub-agents an agent has out, grouped by the model each runs, busiest first.
function subagentModelCounts(subagents: Subagent[]): Array<{ model: string; count: number }> {
  const counts = new Map<string, number>();
  for (const subagent of subagents) counts.set(subagent.model, (counts.get(subagent.model) ?? 0) + 1);
  return [...counts]
    .map(([model, count]) => ({ model, count }))
    .sort((left, right) => right.count - left.count || left.model.localeCompare(right.model));
}

function agentTitle(agent: Agent): string {
  return [`${agentName(agent.kind)} working`, agent.model, agentDeviceLabel(agent)].filter(Boolean).join(' · ');
}

function subagentTitle(agent: Agent, subagent: Subagent): string {
  return [`${agentName(agent.kind)} sub-agent`, subagent.model].filter(Boolean).join(' · ');
}
