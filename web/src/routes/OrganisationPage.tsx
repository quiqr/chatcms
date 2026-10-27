import type { ReactNode } from "react";

import { useHere } from "../lib/whoami";
import { OrganisationShell } from "../shells/OrganisationShell";
import { Placeholder } from "./Placeholder";
import { say, type Key } from "../lib/say";

/**
 * The title comes as a key rather than as a sentence, so that it is said
 * when the screen is drawn rather than when the route is declared. A route
 * is declared once; the language can change after that.
 */
export function OrganisationPage({ title, body }: { title: Key; body?: ReactNode }) {
  const said = say(title);
  const { who, organisation } = useHere();
  const named = who.organisations.find((one) => one.slug === organisation);

  return (
    <OrganisationShell
      organisation={{ name: named?.name ?? organisation ?? "Your work" }}
      organisations={who.organisations}
      viewer={{ name: who.person.displayName || who.person.email }}
    >
      <h1 className="mb-6 text-[28px] font-semibold tracking-tight">{said}</h1>
      {body ?? <Placeholder screen={said} />}
    </OrganisationShell>
  );
}
