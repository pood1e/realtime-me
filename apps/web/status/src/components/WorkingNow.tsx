import type { Agent, Subagent } from '@realtime-me/status-contracts';
import { AgentClip, agentName } from '@/components/AgentCard';
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
// count and the crowd is all of it. What each mascot is, which model it runs and
// which machine it runs on are written into its own title: a public page owes a
// passer-by a glance, and whoever wants the readings behind it has the dashboard.
function WorkingAgent({ agent }: { agent: Agent }) {
  return (
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
  );
}

function agentTitle(agent: Agent): string {
  return [`${agentName(agent.kind)} working`, agent.model, agentDeviceLabel(agent)].filter(Boolean).join(' · ');
}

function subagentTitle(agent: Agent, subagent: Subagent): string {
  return [`${agentName(agent.kind)} sub-agent`, subagent.model].filter(Boolean).join(' · ');
}
