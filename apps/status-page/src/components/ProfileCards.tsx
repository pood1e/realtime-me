import { Archive, CalendarPlus, Clock, ExternalLink, Lock, Star } from 'lucide-react';
import { siGithub } from 'simple-icons/icons';
import type { LanguageShare, Project } from '@/gen/realtime/me/v1/profile_pb';
import { ProjectVisibility } from '@/gen/realtime/me/v1/profile_pb';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { BrandIcon } from '@/components/brand';
import { formatDateTime, formatMonthYear } from '@/lib/format';

export function ProjectCard({ project }: { project: Project }) {
  const isPrivate = project.visibility === ProjectVisibility.PRIVATE;
  const blurb = project.summary || project.description;
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex min-w-0 items-center gap-2">
          <span className="truncate">{project.displayName || '—'}</span>
        </CardTitle>
        <CardAction className="flex items-center gap-1.5">
          {project.archived && (
            <Badge variant="outline" title="Archived">
              <Archive />
              Archived
            </Badge>
          )}
          <Badge variant={isPrivate ? 'secondary' : 'outline'} title={isPrivate ? 'Private' : 'Public'}>
            {isPrivate ? <Lock /> : <BrandIcon icon={siGithub} />}
            {isPrivate ? 'Private' : 'Public'}
          </Badge>
        </CardAction>
      </CardHeader>
      <CardContent className="grid gap-3">
        {blurb && <p className="text-sm text-muted-foreground">{blurb}</p>}
        <LanguageBar languages={project.languages} />
        {project.topics.length > 0 && (
          <div className="flex flex-wrap gap-1.5">
            {project.topics.slice(0, 6).map((topic) => (
              <Badge key={topic} variant="outline">{topic}</Badge>
            ))}
          </div>
        )}
        <CommitSparkline weeks={project.commitActivity} />
        <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
          {!!project.starCount && (
            <span className="flex items-center gap-1"><Star className="size-3.5" />{project.starCount}</span>
          )}
          {project.createTime && (
            <span className="flex items-center gap-1" title="Created"><CalendarPlus className="size-3.5" />{formatMonthYear(project.createTime)}</span>
          )}
          {project.lastPushTime && (
            <span className="flex items-center gap-1" title="Last push"><Clock className="size-3.5" />{formatDateTime(project.lastPushTime)}</span>
          )}
        </div>
        <ProjectLinks project={project} />
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
            style={{ width: `${(Number(language.bytes) / total) * 100}%`, backgroundColor: languageColor(language.name) }}
          />
        ))}
      </div>
      <div className="flex flex-wrap gap-x-3 gap-y-0.5 text-[11px] text-muted-foreground">
        {top.map((language) => (
          <span key={language.name} className="flex items-center gap-1">
            <span className="size-1.5 rounded-full" style={{ backgroundColor: languageColor(language.name) }} />
            {language.name}
            <span className="tabular-nums">{Math.round((Number(language.bytes) / total) * 100)}%</span>
          </span>
        ))}
      </div>
    </div>
  );
}

function CommitSparkline({ weeks }: { weeks: number[] }) {
  if (!weeks || weeks.length === 0) return null;
  const total = weeks.reduce((sum, count) => sum + count, 0);
  if (total <= 0) return null;
  const max = Math.max(...weeks, 1);
  return (
    <div className="flex h-6 items-end gap-px" title={`${total} commits in the last year`} aria-label="Commit activity, last year">
      {weeks.map((count, index) => (
        <div
          key={index}
          className="flex-1 rounded-[1px] bg-primary/50"
          style={{ height: `${Math.max((count / max) * 100, 6)}%` }}
        />
      ))}
    </div>
  );
}

function ProjectLinks({ project }: { project: Project }) {
  if (!project.repositoryUrl && !project.homepageUrl) return null;
  return (
    <div className="flex flex-wrap gap-2">
      {project.repositoryUrl && (
        <Button asChild variant="outline" size="sm">
          <a href={project.repositoryUrl} target="_blank" rel="noreferrer"><BrandIcon icon={siGithub} />Repository</a>
        </Button>
      )}
      {project.homepageUrl && (
        <Button asChild variant="secondary" size="sm">
          <a href={project.homepageUrl} target="_blank" rel="noreferrer"><ExternalLink />Homepage</a>
        </Button>
      )}
    </div>
  );
}

const LANGUAGE_COLORS: Record<string, string> = {
  TypeScript: '#3178c6',
  JavaScript: '#f1e05a',
  Go: '#00add8',
  Python: '#3572a5',
  Kotlin: '#a97bff',
  Java: '#b07219',
  Shell: '#89e051',
  'C++': '#f34b7d',
  C: '#555555',
  'C#': '#178600',
  Rust: '#dea584',
  Ruby: '#701516',
  Dart: '#00b4ab',
  Swift: '#f05138',
  HTML: '#e34c26',
  CSS: '#563d7c',
  Vue: '#41b883',
  PHP: '#4f5d95',
  Dockerfile: '#384d54',
  Makefile: '#427819',
  Lua: '#000080',
  Nix: '#7e7eff',
};

function languageColor(name: string): string {
  return LANGUAGE_COLORS[name] ?? '#94a3b8';
}
