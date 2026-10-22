import { ChevronDown } from "lucide-react";
import type { ReactNode } from "react";
import { NavLink } from "react-router";

import { Avatar } from "../components/ui/avatar";
import { Menu, MenuContent, MenuItem, MenuSeparator, MenuTrigger } from "../components/ui/menu";

export type OrganisationShellProps = {
  organisation: { name: string };
  viewer: { name: string };
  children: ReactNode;
};

export function OrganisationShell({ organisation, viewer, children }: OrganisationShellProps) {
  return (
    <div data-testid="organisation-shell" className="min-h-dvh">
      <header className="flex min-h-14 flex-wrap items-center gap-3 border-b border-border bg-card px-6 py-3">
        <NavLink to="/projects" className="flex min-h-11 items-center" aria-label="Quiqr ChatCMS, go to your projects">
          <img src="/quiqr-logo.svg" alt="Quiqr ChatCMS" className="h-7 w-auto" />
        </NavLink>

        <Menu>
          <MenuTrigger className="flex h-11 items-center gap-1 rounded-md px-3 text-sm font-medium hover:bg-muted">
            {organisation.name}
            <ChevronDown aria-hidden="true" className="size-4" />
          </MenuTrigger>
          <MenuContent>
            <MenuItem>{organisation.name}</MenuItem>
            <MenuSeparator />
            <MenuItem>New organisation</MenuItem>
          </MenuContent>
        </Menu>

        <span className="flex-1" />

        <NavLink to="/help" className="flex h-11 items-center px-3 text-sm text-muted-foreground">
          Help
        </NavLink>
        <Avatar name={viewer.name} />
      </header>

      <main className="mx-auto w-full max-w-[1200px] p-6">{children}</main>
    </div>
  );
}
