import { useState } from "react";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { Table, TableCell, TableHeaderCell, TableWrapper } from "../components/ui/table";
import { useAsync } from "../lib/useAsync";
import { listOrganisationPeople, renameOrganisation, useHere } from "../lib/whoami";

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
      setTrouble(error instanceof Error ? error.message : "That did not work.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex flex-col gap-8">
      <form onSubmit={rename} className="flex max-w-[520px] flex-col gap-4">
        <Field
          label="Name"
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
            Saved.
          </p>
        )}

        <div>
          <Button type="submit" disabled={busy || name.trim() === "" || name.trim() === named?.name}>
            {busy ? "Saving..." : "Save the name"}
          </Button>
        </div>
      </form>

      <section className="flex flex-col gap-4">
        <h2 className="text-[15px] font-semibold">Who is in it</h2>

        {people.state === "loading" && (
          <p className="text-[15px] text-muted-foreground">Looking...</p>
        )}

        {people.state === "failed" && (
          <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
            {people.message}
          </p>
        )}

        {people.state === "ready" && people.value.people.length === 0 && (
          <p className="text-[15px] text-subtle-foreground">Nobody yet.</p>
        )}

        {people.state === "ready" && people.value.people.length > 0 && (
          <TableWrapper>
            <Table>
              <thead>
                <tr>
                  <TableHeaderCell>Person</TableHeaderCell>
                  <TableHeaderCell>In this organisation</TableHeaderCell>
                </tr>
              </thead>
              <tbody>
                {people.value.people.map((person) => (
                  <tr key={person.personId}>
                    <TableCell>
                      <span className="flex flex-wrap items-center gap-2">
                        <span className="text-[15px]">{person.displayName || person.email}</span>
                        {person.stillToAccept && <Badge tone="changed">Still to accept</Badge>}
                      </span>
                      {person.displayName && (
                        <span className="block text-[13px] text-subtle-foreground">{person.email}</span>
                      )}
                    </TableCell>
                    <TableCell>
                      <span className="text-[15px] capitalize">{person.role}</span>
                    </TableCell>
                  </tr>
                ))}
              </tbody>
            </Table>
          </TableWrapper>
        )}

        <p className="text-[13px] text-subtle-foreground">
          Who may edit which site is set on each site, under People and roles.
        </p>
      </section>
    </div>
  );
}
