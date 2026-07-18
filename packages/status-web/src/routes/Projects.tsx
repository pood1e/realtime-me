import type { ShellContext } from "@realtime-me/status-web/components/AppShell";
import { BrandIcon } from "@realtime-me/status-web/components/brand";
import {
  EmptyCard,
  ErrorCard,
  LoadingCard,
  StatusSection,
} from "@realtime-me/status-web/components/layout";
import { ProjectTimeline } from "@realtime-me/status-web/components/ProjectCards";
import { usePolling } from "@realtime-me/status-web/hooks/usePolling";
import { useCallback } from "react";
import { useOutletContext } from "react-router-dom";
import { siGithub } from "simple-icons/icons";

export function ProjectsApp() {
  const { api } = useOutletContext<ShellContext>();
  const fetchProjects = useCallback(
    async (signal: AbortSignal) => (await api.projects.listProjects({}, { signal })).projects,
    [api],
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
