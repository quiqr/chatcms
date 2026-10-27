import { useState } from "react";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { Table, TableCell, TableHeaderCell, TableWrapper } from "../components/ui/table";
import { useAsync } from "../lib/useAsync";
import { listOrganisationPeople, renameOrganisation, useHere } from "../lib/whoami";
import { say } from "../lib/say";
import { whatWentWrong } from "../lib/api";
import { named as roleNamed } from "../shells/roles";

export function OrganisationSettingsPage() {
  const { who, organisation, again } = useHere();
  const named = who.organisations.find((one) => one.slug === organisation);

  const [name, setName] = useState(named?.name ?? "");
  const [busy, setBusy] = useState(false);
  const [trouble, setTrouble] = useState("");
  const [saved, setSaved] = useState(false);

  const [people] = useAsync(() => listOrganisationPeople(organisation), [organisation]);

  async function rename(event: React.FormEvent) {
    event.preventDefault();
    setBusy(true);
    setTrouble("");
    setSaved(false);
    try {
      await renameOrganisation(organisation, name.trim());
      again();
      setSaved(true);
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex flex-col gap-8">
      <form onSubmit={rename} className="flex max-w-[520px] flex-col gap-4">
        <Field
          label={say("common.name")}
          value={name}
          onChange={(event) => setName(event.target.value)}
          required
          hint={`Its address stays /o/${organisation}, so links already sent keep working.`}
        />

        {trouble && (
          <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
            {trouble}
          </p>
        )}
        {saved && (
          <p role="status" className="text-[15px] text-muted-foreground">
            {say("common.saved")}
          </p>
        )}

        <div>
          <Button type="submit" disabled={busy || name.trim() === "" || name.trim() === named?.name}>
            {busy ? say("common.saving") : say("common.save-the-name")}
          </Button>
        </div>
      </form>

      <section className="flex flex-col gap-4">
        <h2 className="text-[15px] font-semibold">{say("organisation-settings.who-is-in-it")}</h2>

        {people.state === "loading" && (
          <p className="text-[15px] text-muted-foreground">{say("common.looking")}</p>
        )}

        {people.state === "failed" && (
          <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
            {people.message}
          </p>
        )}

        {people.state === "ready" && people.value.people.length === 0 && (
          <p className="text-[15px] text-subtle-foreground">{say("organisation-settings.nobody-yet")}</p>
        )}

        {people.state === "ready" && people.value.people.length > 0 && (
          <TableWrapper>
            <Table>
              <thead>
                <tr>
                  <TableHeaderCell>{say("common.person")}</TableHeaderCell>
                  <TableHeaderCell>{say("organisation-settings.in-this-organisation")}</TableHeaderCell>
                </tr>
              </thead>
              <tbody>
                {people.value.people.map((person) => (
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
                      <span className="text-[15px]">{roleNamed(person.role)}</span>
                    </TableCell>
                  </tr>
                ))}
              </tbody>
            </Table>
          </TableWrapper>
        )}

        <p className="text-[13px] text-subtle-foreground">
          {say("organisation-settings.who-may-edit-which-site-is")}
        </p>
      </section>
    </div>
  );
}
