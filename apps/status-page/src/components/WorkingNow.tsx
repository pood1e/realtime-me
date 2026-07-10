import type { Agent, Subagent } from '@/gen/realtime/me/v1/status_pb';
import { AgentClip, SubagentBadge, agentIcon, agentName } from '@/components/AgentCard';
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
// count. Each clip names its own model in its title, so the badge below need only
// say how many are out.
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
          <SubagentBadge subagents={agent.subagents} />
        </div>
      )}
    </div>
  );
}

function agentTitle(agent: Agent): string {
  return [`${agentName(agent.kind)} working`, agent.model, agentDeviceLabel(agent)].filter(Boolean).join(' · ');
}

function subagentTitle(agent: Agent, subagent: Subagent): string {
  return [`${agentName(agent.kind)} sub-agent`, subagent.model].filter(Boolean).join(' · ');
}
