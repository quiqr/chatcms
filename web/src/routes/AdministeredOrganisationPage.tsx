import { useState } from "react";
import { useNavigate, useParams } from "react-router";

import { Ban } from "lucide-react";

import { Badge } from "../components/ui/badge";
import { HowFull, whyThereIsNoRoom, type Usage } from "../components/ui/usage";
import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { Table, TableCell, TableHeaderCell, TableWrapper } from "../components/ui/table";
import {
  putSomebodyInto,
  readOrganisation,
  removeOrganisation,
  setLimits,
  takeSomebodyOut,
} from "../lib/administration";
import { useAsync } from "../lib/useAsync";
import { renameOrganisation } from "../lib/whoami";
import { plural, say } from "../lib/say";
import { whatWentWrong } from "../lib/api";
import { named } from "../shells/roles";

/**
 * One organisation as whoever runs the installation sees it: everything
 * about it in one place. It is for an organisation with nobody left to run
 * it, or one being set up; its own admin manages it day to day.
 */
export function AdministeredOrganisationPage() {
  const { organisation = "" } = useParams();
  const goTo = useNavigate();

  const [whole, again] = useAsync(() => readOrganisation(organisation), [organisation]);

  const [name, setName] = useState<string | null>(null);
  const [email, setEmail] = useState("");
  const [role, setRole] = useState("member");
  const [trouble, setTrouble] = useState("");
  const [busy, setBusy] = useState(false);
  const [removing, setRemoving] = useState(false);
  const [typed, setTyped] = useState("");

  async function run(what: () => Promise<unknown>) {
    setBusy(true);
    setTrouble("");
    try {
      await what();
      again();
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error));
    } finally {
      setBusy(false);
    }
  }

  if (whole.state === "loading") {
    return <p className="text-[15px] text-muted-foreground">{say("common.looking")}</p>;
  }
  if (whole.state === "failed") {
    return (
      <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
        {whole.message}
      </p>
    );
  }

  const here = whole.value;
  const people = here.people ?? [];
  const projects = here.projects ?? [];
  const noRoomForAMember = whyThereIsNoRoom(here.usage.members, "seats");
  const shownName = name ?? here.name;

  return (
    <div className="flex max-w-[680px] flex-col gap-8">
      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      <form
        className="flex flex-col gap-4"
        onSubmit={(event) => {
          event.preventDefault();
          void run(() => renameOrganisation(organisation, shownName.trim()));
        }}
      >
        <Field
          label={say("common.name")}
          value={shownName}
          onChange={(event) => setName(event.target.value)}
          required
          hint={`Its address stays /o/${here.slug}.`}
        />
        <div>
          <Button type="submit" disabled={busy || shownName.trim() === "" || shownName === here.name}>
            {say("common.save-the-name")}
          </Button>
        </div>
      </form>

      <section className="flex flex-col gap-4">
        <h2 className="text-[15px] font-semibold">{say("administered-organisation.who-is-in-it")}</h2>

        <HowFull counted={here.usage.members} of="seats" />

        {noRoomForAMember && (
          <p className="flex items-start gap-2 rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
            <Ban aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
            {noRoomForAMember}
          </p>
        )}

        {people.length === 0 ? (
          <p className="text-[15px] text-subtle-foreground">
            {say("administered-organisation.nobody-give-it-an-owner-so")}
          </p>
        ) : (
          <TableWrapper>
            <Table>
              <thead>
                <tr>
                  <TableHeaderCell>{say("common.person")}</TableHeaderCell>
                  <TableHeaderCell>{say("administered-organisation.role")}</TableHeaderCell>
                  <TableHeaderCell>
                    <span className="sr-only">{say("administered-organisation.actions")}</span>
                  </TableHeaderCell>
                </tr>
              </thead>
              <tbody>
                {people.map((person) => (
                  <tr key={person.personId}>
                    <TableCell>
                      <span className="flex flex-wrap items-center gap-2">
                        <span className="text-[15px]">{person.displayName || person.email}</span>
                        {person.stillToAccept && <Badge tone="changed">{say("common.still-to-accept")}</Badge>}
                      </span>
                    </TableCell>
                    <TableCell>
                      <span className="text-[15px]">{named(person.role)}</span>
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="outline"
                        disabled={busy}
                        onClick={() => void run(() => takeSomebodyOut(organisation, person.email))}
                        aria-label={`Take ${person.email} out`}
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

        <form
          className="flex flex-wrap items-end gap-3"
          onSubmit={(event) => {
            event.preventDefault();
            const asked = email.trim();
            setEmail("");
            void run(() => putSomebodyInto(organisation, asked, role));
          }}
        >
          <div className="min-w-[220px] flex-1">
            <Field
              label={say("administered-organisation.put-somebody-in")}
              type="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              placeholder={say("common.their-email-address")}
            />
          </div>

          <select
            aria-label={say("administered-organisation.role-for-the-person-you-put")}
            value={role}
            onChange={(event) => setRole(event.target.value)}
            className="h-11 rounded-md border border-input bg-card px-2 text-[15px] capitalize"
          >
            <option value="member">{say("common.role-member")}</option>
            <option value="admin">{say("common.role-admin")}</option>
          </select>

          <Button
            type="submit"
            variant="outline"
            disabled={busy || email.trim() === "" || noRoomForAMember !== null}
          >
            {say("administered-organisation.put-them-in")}
          </Button>
        </form>
      </section>

      <section className="flex flex-col gap-4">
        <h2 className="text-[15px] font-semibold">{say("administered-organisation.its-sites")}</h2>

        <HowFull counted={here.usage.projects} of="sites" />

        {projects.length === 0 ? (
          <p className="text-[15px] text-subtle-foreground">{say("administered-organisation.none")}</p>
        ) : (
          <ul className="flex flex-col gap-2">
            {projects.map((project) => (
              <li key={project.slug} className="text-[15px]">
                {project.name}
                {/* chatcms-not-a-string: an address */}
                <span className="block text-[13px] text-subtle-foreground">
                  /o/{here.slug}/p/{project.slug}
                </span>
              </li>
            ))}
          </ul>
        )}
      </section>

      <Limits organisation={organisation} usage={here.usage} onSaved={again} />

      <section className="flex flex-col gap-4 rounded-xl border border-destructive-soft p-4">
        <h2 className="text-[15px] font-semibold text-destructive-foreground">
          {say("platform.delete-organisation")}
        </h2>

        <p className="text-[15px] text-muted-foreground">
          {say("platform.delete-says-what-goes", {
            name: here.name,
            sites: plural(projects.length, "common.site-one", "common.site-many"),
            members: plural(people.length, "common.person-one", "common.person-many"),
          })}
        </p>

        {removing ? (
          <div className="flex flex-col gap-3">
            <h3 className="text-base font-semibold">
              {say("platform.delete-organisation-asked", { name: here.name })}
            </h3>
            <ul className="flex flex-col gap-1 text-[15px]">
              <li>
                {say("platform.sites-are-removed", {
                  sites: plural(projects.length, "common.site-one", "common.site-many"),
                })}
              </li>
              <li>
                {say("platform.members-lose-access", {
                  members: plural(people.length, "common.person-one", "common.person-many"),
                })}
              </li>
              <li>{say("platform.websites-keep-working")}</li>
            </ul>

            <Field
              label={say("platform.type-to-confirm", { name: here.name })}
              value={typed}
              onChange={(event) => setTyped(event.target.value)}
            />

            <div className="flex flex-wrap gap-2">
              <Button
                variant="destructive"
                disabled={busy || typed.trim() !== here.name}
                onClick={() => {
                  setBusy(true);
                  void removeOrganisation(organisation)
                    .then(() => goTo("/administration/organisations"))
                    .catch((error: unknown) => {
                      setTrouble(whatWentWrong(error));
                      setBusy(false);
                    });
                }}
              >
                {say("platform.yes-delete-it")}
              </Button>
              <Button variant="outline" onClick={() => setRemoving(false)}>
                {say("common.keep-it")}
              </Button>
            </div>
          </div>
        ) : (
          <div>
            <Button variant="outline" onClick={() => setRemoving(true)}>
              {say("administered-organisation.remove-this-organisation-2")}
            </Button>
          </div>
        )}
      </section>
    </div>
  );
}

/**
 * What an organisation is allowed. A limit is a number or nothing at all,
 * which is why each one is a checkbox beside a stepper rather than a number
 * with zero meaning something.
 *
 * chatcms-audience: implementer
 */
function Limits({
  organisation,
  usage,
  onSaved,
}: {
  organisation: string;
  usage: Usage;
  onSaved: () => void;
}) {
  const [members, setMembers] = useState<number | null>(usage.members.allowed ?? null);
  const [projects, setProjects] = useState<number | null>(usage.projects.allowed ?? null);
  const [busy, setBusy] = useState(false);
  const [saved, setSaved] = useState(false);
  const [trouble, setTrouble] = useState("");

  function save() {
    setBusy(true);
    setTrouble("");
    setSaved(false);
    void setLimits(organisation, members, projects)
      .then(() => {
        setSaved(true);
        onSaved();
      })
      .catch((error: unknown) => setTrouble(whatWentWrong(error)))
      .finally(() => setBusy(false));
  }

  return (
    <section className="flex flex-col gap-4">
      <h2 className="text-[15px] font-semibold">{say("platform.limits")}</h2>

      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      <Stepper
        label={say("platform.limit-members")}
        value={members}
        held={usage.members.held}
        onChange={setMembers}
      />
      <Stepper
        label={say("platform.limit-sites")}
        value={projects}
        held={usage.projects.held}
        onChange={setProjects}
      />

      <p className="text-[13px] text-subtle-foreground">{say("platform.limits-are-optional-untick")}</p>

      <div className="flex flex-wrap items-center gap-3">
        <Button onClick={save} disabled={busy}>
          {say("platform.save-changes")}
        </Button>
        {saved && (
          <p role="status" className="text-[15px] text-muted-foreground">
            {say("common.saved")}
          </p>
        )}
      </div>
    </section>
  );
}

function Stepper({
  label,
  value,
  held,
  onChange,
}: {
  label: string;
  value: number | null;
  held: number;
  onChange: (next: number | null) => void;
}) {
  const limited = value !== null;

  return (
    <div className="flex flex-wrap items-center gap-3">
      <label className="flex min-h-11 items-center gap-2 text-[15px]">
        {/* chatcms-target: the label around this one carries the hit area */}
        <input
          type="checkbox"
          checked={limited}
          onChange={(event) => onChange(event.target.checked ? Math.max(held, 1) : null)}
          className="size-4"
        />
        {label}
      </label>

      {limited ? (
        <span className="flex items-center gap-2">
          <Button
            variant="outline"
            aria-label={say("platform.fewer")}
            disabled={value <= 1}
            onClick={() => onChange(Math.max(1, value - 1))}
          >
            −
          </Button>
          <span className="min-w-8 text-center text-[15px] tabular-nums">{value}</span>
          <Button variant="outline" aria-label={say("platform.more")} onClick={() => onChange(value + 1)}>
            +
          </Button>
        </span>
      ) : (
        <span className="text-[13px] text-subtle-foreground">
          {say("limits.no-limit", { held })}
        </span>
      )}
    </div>
  );
}
