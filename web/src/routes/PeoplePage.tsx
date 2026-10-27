import { useState } from "react";
import { useParams } from "react-router";

import { Avatar } from "../components/ui/avatar";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { Menu, MenuContent, MenuItem, MenuTrigger } from "../components/ui/menu";
import { Table, TableCell, TableHeaderCell, TableWrapper } from "../components/ui/table";
import { Segmented } from "../components/ui/segmented";
import { useAsync } from "../lib/useAsync";
import {
  changeOrganisationRole,
  inviteToOrganisation,
  listOrganisationPeople,
  removeFromOrganisation,
  sendTheInvitationAgain,
  useHere,
} from "../lib/whoami";
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

// The roles an organisation has, which are not the roles a project has.
const organisationRoles = ["owner", "member"] as const;

function RoleSelect({
  value,
  onChange,
  label,
  choices,
}: {
  value: string;
  onChange: (role: string) => void;
  label: string;
  choices: readonly string[];
}) {
  return (
    <select
      aria-label={label}
      value={value}
      onChange={(event) => onChange(event.target.value)}
      className="h-11 rounded-md border border-input bg-card px-2 text-[15px] capitalize"
    >
      {choices.map((role) => (
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

  // Who may edit this site, and who belongs to the organisation around it,
  // are two different questions about two different sets of people. The
  // design draws them as one screen with a switch rather than two screens
  // somebody has to know about.
  const [level, setLevel] = useState<"project" | "organisation">("project");
  const here = level === "project";

  const [people, loadAgain] = useAsync(
    () => (here ? listPeople(organisation, project) : listOrganisationPeople(organisation)),
    [organisation, project, level],
  );

  const [email, setEmail] = useState("");
  const [role, setRole] = useState<string>("editor");
  const [problem, setProblem] = useState("");
  const [busy, setBusy] = useState(false);

  async function invite() {
    setProblem("");
    setBusy(true);
    try {
      await (here
        ? invitePerson(organisation, project, email, role)
        : inviteToOrganisation(organisation, email, role));
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
      await (here
        ? changeRole(organisation, project, member.personId, next)
        : changeOrganisationRole(organisation, member.personId, next));
      loadAgain();
    } catch (error) {
      setProblem(error instanceof Error ? error.message : "That change could not be made.");
    }
  }

  async function sendAgain(member: Member) {
    setProblem("");
    try {
      await sendTheInvitationAgain(organisation, member.personId);
    } catch (error) {
      setProblem(error instanceof Error ? error.message : "That could not be sent again.");
    }
  }

  async function take(member: Member) {
    setProblem("");
    try {
      await (here
        ? removePerson(organisation, project, member.personId)
        : removeFromOrganisation(organisation, member.personId));
      loadAgain();
    } catch (error) {
      setProblem(error instanceof Error ? error.message : "That person could not be removed.");
    }
  }

  return (
    <div className="flex flex-wrap items-start gap-6">
      <div className="min-w-0 max-w-full shrink grow-[999] basis-[520px]">
        <h2 className="text-[22px] font-semibold tracking-tight">People &amp; roles</h2>

        <div className="mt-4">
          <Segmented
            label="Who to show"
            value={level}
            onValueChange={(next) => {
              setLevel(next as "project" | "organisation");
              setRole(next === "project" ? "editor" : "member");
              setProblem("");
            }}
            options={[
              { value: "project", label: "This project" },
              { value: "organisation", label: "Organisation" },
            ]}
          />
        </div>

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
          <RoleSelect
            value={role}
            onChange={setRole}
            label="Role for the person you invite"
            choices={here ? roles : organisationRoles}
          />
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
                  <TableHeaderCell>Last active</TableHeaderCell>
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
                        choices={here ? roles : organisationRoles}
                      />
                    </TableCell>
                    <TableCell>
                      <span className="text-[13px] text-subtle-foreground">
                        {member.stillToAccept
                          ? "Not yet"
                          : member.lastSeen
                            ? new Date(member.lastSeen).toLocaleString()
                            : "Never"}
                      </span>
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
                          {member.stillToAccept && (
                            <MenuItem onSelect={() => void sendAgain(member)}>
                              Send the invitation again
                            </MenuItem>
                          )}
                          <MenuItem destructive onSelect={() => void take(member)}>
                            {member.stillToAccept
                              ? "Withdraw the invitation"
                              : here
                                ? "Remove from this project"
                                : "Remove from the organisation"}
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
