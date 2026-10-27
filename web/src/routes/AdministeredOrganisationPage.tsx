import { useState } from "react";
import { useNavigate, useParams } from "react-router";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { Table, TableCell, TableHeaderCell, TableWrapper } from "../components/ui/table";
import {
  putSomebodyInto,
  readOrganisation,
  removeOrganisation,
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
 * it, or one being set up; its own owner manages it day to day.
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
            <option value="owner">{say("common.role-owner")}</option>
          </select>

          <Button type="submit" variant="outline" disabled={busy || email.trim() === ""}>
            {say("administered-organisation.put-them-in")}
          </Button>
        </form>
      </section>

      <section className="flex flex-col gap-4">
        <h2 className="text-[15px] font-semibold">{say("administered-organisation.its-sites")}</h2>

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

      <section className="flex flex-col gap-4">
        <h2 className="text-[15px] font-semibold">{say("administered-organisation.remove-this-organisation")}</h2>

        {projects.length > 0 ? (
          <p className="text-[15px] text-muted-foreground">
            {say("administered-organisation.it-still-holds-sites", {
              sites: plural(projects.length, "common.site-one", "common.site-many"),
            })}
          </p>
        ) : removing ? (
          <div className="flex flex-col gap-3">
            <p className="text-[15px]">
              {say("administered-organisation.this-removes-it", { name: here.name })}
            </p>
            <p className="text-[15px]">
              {say("administered-organisation.nothing-in-git-is-touched-every")}
            </p>
            <div className="flex flex-wrap gap-2">
              <Button
                variant="destructive"
                disabled={busy}
                onClick={() => {
                  setBusy(true);
                  void removeOrganisation(organisation)
                    .then(() => goTo("/administration"))
                    .catch((error: unknown) => {
                      setTrouble(whatWentWrong(error));
                      setBusy(false);
                    });
                }}
              >
                {say("common.yes-remove-it")}
              </Button>
              <Button variant="outline" onClick={() => setRemoving(false)}>
                {say("common.keep-it")}
              </Button>
            </div>
          </div>
        ) : (
          <Button variant="outline" onClick={() => setRemoving(true)}>
            {say("administered-organisation.remove-this-organisation-2")}
          </Button>
        )}
      </section>
    </div>
  );
}
