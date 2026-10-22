import { useParams } from "react-router";

import { demoProjects, demoViewer } from "../demo/content";
import { PhoneShell } from "../shells/PhoneShell";
import { ProjectShell } from "../shells/ProjectShell";
import { useIsPhone } from "../shells/useIsPhone";
import { Placeholder } from "./Placeholder";

import type { ReactNode } from "react";

export function ProjectPage({
  title,
  body,
}: {
  title: string;
  body?: ReactNode;
}) {
  const { project: slug } = useParams();
  const project =
    demoProjects.find((candidate) => candidate.slug === slug) ??
    demoProjects[0]!;
  const onAPhone = useIsPhone();

  const content = body ?? <Placeholder screen={title} />;

  if (onAPhone) {
    return (
      <PhoneShell project={project} viewer={demoViewer}>
        {content}
      </PhoneShell>
    );
  }

  return (
    <ProjectShell project={project} viewer={demoViewer} title={title}>
      {content}
    </ProjectShell>
  );
}
