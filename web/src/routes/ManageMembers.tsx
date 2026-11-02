import { Ban } from "lucide-react";
import { useState } from "react";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { Table, TableCell, TableHeaderCell, TableWrapper } from "../components/ui/table";
import { whyThereIsNoRoom } from "../components/ui/usage";
import { whatWentWrong } from "../lib/api";
import { say } from "../lib/say";
import { when } from "../lib/when";
import { named } from "../shells/roles";
import { inviteToOrganisation, removeFromOrganisation } from "../lib/whoami";
import type { OrganisationArea } from "../lib/tenancy";

/**
 * Who belongs to an organisation, managed from the organisation. A seat is
 * held by a membership whether or not it has been accepted, so the count
 * beside the heading is what the limit reads.
 */
export function ManageMembers({ area, again }: { area: OrganisationArea; again: () => void }) {
  const [email, setEmail] = useState("");
  const [role, setRole] = useState("member");
  const [trouble, setTrouble] = useState("");
  const [busy, setBusy] = useState(false);

  const noRoom = whyThereIsNoRoom(area.usage.members, "seats");

  async function run(what: () => Promise<unknown>) {
    setBusy(true);
    setTrouble("");
    try {
      await what();
      again();
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error, "people.the-invitation-could-not-be-sent"));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-[28px] font-semibold tracking-tight">{say("manage.members")}</h1>

      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      {noRoom && (
        <p className="flex items-start gap-2 rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          <Ban aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
          {noRoom}
        </p>
      )}

      <form
        className="flex flex-wrap items-end gap-3 rounded-xl border border-border bg-card p-4"
        onSubmit={(event) => {
          event.preventDefault();
          const asked = email.trim();
          setEmail("");
          void run(() => inviteToOrganisation(area.slug, asked, role));
        }}
      >
        <h2 className="w-full text-[15px] font-semibold">
          {say("manage.invite-to", { name: area.name })}
        </h2>

        <div className="min-w-[220px] flex-1">
          <Field
            label={say("common.email")}
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder={say("common.their-email-address")}
            disabled={noRoom !== null}
          />
        </div>

        <select
          aria-label={say("manage.role-in-the-organisation")}
          value={role}
          onChange={(event) => setRole(event.target.value)}
          disabled={noRoom !== null}
          className="h-11 rounded-md border border-input bg-card px-2 text-[15px]"
        >
          <option value="member">{named("member")}</option>
          <option value="admin">{named("admin")}</option>
        </select>

        <Button type="submit" disabled={busy || email.trim() === "" || noRoom !== null}>
          {say("manage.send-invite")}
        </Button>
      </form>

      {area.people.length === 0 ? (
        <p className="text-[15px] text-subtle-foreground">{say("manage.nobody-yet")}</p>
      ) : (
        <TableWrapper>
          <Table>
            <thead>
              <tr>
                <TableHeaderCell>{say("manage.person")}</TableHeaderCell>
                <TableHeaderCell>{say("manage.role-in-the-organisation")}</TableHeaderCell>
                <TableHeaderCell>{say("manage.their-sites")}</TableHeaderCell>
                <TableHeaderCell>{say("manage.last-active")}</TableHeaderCell>
                <TableHeaderCell>
                  <span className="sr-only">{say("people.actions")}</span>
                </TableHeaderCell>
              </tr>
            </thead>
            <tbody>
              {area.people.map((person) => (
                <tr key={person.personId}>
                  <TableCell>
                    <span className="flex flex-wrap items-center gap-2">
                      <span className="text-[15px]">{person.displayName || person.email}</span>
                      {person.stillToAccept && <Badge tone="changed">{say("common.still-to-accept")}</Badge>}
                    </span>
                    {person.displayName && (
                      <span className="block text-[13px] text-subtle-foreground">{person.email}</span>
                    )}
                  </TableCell>
                  <TableCell>
                    <span className="text-[15px]">{named(person.role)}</span>
                  </TableCell>
                  <TableCell>
                    {person.role === "admin" ? (
                      <span className="text-[13px] text-subtle-foreground">{say("manage.all-sites")}</span>
                    ) : person.projects.length === 0 ? (
                      <span className="text-[13px] text-subtle-foreground">{say("manage.no-sites-yet")}</span>
                    ) : (
                      <span className="flex flex-wrap gap-1">
                        {person.projects.map((project) => (
                          <Badge key={project.slug} tone="neutral">
                            {project.name} · {named(project.role)}
                          </Badge>
                        ))}
                      </span>
                    )}
                  </TableCell>
                  <TableCell>
                    <span className="text-[13px] text-subtle-foreground">
                      {person.lastSeen ? when(person.lastSeen) : say("people.never")}
                    </span>
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      variant="outline"
                      disabled={busy}
                      aria-label={say("manage.remove-from-the-organisation", { name: person.email })}
                      onClick={() => void run(() => removeFromOrganisation(area.slug, person.personId))}
                    >
                      {say("administered-organisation.take-out")}
                    </Button>
                  </TableCell>
                </tr>
              ))}
            </tbody>
          </Table>
        </TableWrapper>
      )}
    </div>
  );
}
