import { useParams } from "react-router";

import { useAsync } from "../lib/useAsync";
import { readProject } from "../lib/tenancy";
import { useHere } from "../lib/whoami";
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
  const { organisation, who } = useHere();
  const { project: slug = "" } = useParams();
  const onAPhone = useIsPhone();

  // The project comes from the server. It used to come from the demo
  // scaffold, which is how the count on the Publish button came to disagree
  // with the publish screen beside it.
  const [loaded] = useAsync(() => readProject(organisation, slug), [organisation, slug]);

  if (loaded.state === "loading") {
    return <p className="p-6 text-[15px] text-muted-foreground">Looking...</p>;
  }

  if (loaded.state === "failed") {
    return (
      <p role="alert" className="m-6 rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
        {loaded.message}
      </p>
    );
  }

  const project = loaded.value;
  const viewer = { name: who.person.displayName || who.person.email, role: project.yourRole };
  const content = body ?? <Placeholder screen={title} />;

  if (onAPhone) {
    return (
      <PhoneShell project={project} viewer={viewer}>
        {content}
      </PhoneShell>
    );
  }

  return (
    <ProjectShell project={project} viewer={viewer} title={title}>
      {content}
    </ProjectShell>
  );
}
