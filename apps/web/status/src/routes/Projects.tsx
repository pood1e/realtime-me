import { useCallback } from "react";
import { siGithub } from "simple-icons/icons";
import { BrandIcon } from "@/components/brand";
import { EmptyCard, ErrorCard, LoadingCard, StatusSection } from "@/components/layout";
import { ProjectTimeline } from "@/components/ProjectCards";
import { usePolling } from "@/hooks/usePolling";
import { projectsClient } from "@/lib/transport";

export function ProjectsApp() {
  const fetchProjects = useCallback(
    async (signal: AbortSignal) => (await projectsClient.listProjects({}, { signal })).projects,
    [],
  );
  const { data: projects, error, refresh } = usePolling(fetchProjects, { intervalMs: 0 });

  return (
    <StatusSection title="Projects" icon={<BrandIcon icon={siGithub} />} columns="grid-cols-1">
      {body()}
    </StatusSection>
  );

  // The projects are fetched once, so a single failed request would otherwise spin
  // forever. Say so, and offer the retry the visitor would reach for anyway.
  function body() {
    if (error !== null && projects == null)
      return <ErrorCard text="Cannot load the projects right now." retry={refresh} />;
    if (projects == null) return <LoadingCard />;
    if (projects.length === 0) return <EmptyCard text="No projects yet" />;
    return <ProjectTimeline projects={projects} />;
  }
}
