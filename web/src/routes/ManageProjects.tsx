import { Ban, MoreHorizontal, Plus } from "lucide-react";
import { useState } from "react";
import { NavLink } from "react-router";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Menu, MenuContent, MenuItem, MenuSeparator, MenuTrigger } from "../components/ui/menu";
import { Tile } from "../components/ui/avatar";
import { whyThereIsNoRoom } from "../components/ui/usage";
import { plural, say } from "../lib/say";
import { projectPath } from "../lib/paths";
import {
  archiveProject,
  duplicateProject,
  moveProject,
  zipPath,
  type OrganisationArea,
  type Project,
} from "../lib/tenancy";
import { whatWentWrong } from "../lib/api";
import { useHere } from "../lib/whoami";
import { Duplicating, Moving } from "./WholeSiteDialogs";

/**
 * The organisation's sites, as a list rather than a table so that it reflows
 * on a phone without a second layout.
 *
 * The menu holds four acts that are not built yet. They are drawn disabled
 * and say so: leaving them out would mean building this menu twice, and
 * drawing them live would be a lie.
 */
export function ManageProjects({ area, again }: { area: OrganisationArea; again: () => void }) {
  const { who } = useHere();
  const noRoom = whyThereIsNoRoom(area.usage.projects, "sites");

  const [doing, setDoing] = useState<{ what: "copy" | "move"; project: Project } | null>(null);
  const [busy, setBusy] = useState(false);
  const [trouble, setTrouble] = useState("");

  async function run(what: () => Promise<unknown>) {
    setBusy(true);
    setTrouble("");
    try {
      await what();
      setDoing(null);
      again();
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error));
    } finally {
      setBusy(false);
    }
  }

  const live = area.projects.filter((project) => !project.archived);
  const archived = area.projects.filter((project) => project.archived);

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

      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      {doing?.what === "copy" && (
        <Duplicating
          area={area}
          project={doing.project}
          busy={busy}
          onYes={(body) => void run(() => duplicateProject(area.slug, doing.project.slug, body))}
          onCancel={() => setDoing(null)}
        />
      )}

      {doing?.what === "move" && (
        <Moving
          who={who}
          from={area.slug}
          project={doing.project}
          busy={busy}
          onYes={(to) => void run(() => moveProject(area.slug, doing.project.slug, to))}
          onCancel={() => setDoing(null)}
        />
      )}

      {live.length === 0 ? (
        <p className="text-[15px] text-subtle-foreground">{say("manage.no-sites-in-it-yet")}</p>
      ) : (
        <ul className="flex flex-col gap-3">
          {live.map((project) => (
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
                  <MenuItem onSelect={() => setDoing({ what: "copy", project })}>
                    {say("manage.duplicate")}
                  </MenuItem>
                  <MenuItem onSelect={() => setDoing({ what: "move", project })}>
                    {say("manage.move-to-another-organisation")}
                  </MenuItem>
                  <MenuItem asChild>
                    <a href={zipPath(area.slug, project.slug)} download>
                      {say("manage.download-as-zip")}
                    </a>
                  </MenuItem>
                  <MenuItem
                    onSelect={() => void run(() => archiveProject(area.slug, project.slug, true))}
                  >
                    {say("manage.archive")}
                  </MenuItem>
                </MenuContent>
              </Menu>
            </li>
          ))}
        </ul>
      )}

      {archived.length > 0 && (
        <section className="flex flex-col gap-2">
          <h2 className="text-[15px] font-semibold">
            {say("archive.archived")} · {archived.length}{" "}
            <span className="font-normal text-subtle-foreground">
              · {say("archive.does-not-count")}
            </span>
          </h2>

          <ul className="flex flex-col gap-2">
            {archived.map((project) => (
              <li
                key={project.slug}
                className="flex flex-wrap items-center gap-3 rounded-xl border border-border-subtle p-3"
              >
                <span className="flex-1 text-[15px] text-muted-foreground">{project.name}</span>
                <Button
                  variant="outline"
                  disabled={busy}
                  onClick={() => void run(() => archiveProject(area.slug, project.slug, false))}
                >
                  {say("archive.bring-it-back")}
                </Button>
              </li>
            ))}
          </ul>
        </section>
      )}

      <p className="text-[13px] text-subtle-foreground">{say("manage.deleting-says-what-stays")}</p>
    </div>
  );
}
