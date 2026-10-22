import { Plus, Search } from "lucide-react";
import { useMemo, useState } from "react";
import { NavLink, useParams } from "react-router";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Tile } from "../components/ui/avatar";
import { useAsync } from "../lib/useAsync";
import { listOrganisations, listProjects, type Project } from "../lib/tenancy";
import { OrganisationShell } from "../shells/OrganisationShell";

function status(project: Project) {
  if (project.unpublished > 0) {
    return <Badge tone="changed">{project.unpublished} unpublished changes</Badge>;
  }
  return <Badge tone="published">Published</Badge>;
}

function ProjectCard({ organisation, project }: { organisation: string; project: Project }) {
  return (
    <NavLink
      to={`/p/${project.slug}/content`}
      className="flex min-h-11 flex-col gap-3 rounded-xl border border-border bg-card p-4 hover:border-input"
    >
      <span className="flex items-center gap-2.5">
        <Tile name={project.name} />
        <span className="flex min-w-0 flex-1 flex-col">
          <span className="truncate text-[15px] font-semibold">{project.name}</span>
          <span className="truncate text-[13px] text-subtle-foreground">
            {project.domain || organisation} · Hugo
          </span>
        </span>
      </span>

      <span>{status(project)}</span>
    </NavLink>
  );
}

export function ProjectsPage() {
  const { organisation: fromAddress } = useParams();
  const [search, setSearch] = useState("");

  const [organisations] = useAsync(listOrganisations, []);
  const current =
    fromAddress ??
    (organisations.state === "ready" ? organisations.value.organisations[0]?.slug : undefined);

  const [projects] = useAsync(
    () => (current ? listProjects(current) : Promise.resolve({ projects: [] })),
    [current],
  );

  const shown = useMemo(() => {
    if (projects.state !== "ready") return [];
    const needle = search.trim().toLowerCase();
    if (!needle) return projects.value.projects;
    return projects.value.projects.filter(
      (project) =>
        project.name.toLowerCase().includes(needle) ||
        project.domain.toLowerCase().includes(needle),
    );
  }, [projects, search]);

  const organisationName =
    organisations.state === "ready"
      ? (organisations.value.organisations.find((one) => one.slug === current)?.name ??
        current ??
        "Your work")
      : "Your work";

  return (
    <OrganisationShell
      organisation={{ name: organisationName }}
      organisations={organisations.state === "ready" ? organisations.value.organisations : []}
    >
      <div className="flex flex-wrap items-center gap-3">
        <h1 className="flex-1 text-[28px] font-semibold tracking-tight">Projects</h1>

        <label className="flex h-11 min-w-[220px] flex-1 items-center gap-2 rounded-md border border-input bg-card px-3 sm:max-w-[320px] sm:flex-none">
          <Search aria-hidden="true" className="size-4 text-subtle-foreground" />
          <span className="sr-only">Search projects</span>
          <input
            // chatcms-target: the label around this one carries the hit area
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search projects"
            className="h-full min-w-0 flex-1 bg-transparent text-[15px] outline-hidden"
          />
        </label>

        <Button variant="ink">
          <Plus aria-hidden="true" className="size-4" />
          New project
        </Button>
      </div>

      {projects.state === "loading" && (
        <p className="mt-6 text-[15px] text-muted-foreground">Looking for your projects...</p>
      )}

      {projects.state === "failed" && (
        <p role="alert" className="mt-6 rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {projects.message}
        </p>
      )}

      {projects.state === "ready" && (
        <div className="mt-6 grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(min(340px,100%),1fr))]">
          {shown.map((project) => (
            <ProjectCard key={project.slug} organisation={current ?? ""} project={project} />
          ))}

          <div className="flex min-h-[132px] items-center justify-center rounded-xl border border-dashed border-input p-4 text-center text-[15px] text-subtle-foreground">
            Import a site or start from a template
          </div>
        </div>
      )}
    </OrganisationShell>
  );
}
