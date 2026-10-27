import type { ReactNode } from "react";

import { useHere } from "../lib/whoami";
import { OrganisationShell } from "../shells/OrganisationShell";
import { Placeholder } from "./Placeholder";

export function OrganisationPage({ title, body }: { title: string; body?: ReactNode }) {
  const { who, organisation } = useHere();
  const named = who.organisations.find((one) => one.slug === organisation);

  return (
    <OrganisationShell
      organisation={{ name: named?.name ?? organisation ?? "Your work" }}
      organisations={who.organisations}
      viewer={{ name: who.person.displayName || who.person.email }}
    >
      <h1 className="mb-6 text-[28px] font-semibold tracking-tight">{title}</h1>
      {body ?? <Placeholder screen={title} />}
    </OrganisationShell>
  );
}
