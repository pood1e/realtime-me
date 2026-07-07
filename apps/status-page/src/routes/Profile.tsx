import { useCallback } from 'react';
import { siGithub } from 'simple-icons/icons';
import { BrandIcon } from '@/components/brand';
import { EmptyCard, HeaderActions, LoadingCard, NavLinks, PageFooter, PageFrame, SiteLogo, StatusSection } from '@/components/layout';
import { ProfileIntro, ProjectCard } from '@/components/ProfileCards';
import { usePolling } from '@/hooks/usePolling';
import { profileClient } from '@/lib/transport';

export function ProfileApp() {
  const fetchProfile = useCallback(async (signal: AbortSignal) => {
    return (await profileClient.getProfilePage({}, { signal })).page;
  }, []);
  const { data: page, error, refresh } = usePolling(fetchProfile, { intervalMs: 0 });
  const failed = error !== null;
  const loaded = page != null;
  const projects = page?.projects ?? [];

  return (
    <PageFrame>
      <header className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <SiteLogo />
          <NavLinks />
        </div>
        <HeaderActions failed={failed} refresh={refresh} />
      </header>

      <ProfileIntro profile={page?.profile} loaded={loaded} />

      <StatusSection title="Projects" icon={<BrandIcon icon={siGithub} />} columns="md:grid-cols-2 xl:grid-cols-3">
        {!loaded ? (
          <LoadingCard />
        ) : projects.length === 0 ? (
          <EmptyCard text="No projects yet" />
        ) : (
          projects.map((project) => <ProjectCard key={project.uid} project={project} />)
        )}
      </StatusSection>

      <PageFooter updatedAt={page?.updateTime} />
    </PageFrame>
  );
}
