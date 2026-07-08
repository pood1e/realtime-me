import { siGithub } from 'simple-icons/icons';
import { useOutletContext } from 'react-router-dom';
import type { ShellContext } from '@/components/AppShell';
import { BrandIcon } from '@/components/brand';
import { EmptyCard, LoadingCard, StatusSection } from '@/components/layout';
import { ProfileIntro, ProjectCard } from '@/components/ProfileCards';

export function ProfileApp() {
  const { page } = useOutletContext<ShellContext>();
  const loaded = page != null;
  const projects = page?.projects ?? [];

  return (
    <div className="grid gap-8">
      <ProfileIntro profile={page?.profile} loaded={loaded} />

      <StatusSection title="Projects" icon={<BrandIcon icon={siGithub} />} columns="md:grid-cols-2 xl:grid-cols-3">
        {!loaded ? (
          <LoadingCard />
        ) : projects.length === 0 ? (
          <EmptyCard text="No projects yet" />
        ) : (
          projects.map((project) => <ProjectCard key={project.uid || project.displayName} project={project} />)
        )}
      </StatusSection>
    </div>
  );
}
