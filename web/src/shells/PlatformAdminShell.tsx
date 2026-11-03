import { Building2, History, Shield, Users } from "lucide-react";
import type { ReactNode } from "react";
import { NavLink } from "react-router";

import { PersonMenu } from "./PersonMenu";
import { say } from "../lib/say";

/**
 * The shell of the platform admin area. Dark on purpose: everything in here
 * is done to somebody else's organisation, and a screen that looks like the
 * rest of the product is a screen somebody can be in without noticing.
 *
 * chatcms-audience: implementer
 */
export function PlatformAdminShell({
  organisations,
  users,
  viewer,
  children,
}: {
  organisations: number;
  users: number;
  viewer?: { name: string };
  children: ReactNode;
}) {
  const item = ({ isActive }: { isActive: boolean }) =>
    `flex min-h-11 items-center gap-2 rounded-md px-2 text-[15px] ${
      isActive ? "bg-admin-area-active font-medium" : "hover:bg-admin-area-active"
    }`;

  return (
    <div className="min-h-dvh sm:flex">
      <aside
        aria-label={say("platform.platform-admin")}
        className="flex flex-col gap-4 bg-admin-area p-4 text-admin-area-foreground sm:min-h-dvh sm:w-[260px] sm:shrink-0"
      >
        <NavLink to="/projects" className="flex min-h-11 items-center gap-2">
          <img src="/quiqr-logo.svg" alt={say("organisation-shell.quiqr-chatcms")} className="h-6 w-auto" />
          <span className="text-sm font-semibold">{say("common.chatcms")}</span>
        </NavLink>

        <p className="flex items-center gap-2 rounded-lg bg-admin-area-raised px-3 py-2">
          <Shield aria-hidden="true" className="size-4" />
          <span className="flex flex-col">
            <span className="text-xs opacity-70">{say("platform.you-are-in")}</span>
            <span className="text-[15px] font-semibold">{say("platform.platform-admin")}</span>
          </span>
        </p>

        <nav className="flex flex-col gap-1">
          <NavLink to="/administration/organisations" className={item}>
            <Building2 aria-hidden="true" className="size-4" />
            {say("platform.organisations")}
            <span className="ml-auto text-[13px] opacity-70">{organisations}</span>
          </NavLink>
          <NavLink to="/administration/users" className={item}>
            <Users aria-hidden="true" className="size-4" />
            {say("platform.users")}
            <span className="ml-auto text-[13px] opacity-70">{users}</span>
          </NavLink>
          <NavLink to="/administration/history" className={item}>
            <History aria-hidden="true" className="size-4" />
            {say("platform.history")}
          </NavLink>
        </nav>

        <span className="flex-1" />

        <NavLink
          to="/projects"
          className="flex min-h-11 items-center px-2 text-sm opacity-80 hover:opacity-100"
        >
          {say("platform.back-to-my-projects")}
        </NavLink>

        {viewer && <PersonMenu name={viewer.name} />}
      </aside>

      <main className="w-full max-w-[1100px] p-6">{children}</main>
    </div>
  );
}
