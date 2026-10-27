import { ChevronsUpDown, MessageSquare } from "lucide-react";
import type { ReactNode } from "react";
import { NavLink } from "react-router";

import { Avatar, Tile } from "../components/ui/avatar";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { cn } from "../lib/cn";
import { useHere } from "../lib/whoami";
import { setupNavigation, siteNavigation, type NavEntry } from "./navigation";
import { canSeeSetup } from "./roles";
import { SyncIndicator, type SyncState } from "./SyncIndicator";
import type { DemoProject, Role } from "../demo/content";

export type ProjectShellProps = {
  project: DemoProject;
  viewer: { name: string; role: Role };
  title: ReactNode;
  actions?: ReactNode;
  syncState?: SyncState;
  children: ReactNode;
};

function NavGroup({
  label,
  entries,
  unpublished,
}: {
  label: string;
  entries: NavEntry[];
  unpublished?: number;
}) {
  return (
    <div role="group" aria-label={label} className="flex flex-col gap-0.5">
      {entries.map((entry) => (
        <NavLink
          key={entry.key}
          to={entry.to}
          className={({ isActive }) =>
            cn(
              "flex h-11 items-center gap-2.5 rounded-lg px-3 text-[15px] text-muted-foreground",
              isActive && "bg-muted font-medium text-foreground",
            )
          }
        >
          <entry.icon aria-hidden="true" className="size-[18px]" />
          <span className="flex-1">{entry.label}</span>
          {entry.key === "publish" && unpublished !== undefined && unpublished > 0 && (
            <Badge tone="changed">{unpublished}</Badge>
          )}
          {entry.soon && <Badge tone="neutral">Soon</Badge>}
        </NavLink>
      ))}
    </div>
  );
}

export function ProjectShell({
  project,
  viewer,
  title,
  actions,
  syncState = "saved",
  children,
}: ProjectShellProps) {
  const { organisation } = useHere();
  return (
    <div data-testid="project-shell" className="hidden min-h-dvh flex-wrap sm:flex">
      <aside
        aria-label="Project navigation"
        className="flex shrink basis-60 grow flex-col gap-4 border-r border-border bg-card p-3"
      >
        <span className="flex items-center gap-2 px-2 pt-1">
          <img src="/quiqr-logo.svg" alt="Quiqr" className="h-6 w-auto" />
          <span className="text-[15px] text-muted-foreground">ChatCMS</span>
        </span>

        <NavLink
          to="/projects"
          aria-label={`${project.name}, switch project`}
          className="flex min-h-14 items-center gap-2.5 rounded-xl border border-border bg-card px-2.5 py-2 hover:bg-muted"
        >
          <Tile name={project.name} />
          <span className="flex min-w-0 flex-1 flex-col">
            <span className="truncate text-[15px] font-semibold">{project.name}</span>
            <span className="truncate text-xs text-subtle-foreground">{project.domain}</span>
          </span>
          <ChevronsUpDown aria-hidden="true" className="size-4 shrink-0 text-subtle-foreground" />
        </NavLink>

        <NavGroup label="Site" entries={siteNavigation(organisation, project.slug)} unpublished={project.unpublished} />
        {canSeeSetup(viewer.role) && (
          <NavGroup label="Setup" entries={setupNavigation(organisation, project.slug)} />
        )}

        <div className="mt-auto flex items-center gap-2 border-t border-border-subtle px-2 pt-3">
          <Avatar name={viewer.name} />
          <span className="flex min-w-0 flex-col">
            <span className="truncate text-sm">{viewer.name}</span>
            <span className="text-xs capitalize text-subtle-foreground">{viewer.role}</span>
          </span>
        </div>
      </aside>

      <div className="flex min-w-0 shrink grow-[999] basis-[560px] flex-col">
        <header className="flex min-h-14 flex-wrap items-center gap-3 border-b border-border bg-card px-8 py-3">
          <h1 className="flex-1 text-lg font-semibold">{title}</h1>
          <SyncIndicator state={syncState} />
          {actions}
          <Button asChild count={project.unpublished}>
            <NavLink to={`/o/${organisation}/p/${project.slug}/publish`}>Publish</NavLink>
          </Button>
        </header>

        <main className="min-w-0 flex-1 p-8">{children}</main>
      </div>
    </div>
  );
}

export function AssistantSlot() {
  return (
    <Button variant="ghost" size="icon" aria-label="Assistant, coming later" disabled>
      <MessageSquare aria-hidden="true" className="size-5" />
    </Button>
  );
}
