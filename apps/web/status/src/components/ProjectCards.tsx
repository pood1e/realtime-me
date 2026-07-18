import type { LanguageShare, Project } from "@realtime-me/status-contracts";
import { ProjectVisibility } from "@realtime-me/status-contracts";
import { Badge } from "@realtime-me/web-ui/badge";
import { Button } from "@realtime-me/web-ui/button";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@realtime-me/web-ui/card";
import { Archive, CalendarPlus, Clock, ExternalLink, Loader2, Lock, Star } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import type { SimpleIcon } from "simple-icons";
import {
  siC,
  siCplusplus,
  siCss,
  siDart,
  siDocker,
  siGithub,
  siGnubash,
  siGo,
  siHtml5,
  siJavascript,
  siKotlin,
  siLua,
  siNixos,
  siOpenjdk,
  siPhp,
  siPython,
  siRuby,
  siRust,
  siSharp,
  siSwift,
  siTypescript,
  siVuedotjs,
} from "simple-icons/icons";
import { BrandIcon } from "@/components/brand";
import { formatDateTime, formatMonthYear } from "@/lib/format";

const TIMELINE_BATCH = 10;

export function ProjectTimeline({ projects }: { projects: Project[] }) {
  const [visible, setVisible] = useState(TIMELINE_BATCH);
  const sentinelRef = useRef<HTMLDivElement | null>(null);
  const hasMore = visible < projects.length;

  useEffect(() => {
    if (!hasMore) return;
    const sentinel = sentinelRef.current;
    if (!sentinel) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) setVisible((current) => current + TIMELINE_BATCH);
      },
      { rootMargin: "300px" },
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [hasMore]);

  const shown = projects.slice(0, visible);
  return (
    <div className="mx-auto w-full max-w-3xl">
      <ol className="relative ml-2 border-l border-border/70">
        {shown.map((project, index) => {
          const label = formatMonthYear(project.lastPushTime);
          const showDate = index === 0 || label !== formatMonthYear(shown[index - 1].lastPushTime);
          return (
            <li key={project.uid || project.displayName} className="relative pb-8 pl-7 last:pb-0">
              <span className="absolute -left-[5px] top-2 size-2.5 rounded-full border-2 border-background bg-primary" />
              {showDate && (
                <time className="mb-2 block text-xs font-semibold tracking-wide text-muted-foreground">
                  {label}
                </time>
              )}
              <ProjectCard project={project} />
            </li>
          );
        })}
      </ol>
      {hasMore && (
        <div
          ref={sentinelRef}
          className="flex items-center justify-center gap-2 py-6 text-xs text-muted-foreground"
        >
          <Loader2 className="size-3.5 animate-spin" />
          {projects.length - visible} more
        </div>
      )}
    </div>
  );
}

export function ProjectCard({ project }: { project: Project }) {
  const isPrivate = project.visibility === ProjectVisibility.PRIVATE;
  const blurb = project.summary || project.description;
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex min-w-0 items-center gap-2">
          <span className="truncate">{project.displayName || "—"}</span>
        </CardTitle>
        <CardAction className="flex items-center gap-1">
          {project.archived && (
            <Badge variant="outline" title="Archived">
              <Archive />
              Archived
            </Badge>
          )}
          {isPrivate ? (
            <Badge variant="secondary" title="Private">
              <Lock />
              Private
            </Badge>
          ) : (
            <div className="flex items-center gap-0.5 text-muted-foreground">
              {project.homepageUrl && (
                <Button
                  asChild
                  variant="ghost"
                  size="icon"
                  aria-label="Homepage"
                  title="Homepage"
                  className="text-muted-foreground hover:text-foreground"
                >
                  <a href={project.homepageUrl} target="_blank" rel="noreferrer">
                    <ExternalLink />
                  </a>
                </Button>
              )}
              {project.repositoryUrl && (
                <Button
                  asChild
                  variant="ghost"
                  size="icon"
                  aria-label="Repository"
                  title="Repository"
                  className="text-muted-foreground hover:text-foreground"
                >
                  <a href={project.repositoryUrl} target="_blank" rel="noreferrer">
                    <BrandIcon icon={siGithub} mono />
                  </a>
                </Button>
              )}
            </div>
          )}
        </CardAction>
      </CardHeader>
      <CardContent className="grid gap-3">
        {blurb && <p className="text-sm text-muted-foreground">{blurb}</p>}
        <LanguageBar languages={project.languages} />
        {project.topics.length > 0 && (
          <div className="flex flex-wrap gap-1.5">
            {project.topics.slice(0, 6).map((topic) => (
              <Badge key={topic} variant="outline">
                {topic}
              </Badge>
            ))}
          </div>
        )}
        <CommitSparkline weeks={project.commitActivity} />
        <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
          {!!project.starCount && (
            <span className="flex items-center gap-1">
              <Star className="size-3.5" />
              {project.starCount}
            </span>
          )}
          {project.createTime && (
            <span className="flex items-center gap-1" title="Created">
              <CalendarPlus className="size-3.5" />
              {formatMonthYear(project.createTime)}
            </span>
          )}
          {project.lastPushTime && (
            <span className="flex items-center gap-1" title="Last push">
              <Clock className="size-3.5" />
              {formatDateTime(project.lastPushTime)}
            </span>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

function LanguageBar({ languages }: { languages: LanguageShare[] }) {
  if (!languages || languages.length === 0) return null;
  const total = languages.reduce((sum, language) => sum + Number(language.bytes), 0);
  if (total <= 0) return null;
  const top = languages.slice(0, 4);
  return (
    <div className="grid gap-1.5">
      <div className="flex h-1.5 gap-0.5 overflow-hidden rounded-full">
        {top.map((language) => (
          <div
            key={language.name}
            style={{
              width: `${(Number(language.bytes) / total) * 100}%`,
              backgroundColor: languageColor(language.name),
            }}
          />
        ))}
      </div>
      <div className="flex flex-wrap gap-x-3 gap-y-0.5 text-[11px] text-muted-foreground">
        {top.map((language) => {
          const icon = LANGUAGE_ICONS[language.name];
          return (
            <span key={language.name} className="flex items-center gap-1">
              {icon ? (
                <BrandIcon icon={icon} className="size-3" />
              ) : (
                <span
                  className="size-1.5 rounded-full"
                  style={{ backgroundColor: languageColor(language.name) }}
                />
              )}
              {language.name}
              <span className="tabular-nums">
                {Math.round((Number(language.bytes) / total) * 100)}%
              </span>
            </span>
          );
        })}
      </div>
    </div>
  );
}

function CommitSparkline({ weeks }: { weeks: number[] }) {
  if (!weeks || weeks.length === 0) return null;
  const total = weeks.reduce((sum, count) => sum + count, 0);
  if (total <= 0) return null;
  const max = Math.max(...weeks, 1);
  const points = weeks
    .map((count, week) => `${week},${100 - Math.max((count / max) * 100, 6)}`)
    .join(" ");
  return (
    <svg
      role="img"
      aria-label="Commit activity, last year"
      viewBox={`0 0 ${Math.max(weeks.length - 1, 1)} 100`}
      preserveAspectRatio="none"
      className="h-6 w-full text-primary/60"
    >
      <title>{total} commits in the last year</title>
      <polyline
        points={points}
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        vectorEffect="non-scaling-stroke"
      />
    </svg>
  );
}

const LANGUAGE_ICONS: Record<string, SimpleIcon> = {
  TypeScript: siTypescript,
  JavaScript: siJavascript,
  Go: siGo,
  Python: siPython,
  Kotlin: siKotlin,
  Java: siOpenjdk,
  Rust: siRust,
  Ruby: siRuby,
  Dart: siDart,
  Swift: siSwift,
  HTML: siHtml5,
  CSS: siCss,
  Vue: siVuedotjs,
  PHP: siPhp,
  Shell: siGnubash,
  Dockerfile: siDocker,
  Lua: siLua,
  Nix: siNixos,
  C: siC,
  "C++": siCplusplus,
  "C#": siSharp,
};

const LANGUAGE_COLORS: Record<string, string> = {
  TypeScript: "#3178c6",
  JavaScript: "#f1e05a",
  Go: "#00add8",
  Python: "#3572a5",
  Kotlin: "#a97bff",
  Java: "#b07219",
  Shell: "#89e051",
  "C++": "#f34b7d",
  C: "#555555",
  "C#": "#178600",
  Rust: "#dea584",
  Ruby: "#701516",
  Dart: "#00b4ab",
  Swift: "#f05138",
  HTML: "#e34c26",
  CSS: "#563d7c",
  Vue: "#41b883",
  PHP: "#4f5d95",
  Dockerfile: "#384d54",
  Makefile: "#427819",
  Lua: "#000080",
  Nix: "#7e7eff",
};

function languageColor(name: string): string {
  return LANGUAGE_COLORS[name] ?? "#94a3b8";
}
