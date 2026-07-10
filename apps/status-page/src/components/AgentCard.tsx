import { AlertTriangle, Bot, CheckCircle2, CircleOff, Laptop } from 'lucide-react';
import { useEffect, useState, type ReactElement } from 'react';
import { siClaude } from 'simple-icons/icons';
import agentOrbitUrl from '@/assets/agents/agent-orbit.svg';
import codexOrbitUrl from '@/assets/agents/codex-orbit.svg';
import codexRibbonsUrl from '@/assets/agents/codex-ribbons.svg';
import codexSparksUrl from '@/assets/agents/codex-sparks.svg';
import codexSwarmUrl from '@/assets/agents/codex-swarm.svg';
import type { Agent, Subagent } from '@/gen/realtime/me/v1/status_pb';
import { AgentState } from '@/gen/realtime/me/v1/status_types_pb';
import { Badge } from '@/components/ui/badge';
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { BrandIcon } from '@/components/brand';
import { EmptyCard, InlineTime } from '@/components/layout';
import { usePrefersReducedMotion } from '@/hooks/usePrefersReducedMotion';
import { agentDeviceLabel, subagentCountLabel, subagentModelSummary } from '@/lib/status';

type AgentMotionAsset = {
  src: string;
  durationMs: number;
  // poster is a still frame shown when the viewer asks for reduced motion.
  poster?: string;
};

// A clip runs its whole loop four times, then a different one takes over. Timing
// the swap to a whole number of loops means it never lands mid-animation.
const CLIP_LOOPS_BEFORE_ROTATE = 4;

// Neither agent's clips are distributed with this repository: Claw'd is
// Anthropic's artwork and the Codex pets are OpenAI's — see
// src/assets/agents/NOTICE.md. Both are discovered at build time, so a checkout
// without them still builds and each agent falls back to an original mascot.
// durationMs is each clip's own loop length, so a clip is only swapped out on a
// whole-loop boundary rather than mid-animation.

// Every Codex pet runs the same six-frame "running" animation, so they share a
// loop length. scripts/operator/normalize-codex-pets.py prints it.
const CODEX_PET_DURATION_MS = 1_640;

const CLAWD_CLIP_DURATIONS_MS: Record<string, number> = {
  'clawd-laptop': 3_580,
  'clawd-magnifier': 9_410,
  'clawd-crab-walking': 1_660,
  'clawd-lurking': 5_580,
  'clawd-racing-car': 4_010,
  'clawd-soccer': 4_880,
  'clawd-dancing': 3_330,
  'clawd-jumping-happy': 1_760,
  'clawd-waving': 1_410,
};

const clawdClipUrls = import.meta.glob('../assets/agents/clawd/*.gif', { eager: true, import: 'default' }) as Record<string, string>;
const clawdPosterUrls = import.meta.glob('../assets/agents/clawd/*.png', { eager: true, import: 'default' }) as Record<string, string>;
const codexClipUrls = import.meta.glob('../assets/agents/codex/*.gif', { eager: true, import: 'default' }) as Record<string, string>;
const codexPosterUrls = import.meta.glob('../assets/agents/codex/*.png', { eager: true, import: 'default' }) as Record<string, string>;

const DEFAULT_MOTION_ASSETS: AgentMotionAsset[] = [{ src: agentOrbitUrl, durationMs: 4_000 }];
const CODEX_FALLBACK_ASSETS: AgentMotionAsset[] = [codexOrbitUrl, codexSparksUrl, codexRibbonsUrl, codexSwarmUrl].map((src) => ({
  src,
  durationMs: 4_000,
}));
const CLAWD_MOTION_ASSETS: AgentMotionAsset[] = clawdMotionAssets();
const CODEX_MOTION_ASSETS: AgentMotionAsset[] = motionAssets(codexClipUrls, codexPosterUrls, () => CODEX_PET_DURATION_MS);

// A clip counts only once its loop length, its animation and its reduced-motion
// still all resolve; a partial clip would either cycle on a NaN delay or animate
// at a viewer who asked it not to.
function clawdMotionAssets(): AgentMotionAsset[] {
  const urls = Object.fromEntries(
    Object.keys(CLAWD_CLIP_DURATIONS_MS).map((name) => [`../assets/agents/clawd/${name}.gif`, clawdClipUrls[`../assets/agents/clawd/${name}.gif`]]),
  );
  return motionAssets(urls, clawdPosterUrls, (name) => CLAWD_CLIP_DURATIONS_MS[name]);
}

// Pair each discovered clip with its poster and its loop length, dropping any
// that is missing either. The clips are keyed by their own file name, so a pet
// added to the asset directory joins the rotation without further wiring.
function motionAssets(clipUrls: Record<string, string>, posterUrls: Record<string, string>, durationMs: (name: string) => number): AgentMotionAsset[] {
  return Object.entries(clipUrls)
    .sort(([left], [right]) => left.localeCompare(right))
    .flatMap(([path, src]) => {
      const poster = posterUrls[path.replace(/\.gif$/, '.png')];
      const duration = durationMs(clipName(path));
      if (!src || !poster || !duration) return [];
      return [{ src, durationMs: duration, poster }];
    });
}

function clipName(path: string): string {
  return path.slice(path.lastIndexOf('/') + 1, -'.gif'.length);
}

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
        <div className="flex flex-wrap items-center gap-2">
          <AgentDeviceBadge agent={agent} />
          <SubagentBadge subagents={agent.subagents} />
        </div>
        {agent.model && <p className="truncate text-xs text-muted-foreground">{agent.model}</p>}
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
  const label = agentMotionLabel(agent);
  return (
    <div className="agent-motion" title={label}>
      <AgentClip kind={agent.kind} seed={agent.uid} className="agent-motion-image" alt={label} />
    </div>
  );
}

// AgentClip cycles an agent's clips, swapping only on a whole-loop boundary. The
// seed decides which clip it opens on, so an agent and each of its sub-agents
// start on different pictures rather than moving in lockstep.
export function AgentClip({ kind, seed, className, alt, title }: { kind: string; seed: string; className: string; alt: string; title?: string }) {
  const assets = agentMotionAssets(kind);
  const reducedMotion = usePrefersReducedMotion();
  const initialIndex = hashString(seed) % assets.length;
  const [clip, setClip] = useState(() => firstClip(initialIndex, assets.length));
  const asset = assets[clip.index % assets.length];
  const nextAsset = assets[clip.next % assets.length];

  useEffect(() => {
    setClip(firstClip(initialIndex, assets.length));
  }, [initialIndex, assets.length]);

  // Fetch the next clip while this one plays. Swapping to a clip the browser has
  // never seen would otherwise leave the box empty until it decodes.
  useEffect(() => {
    if (reducedMotion || assets.length <= 1) return;
    new Image().src = nextAsset.src;
  }, [assets.length, nextAsset, reducedMotion]);

  useEffect(() => {
    if (reducedMotion || assets.length <= 1) return;
    const timeout = window.setTimeout(() => {
      setClip((current) => ({ index: current.next, next: nextClipIndex(current.next, assets.length) }));
    }, asset.durationMs * CLIP_LOOPS_BEFORE_ROTATE);
    return () => window.clearTimeout(timeout);
  }, [asset, assets.length, reducedMotion]);

  // A GIF animates no matter what the stylesheet says, so a viewer who asked for
  // reduced motion gets a still frame instead.
  const src = reducedMotion ? asset.poster ?? asset.src : asset.src;
  return <img key={src} className={className} src={src} alt={alt} title={title} />;
}

// The card no longer spells the agent out, so the label carries what the picture
// cannot: which agent, and how much budget it has left.
export function agentMotionLabel(agent: Agent): string {
  const name = `${agentName(agent.kind)} working`;
  return agent.budgetRemainingPercent === undefined ? name : `${name} · ${agent.budgetRemainingPercent}% budget left`;
}

// The badge counts the sub-agents out; the models they run are named in its
// title. A sub-agent almost always runs the model that spawned it, which the
// agent already shows, so the count is the news and the model is the exception.
export function SubagentBadge({ subagents }: { subagents: Subagent[] }) {
  if (subagents.length === 0) return null;
  const count = subagentCountLabel(subagents.length);
  const models = subagentModelSummary(subagents);
  return (
    <Badge variant="secondary" className="min-w-0 shrink font-normal" title={models || count}>
      <span className="truncate">{count}</span>
    </Badge>
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
  if (isClaudeAgent(kind) && CLAWD_MOTION_ASSETS.length > 0) return CLAWD_MOTION_ASSETS;
  if (kind === 'codex') return CODEX_MOTION_ASSETS.length > 0 ? CODEX_MOTION_ASSETS : CODEX_FALLBACK_ASSETS;
  return DEFAULT_MOTION_ASSETS;
}

// The clip playing now, and the one after it. Choosing the successor up front is
// what lets it be fetched before it is shown.
function firstClip(index: number, length: number): { index: number; next: number } {
  return { index, next: nextClipIndex(index, length) };
}

// Any clip but the one just shown, drawn uniformly, so a rotation always changes
// the picture and the order never repeats itself.
function nextClipIndex(current: number, length: number): number {
  if (length <= 1) return current;
  const offset = 1 + Math.floor(Math.random() * (length - 1));
  return (current + offset) % length;
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
