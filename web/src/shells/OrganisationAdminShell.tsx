import { FolderGit2, Users } from "lucide-react";
import type { ReactNode } from "react";
import { NavLink } from "react-router";

import { PersonMenu } from "./PersonMenu";
import { Tile } from "../components/ui/avatar";
import { HowFull } from "../components/ui/usage";
import { say } from "../lib/say";
import type { Usage } from "../components/ui/usage";

/**
 * The shell of an organisation's own area. It is the project shell's sibling:
 * a light sidebar, the thing it is about at the top, and what that thing
 * holds down the side. The usage box says how full the organisation is and
 * who sets that, so that somebody who cannot change a limit knows who can.
 */
export function OrganisationAdminShell({
  organisation,
  name,
  usage,
  viewer,
  children,
}: {
  organisation: string;
  name: string;
  usage: Usage;
  viewer?: { name: string };
  children: ReactNode;
}) {
  const here = `/o/${organisation}/manage`;

  return (
    <div className="min-h-dvh sm:flex">
      <aside
        aria-label={say("manage.manage-organisation")}
        className="flex flex-col gap-4 border-b border-border bg-card p-4 sm:min-h-dvh sm:w-[260px] sm:shrink-0 sm:border-b-0 sm:border-r"
      >
        <NavLink to="/projects" className="flex min-h-11 items-center gap-2">
          <img src="/quiqr-logo.svg" alt={say("organisation-shell.quiqr-chatcms")} className="h-6 w-auto" />
          <span className="text-sm font-semibold">{say("common.chatcms")}</span>
        </NavLink>

        <div className="flex items-center gap-2">
          <Tile name={name} />
          <span className="flex min-w-0 flex-col">
            <span className="truncate text-[15px] font-semibold">{name}</span>
            <span className="text-[13px] text-subtle-foreground">{say("manage.organisation")}</span>
          </span>
        </div>

        <nav className="flex flex-col gap-1">
          <p className="px-2 text-xs font-medium text-subtle-foreground">
            {say("manage.manage-organisation")}
          </p>
          <NavLink
            to={`${here}/projects`}
            className={({ isActive }) =>
              `flex min-h-11 items-center gap-2 rounded-md px-2 text-[15px] ${
                isActive ? "bg-muted font-medium" : "hover:bg-muted"
              }`
            }
          >
            <FolderGit2 aria-hidden="true" className="size-4" />
            {say("manage.projects")}
            <span className="ml-auto text-[13px] text-subtle-foreground">
              {usage.projects.held}
            </span>
          </NavLink>
          <NavLink
            to={`${here}/members`}
            className={({ isActive }) =>
              `flex min-h-11 items-center gap-2 rounded-md px-2 text-[15px] ${
                isActive ? "bg-muted font-medium" : "hover:bg-muted"
              }`
            }
          >
            <Users aria-hidden="true" className="size-4" />
            {say("manage.members")}
            <span className="ml-auto text-[13px] text-subtle-foreground">{usage.members.held}</span>
          </NavLink>
        </nav>

        <div className="flex flex-col gap-3 rounded-lg bg-background p-3">
          <HowFull counted={usage.members} of="seats" />
          <HowFull counted={usage.projects} of="sites" />
          <p className="text-[13px] text-subtle-foreground">
            {say("manage.limits-are-set-by-the-platform-admin")}
          </p>
        </div>

        <span className="flex-1" />

        <NavLink
          to={`/o/${organisation}/projects`}
          className="flex min-h-11 items-center px-2 text-sm text-muted-foreground hover:text-foreground"
        >
          {say("manage.back-to-my-projects")}
        </NavLink>

        {viewer && <PersonMenu name={viewer.name} />}
      </aside>

      <main className="w-full max-w-[1000px] p-6">{children}</main>
    </div>
  );
}
