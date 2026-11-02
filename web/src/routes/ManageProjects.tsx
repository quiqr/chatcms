import { Ban, MoreHorizontal, Plus } from "lucide-react";
import { NavLink } from "react-router";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Menu, MenuContent, MenuItem, MenuSeparator, MenuTrigger } from "../components/ui/menu";
import { Tile } from "../components/ui/avatar";
import { whyThereIsNoRoom } from "../components/ui/usage";
import { plural, say } from "../lib/say";
import { projectPath } from "../lib/paths";
import type { OrganisationArea } from "../lib/tenancy";

/**
 * The organisation's sites, as a list rather than a table so that it reflows
 * on a phone without a second layout.
 *
 * The menu holds four acts that are not built yet. They are drawn disabled
 * and say so: leaving them out would mean building this menu twice, and
 * drawing them live would be a lie.
 */
export function ManageProjects({ area }: { area: OrganisationArea }) {
  const noRoom = whyThereIsNoRoom(area.usage.projects, "sites");

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center gap-3">
        <h1 className="flex-1 text-[28px] font-semibold tracking-tight">{say("manage.projects")}</h1>

        <Button variant="ink" asChild disabled={noRoom !== null}>
          <NavLink to={`/o/${area.slug}/projects`}>
            <Plus aria-hidden="true" className="size-4" />
            {say("projects.new-project")}
          </NavLink>
        </Button>
      </div>

      {noRoom && (
        <p className="flex items-start gap-2 rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          <Ban aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
          {noRoom}
        </p>
      )}

      {area.projects.length === 0 ? (
        <p className="text-[15px] text-subtle-foreground">{say("manage.no-sites-in-it-yet")}</p>
      ) : (
        <ul className="flex flex-col gap-3">
          {area.projects.map((project) => (
            <li
              key={project.slug}
              className="flex flex-wrap items-center gap-3 rounded-xl border border-border bg-card p-4"
            >
              <Tile name={project.name} />

              <span className="flex min-w-0 flex-1 flex-col">
                <span className="truncate text-[15px] font-semibold">{project.name}</span>
                <span className="truncate text-[13px] text-subtle-foreground">
                  {project.domain || area.slug}
                </span>
              </span>

              {project.unpublished > 0 ? (
                <Badge tone="changed">
                  {plural(project.unpublished, "projects.unpublished-one", "projects.unpublished-many")}
                </Badge>
              ) : (
                <Badge tone="published">{say("projects.published")}</Badge>
              )}

              <Button variant="outline" asChild>
                <NavLink to={projectPath(area.slug, project.slug)}>{say("manage.open")}</NavLink>
              </Button>

              <Menu>
                <MenuTrigger
                  aria-label={say("manage.what-can-be-done-with-it", { name: project.name })}
                  className="flex size-11 items-center justify-center rounded-md hover:bg-muted"
                >
                  <MoreHorizontal aria-hidden="true" className="size-4" />
                </MenuTrigger>
                <MenuContent>
                  <MenuItem asChild>
                    <NavLink to={projectPath(area.slug, project.slug, "people")}>
                      {say("manage.people-and-roles")}
                    </NavLink>
                  </MenuItem>
                  <MenuItem asChild>
                    <NavLink to={projectPath(area.slug, project.slug, "settings")}>
                      {say("manage.project-settings")}
                    </NavLink>
                  </MenuItem>
                  <MenuSeparator />
                  {(
                    [
                      ["manage.duplicate", "duplicate"],
                      ["manage.move-to-another-organisation", "move"],
                      ["manage.download-as-zip", "zip"],
                      ["manage.archive", "archive"],
                    ] as const
                  ).map(([key, what]) => (
                    <MenuItem key={what} disabled>
                      {say(key)}
                      <span className="ml-auto text-xs text-subtle-foreground">
                        {say("manage.not-built-yet")}
                      </span>
                    </MenuItem>
                  ))}
                </MenuContent>
              </Menu>
            </li>
          ))}
        </ul>
      )}

      <p className="text-[13px] text-subtle-foreground">{say("manage.deleting-says-what-stays")}</p>
    </div>
  );
}
