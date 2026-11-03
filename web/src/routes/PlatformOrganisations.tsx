import { Plus, Search } from "lucide-react";
import { useState } from "react";
import { NavLink } from "react-router";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Table, TableCell, TableHeaderCell, TableWrapper } from "../components/ui/table";
import { Tile } from "../components/ui/avatar";
import { HowFull } from "../components/ui/usage";
import { listEveryOrganisation } from "../lib/administration";
import { say } from "../lib/say";
import { onADay } from "../lib/when";
import { useAsync } from "../lib/useAsync";

/**
 * Every organisation, with what it holds and how full it is. The filters are
 * the server's answer: "at a limit" is a question about all of them, and a
 * thousand are not sent here to be counted.
 *
 * chatcms-audience: implementer
 */
export function PlatformOrganisations({ onNew }: { onNew: () => void }) {
  const [search, setSearch] = useState("");
  const [asked, setAsked] = useState("");
  const [filter, setFilter] = useState("");

  const [organisations] = useAsync(() => listEveryOrganisation(asked, filter), [asked, filter]);

  const chips = [
    { value: "", label: say("platform.all") },
    { value: "at-a-limit", label: say("platform.at-a-limit") },
    { value: "no-admin", label: say("platform.no-admin-yet") },
  ];

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center gap-3">
        <h1 className="flex-1 text-[28px] font-semibold tracking-tight">
          {say("platform.organisations")}
        </h1>

        <form
          onSubmit={(event) => {
            event.preventDefault();
            setAsked(search);
          }}
        >
          {/* A label rather than a bare input in a box: what a thumb aims at
              is the label, which is the whole field. */}
          <label className="flex h-11 min-w-[220px] items-center gap-2 rounded-md border border-input bg-card px-3">
            <Search aria-hidden="true" className="size-4 text-subtle-foreground" />
            <span className="sr-only">{say("platform.search-organisations")}</span>
            <input
              // chatcms-target: the label around this one carries the hit area
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder={say("platform.search-organisations")}
              className="h-full min-w-0 flex-1 bg-transparent text-[15px] outline-hidden"
            />
          </label>
        </form>

        <Button variant="ink" onClick={onNew}>
          <Plus aria-hidden="true" className="size-4" />
          {say("platform.new-organisation")}
        </Button>
      </div>

      <div className="flex flex-wrap gap-2">
        {chips.map((chip) => (
          <button
            key={chip.value || "all"}
            type="button"
            aria-pressed={filter === chip.value}
            onClick={() => setFilter(chip.value)}
            className={`flex min-h-11 items-center rounded-full border px-3 text-sm ${
              filter === chip.value
                ? "border-primary bg-primary-tint font-medium text-primary-soft-foreground"
                : "border-border text-muted-foreground hover:bg-muted"
            }`}
          >
            {chip.label}
          </button>
        ))}
      </div>

      {organisations.state === "loading" && (
        <p className="text-[15px] text-muted-foreground">{say("common.looking")}</p>
      )}

      {organisations.state === "failed" && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {organisations.message}
        </p>
      )}

      {organisations.state === "ready" && organisations.value.organisations.length === 0 && (
        <p className="text-[15px] text-subtle-foreground">{say("platform.nothing-matches-that")}</p>
      )}

      {organisations.state === "ready" && organisations.value.organisations.length > 0 && (
        <TableWrapper>
          <Table>
            <thead>
              <tr>
                <TableHeaderCell>{say("platform.organisation")}</TableHeaderCell>
                <TableHeaderCell>{say("platform.its-admins")}</TableHeaderCell>
                <TableHeaderCell>{say("platform.members")}</TableHeaderCell>
                <TableHeaderCell>{say("platform.sites")}</TableHeaderCell>
                <TableHeaderCell>{say("platform.made")}</TableHeaderCell>
                <TableHeaderCell>
                  <span className="sr-only">{say("people.actions")}</span>
                </TableHeaderCell>
              </tr>
            </thead>
            <tbody>
              {organisations.value.organisations.map((organisation) => (
                <tr key={organisation.slug}>
                  <TableCell>
                    <NavLink
                      to={`/administration/organisations/${organisation.slug}`}
                      className="flex min-h-11 items-center gap-2"
                    >
                      <Tile name={organisation.name} />
                      <span className="flex min-w-0 flex-col">
                        <span className="truncate text-[15px] font-semibold">{organisation.name}</span>
                        {/* chatcms-not-a-string: an address */}
                        <span className="text-[13px] text-subtle-foreground">
                          /o/{organisation.slug}
                        </span>
                      </span>
                    </NavLink>
                  </TableCell>
                  <TableCell>
                    {organisation.admins.length === 0 ? (
                      <Badge tone="changed">{say("platform.no-admin-yet-invite-pending")}</Badge>
                    ) : (
                      <span className="text-[13px] text-subtle-foreground">
                        {organisation.admins
                          .map((admin) => admin.displayName || admin.email)
                          .join(", ")}
                      </span>
                    )}
                  </TableCell>
                  <TableCell>
                    <HowFull counted={organisation.usage.members} of="seats" />
                  </TableCell>
                  <TableCell>
                    <HowFull counted={organisation.usage.projects} of="sites" />
                  </TableCell>
                  <TableCell>
                    <span className="text-[13px] text-subtle-foreground">
                      {organisation.madeBy
                        ? say("platform.made-by", {
                            when: onADay(organisation.madeOn),
                            who: organisation.madeBy,
                          })
                        : onADay(organisation.madeOn)}
                    </span>
                  </TableCell>
                  <TableCell className="text-right">
                    <Button variant="outline" asChild>
                      <NavLink to={`/administration/organisations/${organisation.slug}`}>
                        {say("platform.manage")}
                      </NavLink>
                    </Button>
                  </TableCell>
                </tr>
              ))}
            </tbody>
          </Table>
        </TableWrapper>
      )}

      <p className="text-[13px] text-subtle-foreground">{say("platform.limits-are-optional")}</p>
    </div>
  );
}
