import type { ReactNode } from "react";
import { useParams } from "react-router";

import { OrganisationAdminShell } from "../shells/OrganisationAdminShell";
import { readOrganisationArea, type OrganisationArea } from "../lib/tenancy";
import { say } from "../lib/say";
import { useAsync } from "../lib/useAsync";
import { useHere } from "../lib/whoami";

/**
 * ManageOrganisationPage loads the organisation once and gives it to whichever
 * screen of the area is open, because the shell needs the counts either way.
 */
export function ManageOrganisationPage({
  screen,
}: {
  screen: (area: OrganisationArea, again: () => void) => ReactNode;
}) {
  const { organisation = "" } = useParams();
  const { who } = useHere();
  const [area, again] = useAsync(() => readOrganisationArea(organisation), [organisation]);

  if (area.state === "loading") {
    return <p className="p-6 text-[15px] text-muted-foreground">{say("common.looking")}</p>;
  }
  if (area.state === "failed") {
    return (
      <p role="alert" className="m-6 rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
        {area.message}
      </p>
    );
  }

  return (
    <OrganisationAdminShell
      organisation={organisation}
      name={area.value.name}
      usage={area.value.usage}
      viewer={{ name: who.person.displayName || who.person.email }}
    >
      {screen(area.value, again)}
    </OrganisationAdminShell>
  );
}
