import type { ReactNode } from "react";
import { NavLink } from "react-router";

import { Avatar, Tile } from "../components/ui/avatar";
import { cn } from "../lib/cn";
import { siteNavigation } from "./navigation";
import { SyncIndicator } from "./SyncIndicator";
import type { DemoProject, Role } from "../demo/content";

export type PhoneShellProps = {
  project: DemoProject;
  viewer: { name: string; role: Role };
  chips?: ReactNode;
  bottomBar?: ReactNode;
  children: ReactNode;
};

export function PhoneShell({
  project,
  viewer,
  chips,
  bottomBar,
  children,
}: PhoneShellProps) {
  const tabs = siteNavigation(project.slug);

  return (
    <div
      data-testid="phone-shell"
      className="flex min-h-dvh flex-col overflow-x-clip sm:hidden"
    >
      <header className="flex items-center gap-2 border-b border-border bg-card px-4 py-3">
        <Tile name={project.name} />
        <span className="flex min-w-0 flex-1 flex-col">
          <span className="truncate text-sm font-semibold">{project.name}</span>
          <SyncIndicator />
        </span>
        <Avatar name={viewer.name} />
      </header>

      {chips && (
        <div className="flex gap-2 overflow-x-auto border-b border-border bg-card px-4 py-2">
          {chips}
        </div>
      )}

      <main className="min-w-0 flex-1 overflow-x-clip p-4 pb-28">{children}</main>

      {bottomBar ? (
        <div
          data-testid="phone-action-bar"
          className="pb-safe-2 fixed inset-x-0 bottom-0 flex items-center gap-2 border-t border-border bg-card px-4 pt-2"
        >
          {bottomBar}
        </div>
      ) : (
        <nav
          aria-label="Site sections"
          data-testid="phone-tab-bar"
          className="pb-safe fixed inset-x-0 bottom-0 flex border-t border-border bg-card"
        >
          {tabs.map((tab) => (
            <NavLink
              key={tab.key}
              to={tab.to}
              className={({ isActive }) =>
                cn(
                  "flex h-[52px] flex-1 flex-col items-center justify-center gap-0.5 text-[11px]",
                  isActive ? "text-primary" : "text-muted-foreground",
                )
              }
            >
              <span className="relative">
                <tab.icon aria-hidden="true" className="size-5" />
                {tab.key === "publish" && project.unpublished > 0 && (
                  <span className="absolute -right-2.5 -top-1 rounded-full bg-warning-soft px-1.5 text-[10px] font-semibold leading-4 text-warning-foreground">
                    {project.unpublished}
                  </span>
                )}
              </span>
              {tab.label}
              {tab.key === "publish" && project.unpublished > 0 && (
                <span className="sr-only">unpublished changes</span>
              )}
            </NavLink>
          ))}
        </nav>
      )}
    </div>
  );
}
