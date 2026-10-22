import { demoOrganisation, demoViewer } from "../demo/content";
import { OrganisationShell } from "../shells/OrganisationShell";
import { Placeholder } from "./Placeholder";

export function OrganisationPage({ title }: { title: string }) {
  return (
    <OrganisationShell organisation={demoOrganisation} viewer={demoViewer}>
      <Placeholder screen={title} />
    </OrganisationShell>
  );
}
