import { Clock, Code, ExternalLink, Globe, Lock, Mail, MapPin, Star } from 'lucide-react';
import type { ReactElement } from 'react';
import { siDiscord, siGithub, siGmail, siTelegram } from 'simple-icons/icons';
import type { Profile, ProfileLink, Project } from '@/gen/realtime/me/v1/profile_pb';
import { ProjectVisibility } from '@/gen/realtime/me/v1/profile_pb';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { BrandIcon } from '@/components/brand';
import { EmptyCard, LoadingCard } from '@/components/layout';
import { formatDateTime } from '@/lib/format';

export function ProfileIntro({ profile, loaded }: { profile?: Profile; loaded: boolean }) {
  if (!loaded) return <LoadingCard />;
  if (!profile) return <EmptyCard text="Profile not configured" />;
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-4">
          {profile.avatarUrl && (
            <img src={profile.avatarUrl} alt={profile.displayName || 'avatar'} className="size-16 rounded-full border border-border" width={64} height={64} />
          )}
          <div className="grid gap-1">
            <CardTitle className="font-heading text-3xl tracking-tight">{profile.displayName || '—'}</CardTitle>
            {profile.headline && <CardDescription className="text-base">{profile.headline}</CardDescription>}
          </div>
        </div>
      </CardHeader>
      <CardContent className="grid gap-4">
        {profile.bio && <p className="whitespace-pre-line text-sm text-muted-foreground">{profile.bio}</p>}
        <div className="flex flex-wrap items-center gap-2">
          {profile.location && <Badge variant="outline"><MapPin />{profile.location}</Badge>}
          {profile.githubLogin && (
            <a href={`https://github.com/${profile.githubLogin}`} target="_blank" rel="noreferrer" aria-label="GitHub profile">
              <Badge variant="secondary"><BrandIcon icon={siGithub} />{profile.githubLogin}</Badge>
            </a>
          )}
        </div>
        <ProfileLinks links={profile.links} />
      </CardContent>
    </Card>
  );
}

function ProfileLinks({ links }: { links: ProfileLink[] }) {
  if (links.length === 0) return null;
  return (
    <div className="flex flex-wrap gap-2">
      {links.map((link) => (
        <Button key={`${link.platform}:${link.uri}`} asChild variant="outline" size="sm">
          <a href={link.uri} target="_blank" rel="noreferrer">
            {linkIcon(link)}
            {link.label || link.platform || link.uri}
          </a>
        </Button>
      ))}
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
          <span className="truncate">{project.displayName || '—'}</span>
        </CardTitle>
        <CardAction>
          <Badge variant={isPrivate ? 'secondary' : 'outline'} title={isPrivate ? 'Private' : 'Public'}>
            {isPrivate ? <Lock /> : <BrandIcon icon={siGithub} />}
            {isPrivate ? 'Private' : 'Public'}
          </Badge>
        </CardAction>
      </CardHeader>
      <CardContent className="grid gap-3">
        {blurb && <p className="text-sm text-muted-foreground">{blurb}</p>}
        {project.topics.length > 0 && (
          <div className="flex flex-wrap gap-1.5">
            {project.topics.map((topic) => <Badge key={topic} variant="outline">{topic}</Badge>)}
          </div>
        )}
        <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
          {project.primaryLanguage && <span className="flex items-center gap-1"><Code className="size-3.5" />{project.primaryLanguage}</span>}
          {!!project.starCount && <span className="flex items-center gap-1"><Star className="size-3.5" />{project.starCount}</span>}
          {project.lastPushTime && <span className="flex items-center gap-1"><Clock className="size-3.5" />{formatDateTime(project.lastPushTime)}</span>}
        </div>
        <ProjectLinks project={project} />
      </CardContent>
    </Card>
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

function linkIcon(link: ProfileLink): ReactElement {
  const platform = link.platform.toLowerCase();
  if (platform === 'github') return <BrandIcon icon={siGithub} />;
  if (platform === 'telegram') return <BrandIcon icon={siTelegram} />;
  if (platform === 'discord') return <BrandIcon icon={siDiscord} />;
  if (platform === 'email' || link.uri.startsWith('mailto:')) {
    return link.uri.includes('gmail') ? <BrandIcon icon={siGmail} /> : <Mail />;
  }
  return <Globe />;
}
