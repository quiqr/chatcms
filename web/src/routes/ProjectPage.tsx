import { useParams } from "react-router";

import { useAsync } from "../lib/useAsync";
import { readProject } from "../lib/tenancy";
import { useHere } from "../lib/whoami";
import { PhoneShell } from "../shells/PhoneShell";
import { ProjectShell } from "../shells/ProjectShell";
import { useIsPhone } from "../shells/useIsPhone";
import { Placeholder } from "./Placeholder";

import type { ReactNode } from "react";
import { say, type Key } from "../lib/say";

export function ProjectPage({
  title,
  body,
}: {
  // A key rather than a sentence: a route is declared once, and the
  // language can change after that.
  title: Key;
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
    return <p className="p-6 text-[15px] text-muted-foreground">{say("common.looking")}</p>;
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
  const said = say(title);
  const content = body ?? <Placeholder screen={said} />;

  if (onAPhone) {
    return (
      <PhoneShell project={project} viewer={viewer}>
        {content}
      </PhoneShell>
    );
  }

  return (
    <ProjectShell project={project} viewer={viewer} title={said}>
      {content}
    </ProjectShell>
  );
}
