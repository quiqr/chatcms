import { MoreHorizontal, Search, UserPlus } from "lucide-react";
import { useState } from "react";
import { NavLink } from "react-router";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { Menu, MenuContent, MenuItem, MenuSeparator, MenuTrigger } from "../components/ui/menu";
import { Table, TableCell, TableHeaderCell, TableWrapper } from "../components/ui/table";
import { whatWentWrong } from "../lib/api";
import {
  appoint,
  changeAnAccount,
  deleteAnAccount,
  listEveryOrganisation,
  listEveryPerson,
  putSomebodyInto,
  sendAPasswordReset,
  standDown,
  turnAccountOff,
  turnAccountOn,
  type PersonWhole,
} from "../lib/administration";
import { say } from "../lib/say";
import { when } from "../lib/when";
import { named } from "../shells/roles";
import { useAsync } from "../lib/useAsync";
import { useHere } from "../lib/whoami";

const aPage = 8;

/**
 * Every account on the installation, and what may be done to one.
 *
 * An account is in one of three states, and only one of them is stored:
 * blocked is a column, invited is "has never set a password", and active is
 * neither. A third column would be a second thing to keep in step with the
 * first.
 *
 * chatcms-audience: implementer
 */
export function PlatformUsers() {
  const { who } = useHere();

  const [search, setSearch] = useState("");
  const [asked, setAsked] = useState("");
  const [only, setOnly] = useState("");
  const [pages, setPages] = useState<string[]>([""]);

  const after = pages[pages.length - 1] ?? "";
  const [people, again] = useAsync(
    () => listEveryPerson(asked, only, after, aPage),
    [asked, only, after],
  );

  const [trouble, setTrouble] = useState("");
  const [said, setSaid] = useState("");
  const [editing, setEditing] = useState<PersonWhole | null>(null);
  const [deleting, setDeleting] = useState<PersonWhole | null>(null);
  const [adding, setAdding] = useState<PersonWhole | null>(null);

  async function run(what: () => Promise<unknown>, afterwards = "") {
    setTrouble("");
    setSaid("");
    try {
      await what();
      setSaid(afterwards);
      again();
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error));
    }
  }

  function stateOf(person: PersonWhole) {
    if (!person.accountOn) return { tone: "changed" as const, says: say("users.blocked") };
    if (person.stillToAccept) return { tone: "changed" as const, says: say("users.invited") };
    return { tone: "published" as const, says: say("users.active") };
  }

  const chips = [
    { value: "", label: say("users.all") },
    { value: "platform-admins", label: say("users.platform-admins") },
    { value: "invited", label: say("users.invited") },
    { value: "blocked", label: say("users.blocked") },
    { value: "in-no-organisation", label: say("users.in-no-organisation") },
  ];

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center gap-3">
        <h1 className="flex-1 text-[28px] font-semibold tracking-tight">{say("platform.users")}</h1>

        <form
          onSubmit={(event) => {
            event.preventDefault();
            setPages([""]);
            setAsked(search);
          }}
        >
          <label className="flex h-11 min-w-[220px] items-center gap-2 rounded-md border border-input bg-card px-3">
            <Search aria-hidden="true" className="size-4 text-subtle-foreground" />
            <span className="sr-only">{say("users.search-by-name-or-email")}</span>
            <input
              // chatcms-target: the label around this one carries the hit area
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder={say("users.search-by-name-or-email")}
              className="h-full min-w-0 flex-1 bg-transparent text-[15px] outline-hidden"
            />
          </label>
        </form>

        <Button variant="ink" asChild>
          <NavLink to="/administration/users/new">
            <UserPlus aria-hidden="true" className="size-4" />
            {say("users.invite-user")}
          </NavLink>
        </Button>
      </div>

      <div className="flex flex-wrap gap-2">
        {chips.map((chip) => (
          <button
            key={chip.value || "all"}
            type="button"
            aria-pressed={only === chip.value}
            onClick={() => {
              setPages([""]);
              setOnly(chip.value);
            }}
            className={`flex min-h-11 items-center rounded-full border px-3 text-sm ${
              only === chip.value
                ? "border-primary bg-primary-tint font-medium text-primary-soft-foreground"
                : "border-border text-muted-foreground hover:bg-muted"
            }`}
          >
            {chip.label}
          </button>
        ))}
      </div>

      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}
      {said && (
        <p role="status" className="text-[15px] text-muted-foreground">
          {said}
        </p>
      )}

      {editing && (
        <Editing
          person={editing}
          onDone={(name, address) => {
            const was = editing.email;
            setEditing(null);
            void run(() => changeAnAccount(was, name, address));
          }}
          onCancel={() => setEditing(null)}
        />
      )}

      {adding && (
        <Adding
          person={adding}
          onDone={(organisation, role) => {
            const them = adding.email;
            setAdding(null);
            void run(() => putSomebodyInto(organisation, them, role));
          }}
          onCancel={() => setAdding(null)}
        />
      )}

      {deleting && (
        <Deleting
          person={deleting}
          onYes={() => {
            const gone = deleting.email;
            setDeleting(null);
            void run(() => deleteAnAccount(gone));
          }}
          onCancel={() => setDeleting(null)}
        />
      )}

      {people.state === "loading" && (
        <p className="text-[15px] text-muted-foreground">{say("common.looking")}</p>
      )}

      {people.state === "failed" && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {people.message}
        </p>
      )}

      {people.state === "ready" && people.value.people.length === 0 && (
        <p className="text-[15px] text-subtle-foreground">{say("users.nobody-matches-that")}</p>
      )}

      {people.state === "ready" && people.value.people.length > 0 && (
        <>
          <TableWrapper>
            <Table>
              <thead>
                <tr>
                  <TableHeaderCell>{say("users.user")}</TableHeaderCell>
                  <TableHeaderCell>{say("users.organisations")}</TableHeaderCell>
                  <TableHeaderCell>{say("users.status")}</TableHeaderCell>
                  <TableHeaderCell>{say("users.last-sign-in")}</TableHeaderCell>
                  <TableHeaderCell>
                    <span className="sr-only">{say("people.actions")}</span>
                  </TableHeaderCell>
                </tr>
              </thead>
              <tbody>
                {people.value.people.map((person) => {
                  const state = stateOf(person);
                  const yourself = person.email === who.person.email;

                  return (
                    <tr key={person.personId}>
                      <TableCell>
                        <NavLink
                          to={`/administration/people/${encodeURIComponent(person.email)}`}
                          className="inline-flex min-h-11 items-center gap-2 text-[15px] font-semibold underline-offset-2 hover:underline"
                        >
                          {person.displayName || person.email}
                          {person.administers && (
                            <Badge tone="neutral">{say("platform.platform-admin")}</Badge>
                          )}
                        </NavLink>
                        <span className="block text-[13px] text-subtle-foreground">{person.email}</span>
                      </TableCell>
                      <TableCell>
                        {person.organisations.length === 0 ? (
                          <span className="text-[13px] text-subtle-foreground">
                            {say("users.in-no-organisation")}
                          </span>
                        ) : (
                          <span className="text-[13px] text-subtle-foreground">
                            {person.organisations
                              .map((where) =>
                                where.role === "admin"
                                  ? `${where.name} (${named(where.role)})`
                                  : where.name,
                              )
                              .join(", ")}
                          </span>
                        )}
                      </TableCell>
                      <TableCell>
                        <Badge tone={state.tone}>{state.says}</Badge>
                      </TableCell>
                      <TableCell>
                        <span className="text-[13px] text-subtle-foreground">
                          {person.lastSeen ? when(person.lastSeen) : say("people.never")}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        <Menu>
                          <MenuTrigger
                            aria-label={say("users.what-can-be-done-with-them", {
                              name: person.displayName || person.email,
                            })}
                            className="flex size-11 items-center justify-center rounded-md hover:bg-muted"
                          >
                            <MoreHorizontal aria-hidden="true" className="size-4" />
                          </MenuTrigger>
                          <MenuContent>
                            <MenuItem onSelect={() => setEditing(person)}>
                              {say("users.edit-name-and-email")}
                            </MenuItem>
                            <MenuItem
                              onSelect={() =>
                                void run(
                                  () => sendAPasswordReset(person.email),
                                  say("users.reset-sent", { address: person.email }),
                                )
                              }
                            >
                              {say("users.send-password-reset")}
                            </MenuItem>
                            <MenuItem onSelect={() => setAdding(person)}>
                              {say("users.add-to-organisation")}
                            </MenuItem>
                            <MenuSeparator />
                            <MenuItem
                              disabled={yourself}
                              onSelect={() =>
                                void run(() =>
                                  person.administers
                                    ? standDown(person.email)
                                    : appoint(person.email),
                                )
                              }
                            >
                              {person.administers
                                ? say("users.stop-being-platform-admin")
                                : say("users.make-platform-admin")}
                            </MenuItem>
                            <MenuItem
                              disabled={yourself}
                              onSelect={() =>
                                void run(() =>
                                  person.accountOn
                                    ? turnAccountOff(person.email)
                                    : turnAccountOn(person.email),
                                )
                              }
                            >
                              {person.accountOn
                                ? say("users.block-sign-in")
                                : say("users.unblock-sign-in")}
                            </MenuItem>
                            <MenuSeparator />
                            <MenuItem
                              destructive
                              disabled={yourself}
                              onSelect={() => setDeleting(person)}
                            >
                              {say("users.delete-user")}
                            </MenuItem>
                          </MenuContent>
                        </Menu>
                      </TableCell>
                    </tr>
                  );
                })}
              </tbody>
            </Table>
          </TableWrapper>

          <div className="flex flex-wrap items-center gap-3">
            <p className="flex-1 text-[13px] text-subtle-foreground">
              {say("users.showing", {
                shown: people.value.shown,
                altogether: people.value.altogether,
              })}
            </p>

            {pages.length > 1 && (
              <Button variant="outline" onClick={() => setPages(pages.slice(0, -1))}>
                {say("users.back")}
              </Button>
            )}
            {people.value.after && (
              <Button
                variant="outline"
                onClick={() => setPages([...pages, people.value.after ?? ""])}
              >
                {say("users.more")}
              </Button>
            )}
          </div>
        </>
      )}
    </div>
  );
}

function Editing({
  person,
  onDone,
  onCancel,
}: {
  person: PersonWhole;
  onDone: (name: string, address: string) => void;
  onCancel: () => void;
}) {
  const [name, setName] = useState(person.displayName);
  const [address, setAddress] = useState(person.email);

  return (
    <form
      className="flex flex-wrap items-end gap-3 rounded-xl border border-border bg-card p-4"
      onSubmit={(event) => {
        event.preventDefault();
        onDone(name.trim(), address.trim());
      }}
    >
      <div className="min-w-[200px] flex-1">
        <Field
          label={say("users.their-name")}
          value={name}
          onChange={(event) => setName(event.target.value)}
        />
      </div>
      <div className="min-w-[200px] flex-1">
        <Field
          label={say("users.their-email")}
          type="email"
          value={address}
          onChange={(event) => setAddress(event.target.value)}
          required
        />
      </div>
      <Button type="submit">{say("users.save")}</Button>
      <Button variant="ghost" onClick={onCancel}>
        {say("users.cancel")}
      </Button>
    </form>
  );
}

function Deleting({
  person,
  onYes,
  onCancel,
}: {
  person: PersonWhole;
  onYes: () => void;
  onCancel: () => void;
}) {
  return (
    <section
      aria-labelledby="delete-user-title"
      className="flex flex-col gap-3 rounded-xl border border-destructive-soft p-4"
    >
      <h2 id="delete-user-title" className="text-base font-semibold">
        {say("users.delete-asked", { name: person.displayName || person.email })}
      </h2>
      <p className="text-[15px]">{say("users.delete-says-what-stays")}</p>
      <div className="flex flex-wrap gap-2">
        <Button variant="destructive" onClick={onYes}>
          {say("users.yes-delete-them")}
        </Button>
        <Button variant="outline" onClick={onCancel}>
          {say("users.cancel")}
        </Button>
      </div>
    </section>
  );
}

/**
 * Putting somebody into an organisation from here. The organisations are
 * asked for when the form opens rather than with every row, because this is
 * the only place that needs them.
 */
function Adding({
  person,
  onDone,
  onCancel,
}: {
  person: PersonWhole;
  onDone: (organisation: string, role: string) => void;
  onCancel: () => void;
}) {
  const [organisations] = useAsync(() => listEveryOrganisation(), []);
  const [chosen, setChosen] = useState("");
  const [role, setRole] = useState("member");

  const theirs = organisations.state === "ready" ? organisations.value.organisations : [];

  return (
    <form
      className="flex flex-wrap items-end gap-3 rounded-xl border border-border bg-card p-4"
      onSubmit={(event) => {
        event.preventDefault();
        onDone(chosen || (theirs[0]?.slug ?? ""), role);
      }}
    >
      <h2 className="w-full text-[15px] font-semibold">
        {say("users.add-them-to", { name: person.displayName || person.email })}
      </h2>

      <label className="flex min-w-[200px] flex-1 flex-col gap-1 text-sm font-medium">
        {say("users.which-organisation")}
        <select
          value={chosen}
          onChange={(event) => setChosen(event.target.value)}
          className="h-11 rounded-md border border-input bg-card px-2 text-[15px]"
        >
          {theirs.map((organisation) => (
            <option key={organisation.slug} value={organisation.slug}>
              {organisation.name}
            </option>
          ))}
        </select>
      </label>

      <select
        aria-label={say("manage.role-in-the-organisation")}
        value={role}
        onChange={(event) => setRole(event.target.value)}
        className="h-11 rounded-md border border-input bg-card px-2 text-[15px]"
      >
        <option value="member">{named("member")}</option>
        <option value="admin">{named("admin")}</option>
      </select>

      <Button type="submit" disabled={theirs.length === 0}>
        {say("users.add")}
      </Button>
      <Button variant="ghost" onClick={onCancel}>
        {say("users.cancel")}
      </Button>
    </form>
  );
}
