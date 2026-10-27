import { useState } from "react";
import { useParams } from "react-router";

import { Avatar } from "../components/ui/avatar";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { Menu, MenuContent, MenuItem, MenuTrigger } from "../components/ui/menu";
import { Table, TableCell, TableHeaderCell, TableWrapper } from "../components/ui/table";
import { useAsync } from "../lib/useAsync";
import { useHere } from "../lib/whoami";
import {
  changeRole,
  invitePerson,
  listPeople,
  removePerson,
  roles,
  type Member,
} from "../lib/tenancy";

const whatEachRoleCanDo: { role: string; says: string }[] = [
  { role: "Owner", says: "Everything, including billing and deleting the project." },
  {
    role: "Implementer",
    says: "Sets up where the site lives, how it is published, and who may use it; sees technical detail.",
  },
  { role: "Editor", says: "Edits pages and media, and publishes the site." },
  { role: "Viewer", says: "Reads content and previews; cannot change anything." },
];

function RoleSelect({
  value,
  onChange,
  label,
}: {
  value: string;
  onChange: (role: string) => void;
  label: string;
}) {
  return (
    <select
      aria-label={label}
      value={value}
      onChange={(event) => onChange(event.target.value)}
      className="h-11 rounded-md border border-input bg-card px-2 text-[15px] capitalize"
    >
      {roles.map((role) => (
        <option key={role} value={role}>
          {role}
        </option>
      ))}
    </select>
  );
}

export function PeoplePage() {
  const { organisation } = useHere();
  const { project = "" } = useParams();
  const [people, loadAgain] = useAsync(() => listPeople(organisation, project), [organisation, project]);

  const [email, setEmail] = useState("");
  const [role, setRole] = useState<string>("editor");
  const [problem, setProblem] = useState("");
  const [busy, setBusy] = useState(false);

  async function invite() {
    setProblem("");
    setBusy(true);
    try {
      await invitePerson(organisation, project, email, role);
      setEmail("");
      loadAgain();
    } catch (error) {
      setProblem(error instanceof Error ? error.message : "The invitation could not be sent.");
    } finally {
      setBusy(false);
    }
  }

  async function change(member: Member, next: string) {
    setProblem("");
    try {
      await changeRole(organisation, project, member.personId, next);
      loadAgain();
    } catch (error) {
      setProblem(error instanceof Error ? error.message : "That change could not be made.");
    }
  }

  async function take(member: Member) {
    setProblem("");
    try {
      await removePerson(organisation, project, member.personId);
      loadAgain();
    } catch (error) {
      setProblem(error instanceof Error ? error.message : "That person could not be removed.");
    }
  }

  return (
    <div className="flex flex-wrap items-start gap-6">
      <div className="min-w-0 max-w-full shrink grow-[999] basis-[520px]">
        <h2 className="text-[22px] font-semibold tracking-tight">People &amp; roles</h2>

        <form
          onSubmit={(event) => {
            event.preventDefault();
            void invite();
          }}
          className="mt-4 flex flex-wrap items-end gap-3"
        >
          <div className="min-w-[220px] flex-1">
            <Field
              label="Invite someone"
              type="email"
              required
              placeholder="their email address"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
            />
          </div>
          <RoleSelect value={role} onChange={setRole} label="Role for the person you invite" />
          <Button type="submit" variant="ink" disabled={busy}>
            {busy ? "Sending..." : "Send invite"}
          </Button>
        </form>

        {problem && (
          <p role="alert" className="mt-3 rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
            {problem}
          </p>
        )}

        {people.state === "loading" && (
          <p className="mt-6 text-[15px] text-muted-foreground">Looking up who is here...</p>
        )}

        {people.state === "failed" && (
          <p role="alert" className="mt-6 rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
            {people.message}
          </p>
        )}

        {people.state === "ready" && (
          <TableWrapper className="mt-6 rounded-xl border border-border bg-card">
            <Table>
              <thead>
                <tr>
                  <TableHeaderCell>Person</TableHeaderCell>
                  <TableHeaderCell>Role</TableHeaderCell>
                  <TableHeaderCell>
                    <span className="sr-only">Actions</span>
                  </TableHeaderCell>
                </tr>
              </thead>
              <tbody>
                {people.value.people.map((member) => (
                  <tr key={member.personId}>
                    <TableCell>
                      <span className="flex items-center gap-2.5">
                        <Avatar name={member.displayName || member.email} pending={member.stillToAccept} />
                        <span className="flex min-w-0 flex-col">
                          <span className="truncate text-[15px]">
                            {member.displayName || member.email}
                          </span>
                          {member.displayName && (
                            <span className="truncate text-[13px] text-subtle-foreground">
                              {member.email}
                            </span>
                          )}
                        </span>
                        {member.stillToAccept && <Badge tone="changed">Still to accept</Badge>}
                      </span>
                    </TableCell>
                    <TableCell>
                      <RoleSelect
                        value={member.role}
                        onChange={(next) => void change(member, next)}
                        label={`Role for ${member.displayName || member.email}`}
                      />
                    </TableCell>
                    <TableCell className="text-right">
                      <Menu>
                        <MenuTrigger
                          aria-label={`Actions for ${member.displayName || member.email}`}
                          className="inline-flex size-11 items-center justify-center rounded-md hover:bg-muted"
                        >
                          ...
                        </MenuTrigger>
                        <MenuContent align="end">
                          <MenuItem destructive onSelect={() => void take(member)}>
                            Remove from this project
                          </MenuItem>
                        </MenuContent>
                      </Menu>
                    </TableCell>
                  </tr>
                ))}
              </tbody>
            </Table>
          </TableWrapper>
        )}
      </div>

      <aside className="min-w-0 shrink grow basis-[280px] rounded-xl border border-border bg-card p-4">
        <h3 className="text-[15px] font-semibold">What each role can do</h3>
        <dl className="mt-3 flex flex-col gap-3">
          {whatEachRoleCanDo.map((entry) => (
            <div key={entry.role}>
              <dt className="text-sm font-medium">{entry.role}</dt>
              <dd className="text-[13px] text-subtle-foreground">{entry.says}</dd>
            </div>
          ))}
        </dl>
      </aside>
    </div>
  );
}
