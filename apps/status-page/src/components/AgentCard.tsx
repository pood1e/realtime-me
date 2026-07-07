import { AlertTriangle, Bot, CheckCircle2, CircleOff, Laptop } from 'lucide-react';
import { useEffect, useState, type ReactElement } from 'react';
import { siClaude } from 'simple-icons/icons';
import agentOrbitUrl from '@/assets/agents/agent-orbit.svg';
import clawdBuildingUrl from '@/assets/agents/clawd-working-building.gif';
import clawdDebuggerUrl from '@/assets/agents/clawd-working-debugger.gif';
import clawdJugglingUrl from '@/assets/agents/clawd-working-juggling.gif';
import clawdSweepingUrl from '@/assets/agents/clawd-working-sweeping.gif';
import clawdThinkingUrl from '@/assets/agents/clawd-working-thinking.gif';
import clawdTypingUrl from '@/assets/agents/clawd-working-typing.gif';
import codexOrbitUrl from '@/assets/agents/codex-orbit.svg';
import codexRibbonsUrl from '@/assets/agents/codex-ribbons.svg';
import codexSparksUrl from '@/assets/agents/codex-sparks.svg';
import type { Agent } from '@/gen/realtime/me/v1/status_pb';
import { AgentState } from '@/gen/realtime/me/v1/status_types_pb';
import { Badge } from '@/components/ui/badge';
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { BrandIcon } from '@/components/brand';
import { EmptyCard, InlineTime } from '@/components/layout';
import { agentDeviceLabel } from '@/lib/status';

type AgentMotionAsset = {
  src: string;
  durationMs: number;
};

const AGENT_MOTION_MIN_VISIBLE_MS = 10_000;
const CLAWD_MOTION_ASSETS: AgentMotionAsset[] = [
  { src: clawdTypingUrl, durationMs: 1_440 },
  { src: clawdBuildingUrl, durationMs: 960 },
  { src: clawdDebuggerUrl, durationMs: 2_880 },
  { src: clawdThinkingUrl, durationMs: 3_840 },
  { src: clawdSweepingUrl, durationMs: 1_440 },
  { src: clawdJugglingUrl, durationMs: 1_120 },
];
const CODEX_MOTION_ASSETS: AgentMotionAsset[] = [
  { src: codexOrbitUrl, durationMs: 4_000 },
  { src: codexRibbonsUrl, durationMs: 4_000 },
  { src: codexSparksUrl, durationMs: 4_000 },
];
const DEFAULT_MOTION_ASSETS: AgentMotionAsset[] = [{ src: agentOrbitUrl, durationMs: 4_000 }];

export function AgentCard({ agent }: { agent: Agent }) {
  const stateLabel = agentStateLabel(agent.state);
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex min-w-0 items-center gap-2">
          {agentIcon(agent.kind)}
          <span className="truncate">{agentName(agent.kind)}</span>
        </CardTitle>
        <CardAction className="flex items-center gap-2">
          <InlineTime value={agent.updateTime} />
          <Badge variant={agentBadgeVariant(agent.state)} title={stateLabel} aria-label={stateLabel}>{agentStateIcon(agent.state)}</Badge>
        </CardAction>
      </CardHeader>
      <CardContent className="grid gap-4">
        <AgentMotion agent={agent} />
        <AgentDeviceBadge agent={agent} />
        {agent.budgetRemainingPercent !== undefined && (
          <div className="grid gap-2">
            <div className="flex items-center justify-between gap-3 text-xs text-muted-foreground">
              <span>Budget</span>
              <span>{agent.budgetRemainingPercent}%</span>
            </div>
            <Progress value={agent.budgetRemainingPercent} />
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export function EmptyAgentCard() {
  return <EmptyCard text="No active agents" />;
}

export function agentKey(agent: Agent): string {
  return agent.uid;
}

export function agentName(kind: string): string {
  if (isClaudeAgent(kind)) return 'Claude Code';
  if (kind === 'codex') return 'Codex';
  return kind || 'Agent';
}

export function agentIcon(kind: string): ReactElement {
  if (isClaudeAgent(kind)) return <BrandIcon icon={siClaude} />;
  if (kind === 'codex') return <CodexIcon />;
  return <Bot className="size-4" />;
}

function AgentMotion({ agent }: { agent: Agent }) {
  const assets = agentMotionAssets(agent.kind);
  const initialIndex = hashString(agent.uid) % assets.length;
  const [index, setIndex] = useState(initialIndex);
  const asset = assets[index % assets.length];
  const imageClassName = isClaudeAgent(agent.kind) ? 'agent-motion-image agent-motion-image-pixel' : 'agent-motion-image';

  useEffect(() => {
    setIndex(initialIndex);
  }, [initialIndex]);

  useEffect(() => {
    if (assets.length <= 1) return;
    const timeout = window.setTimeout(() => {
      setIndex((current) => (current + 1) % assets.length);
    }, agentMotionDelayMs(asset));
    return () => window.clearTimeout(timeout);
  }, [asset, assets.length, index]);

  return (
    <div className="agent-motion">
      <img key={asset.src} className={imageClassName} src={asset.src} alt={`${agentName(agent.kind)} working`} />
    </div>
  );
}

function AgentDeviceBadge({ agent }: { agent: Agent }) {
  const device = agentDeviceLabel(agent);
  if (!device) return null;
  return (
    <Badge variant="outline" className="min-w-0 shrink" title={device}>
      <Laptop />
      <span className="truncate">{device}</span>
    </Badge>
  );
}

function agentMotionAssets(kind: string): AgentMotionAsset[] {
  if (isClaudeAgent(kind)) return CLAWD_MOTION_ASSETS;
  if (kind === 'codex') return CODEX_MOTION_ASSETS;
  return DEFAULT_MOTION_ASSETS;
}

function agentMotionDelayMs(asset: AgentMotionAsset): number {
  if (asset.durationMs <= 0) return AGENT_MOTION_MIN_VISIBLE_MS;
  return Math.ceil(AGENT_MOTION_MIN_VISIBLE_MS / asset.durationMs) * asset.durationMs;
}

function agentStateIcon(state: AgentState): ReactElement {
  if (state === AgentState.FAILED) return <AlertTriangle />;
  if (state === AgentState.RUNNING) return <CheckCircle2 />;
  return <CircleOff />;
}

function agentBadgeVariant(state: AgentState): 'default' | 'secondary' | 'destructive' {
  if (state === AgentState.FAILED) return 'destructive';
  if (state === AgentState.RUNNING) return 'default';
  return 'secondary';
}

function agentStateLabel(state: AgentState): string {
  if (state === AgentState.RUNNING) return 'running';
  if (state === AgentState.FAILED) return 'failed';
  if (state === AgentState.IDLE) return 'idle';
  return 'unknown';
}

function isClaudeAgent(kind: string): boolean {
  return kind === 'claude';
}

function hashString(value: string): number {
  let hash = 0;
  for (const character of value) {
    hash = Math.imul(31, hash) + character.charCodeAt(0);
  }
  return Math.abs(hash);
}

function CodexIcon({ className = 'size-4' }: { className?: string }) {
  return (
    <svg aria-label="Codex" className={`${className} shrink-0`} fill="currentColor" fillRule="evenodd" role="img" viewBox="0 0 24 24">
      <title>Codex</title>
      <path
        clipRule="evenodd"
        d="M8.086.457a6.105 6.105 0 013.046-.415c1.333.153 2.521.72 3.564 1.7a.117.117 0 00.107.029c1.408-.346 2.762-.224 4.061.366l.063.03.154.076c1.357.703 2.33 1.77 2.918 3.198.278.679.418 1.388.421 2.126a5.655 5.655 0 01-.18 1.631.167.167 0 00.04.155 5.982 5.982 0 011.578 2.891c.385 1.901-.01 3.615-1.183 5.14l-.182.22a6.063 6.063 0 01-2.934 1.851.162.162 0 00-.108.102c-.255.736-.511 1.364-.987 1.992-1.199 1.582-2.962 2.462-4.948 2.451-1.583-.008-2.986-.587-4.21-1.736a.145.145 0 00-.14-.032c-.518.167-1.04.191-1.604.185a5.924 5.924 0 01-2.595-.622 6.058 6.058 0 01-2.146-1.781c-.203-.269-.404-.522-.551-.821a7.74 7.74 0 01-.495-1.283 6.11 6.11 0 01-.017-3.064.166.166 0 00.008-.074.115.115 0 00-.037-.064 5.958 5.958 0 01-1.38-2.202 5.196 5.196 0 01-.333-1.589 6.915 6.915 0 01.188-2.132c.45-1.484 1.309-2.648 2.577-3.493.282-.188.55-.334.802-.438.286-.12.573-.22.861-.304a.129.129 0 00.087-.087A6.016 6.016 0 015.635 2.31C6.315 1.464 7.132.846 8.086.457zm-.804 7.85a.848.848 0 00-1.473.842l1.694 2.965-1.688 2.848a.849.849 0 001.46.864l1.94-3.272a.849.849 0 00.007-.854l-1.94-3.393zm5.446 6.24a.849.849 0 000 1.695h4.848a.849.849 0 000-1.696h-4.848z"
      />
    </svg>
  );
}
