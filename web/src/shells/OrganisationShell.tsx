import { ChevronDown } from "lucide-react";
import type { ReactNode } from "react";
import { NavLink } from "react-router";

import { PersonMenu } from "./PersonMenu";
import { Menu, MenuContent, MenuItem, MenuSeparator, MenuTrigger } from "../components/ui/menu";
import { useAdministers } from "./useAdministers";
import { say } from "../lib/say";

export type OrganisationShellProps = {
  organisation: { name: string };
  organisations?: { slug: string; name: string }[];
  viewer?: { name: string };
  children: ReactNode;
};

export function OrganisationShell({
  organisation,
  organisations = [],
  viewer,
  children,
}: OrganisationShellProps) {
  const administers = useAdministers();

  return (
    <div data-testid="organisation-shell" className="min-h-dvh">
      <header className="flex min-h-14 flex-wrap items-center gap-3 border-b border-border bg-card px-6 py-3">
        <NavLink to="/projects" className="flex min-h-11 items-center" aria-label={say("organisation-shell.quiqr-chatcms-go-to-your-projects")}>
          <img src="/quiqr-logo.svg" alt={say("organisation-shell.quiqr-chatcms")} className="h-7 w-auto" />
        </NavLink>

        <Menu>
          <MenuTrigger className="flex h-11 items-center gap-1 rounded-md px-3 text-sm font-medium hover:bg-muted">
            {organisation.name}
            <ChevronDown aria-hidden="true" className="size-4" />
          </MenuTrigger>
          <MenuContent>
            {organisations.length === 0 ? (
              <MenuItem>{organisation.name}</MenuItem>
            ) : (
              organisations.map((one) => <MenuItem key={one.slug}>{one.name}</MenuItem>)
            )}
            {administers && (
              <>
                <MenuSeparator />
                <MenuItem asChild>
                  <NavLink to="/administration">{say("organisation-shell.start-an-organisation")}</NavLink>
                </MenuItem>
              </>
            )}
          </MenuContent>
        </Menu>

        <span className="flex-1" />

        {administers && (
          <NavLink
            to="/administration"
            className="flex h-11 items-center px-3 text-sm text-muted-foreground"
          >
            {say("organisation-shell.administration")}
          </NavLink>
        )}
        <NavLink to="/help" className="flex h-11 items-center px-3 text-sm text-muted-foreground">
          {say("common.help")}
        </NavLink>
        {viewer && <PersonMenu name={viewer.name} />}
      </header>

      <main className="mx-auto w-full max-w-[1200px] p-6">{children}</main>
    </div>
  );
}
