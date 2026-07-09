import { siGithub } from 'simple-icons/icons';
import { useOutletContext } from 'react-router-dom';
import type { ShellContext } from '@/components/AppShell';
import { BrandIcon } from '@/components/brand';
import { EmptyCard, ErrorCard, LoadingCard, StatusSection } from '@/components/layout';
import { ProjectTimeline } from '@/components/ProfileCards';

export function ProfileApp() {
  const { page, pageFailed, retryPage } = useOutletContext<ShellContext>();
  const projects = page?.projects ?? [];

  return (
    <StatusSection title="Projects" icon={<BrandIcon icon={siGithub} />} columns="grid-cols-1">
      {body()}
    </StatusSection>
  );

  // The profile is fetched once, so a single failed request would otherwise spin
  // forever. Say so, and offer the retry the visitor would reach for anyway.
  function body() {
    if (pageFailed && page == null) return <ErrorCard text="Cannot load the profile right now." retry={retryPage} />;
    if (page == null) return <LoadingCard />;
    if (projects.length === 0) return <EmptyCard text="No projects yet" />;
    return <ProjectTimeline projects={projects} />;
  }
}
