import { siGithub } from 'simple-icons/icons';
import { useOutletContext } from 'react-router-dom';
import type { ShellContext } from '@/components/AppShell';
import { BrandIcon } from '@/components/brand';
import { EmptyCard, LoadingCard, StatusSection } from '@/components/layout';
import { ProjectTimeline } from '@/components/ProfileCards';

export function ProfileApp() {
  const { page } = useOutletContext<ShellContext>();
  const loaded = page != null;
  const projects = page?.projects ?? [];

  return (
    <StatusSection title="Projects" icon={<BrandIcon icon={siGithub} />} columns="grid-cols-1">
      {!loaded ? (
        <LoadingCard />
      ) : projects.length === 0 ? (
        <EmptyCard text="No projects yet" />
      ) : (
        <ProjectTimeline projects={projects} />
      )}
    </StatusSection>
  );
}
