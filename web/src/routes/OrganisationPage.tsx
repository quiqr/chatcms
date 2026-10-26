import type { ReactNode } from "react";

import { demoOrganisation, demoViewer } from "../demo/content";
import { OrganisationShell } from "../shells/OrganisationShell";
import { Placeholder } from "./Placeholder";

export function OrganisationPage({ title, body }: { title: string; body?: ReactNode }) {
  return (
    <OrganisationShell organisation={demoOrganisation} viewer={demoViewer}>
      <h1 className="mb-6 text-[28px] font-semibold tracking-tight">{title}</h1>
      {body ?? <Placeholder screen={title} />}
    </OrganisationShell>
  );
}
