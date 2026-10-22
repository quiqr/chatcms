import { useParams } from "react-router";

import { demoProjects, demoViewer } from "../demo/content";
import { PhoneShell } from "../shells/PhoneShell";
import { ProjectShell } from "../shells/ProjectShell";
import { Placeholder } from "./Placeholder";

export function ProjectPage({ title }: { title: string }) {
  const { project: slug } = useParams();
  const project = demoProjects.find((candidate) => candidate.slug === slug) ?? demoProjects[0]!;

  return (
    <>
      <ProjectShell project={project} viewer={demoViewer} title={title}>
        <Placeholder screen={title} />
      </ProjectShell>
      <PhoneShell project={project} viewer={demoViewer}>
        <Placeholder screen={title} />
      </PhoneShell>
    </>
  );
}
