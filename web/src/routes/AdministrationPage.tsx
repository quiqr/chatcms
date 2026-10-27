import { useState } from "react";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { Table, TableCell, TableHeaderCell, TableWrapper } from "../components/ui/table";
import { useAsync } from "../lib/useAsync";
import {
  appoint,
  listAdministrators,
  listEveryOrganisation,
  listRecords,
  makeOrganisation,
  standDown,
  turnAccountOff,
  turnAccountOn,
} from "../lib/administration";

function slugFrom(name: string) {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

function NewOrganisation({ onMade }: { onMade: () => void }) {
  const [name, setName] = useState("");
  const [ownerEmail, setOwnerEmail] = useState("");
  const [busy, setBusy] = useState(false);
  const [trouble, setTrouble] = useState("");

  const slug = slugFrom(name);

  async function make(event: React.FormEvent) {
    event.preventDefault();
    setBusy(true);
    setTrouble("");
    try {
      await makeOrganisation({ slug, name: name.trim(), ownerEmail: ownerEmail.trim() });
      setName("");
      setOwnerEmail("");
      onMade();
    } catch (error: unknown) {
      setTrouble(error instanceof Error ? error.message : "That did not work.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form onSubmit={make} className="flex flex-col gap-4 rounded-xl border border-border bg-card p-4">
      <h2 className="text-[15px] font-semibold">Start an organisation</h2>

      <Field
        label="Name"
        value={name}
        onChange={(event) => setName(event.target.value)}
        required
        hint={slug ? `Its address will be /o/${slug}` : undefined}
      />

      <Field
        label="Who owns it"
        type="email"
        value={ownerEmail}
        onChange={(event) => setOwnerEmail(event.target.value)}
        required
        hint="Their email address. They must already have an account."
      />

      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      <Button type="submit" disabled={busy || !name.trim() || !ownerEmail.trim()}>
        {busy ? "Starting..." : "Start it"}
      </Button>

      <p className="text-[13px] text-subtle-foreground">
        You will not be in it. To work inside an organisation, somebody there has to invite you.
      </p>
    </form>
  );
}

function Administrators({ onChanged }: { onChanged: () => void }) {
  const [people, again] = useAsync(listAdministrators, []);
  const [email, setEmail] = useState("");
  const [trouble, setTrouble] = useState("");

  async function run(what: () => Promise<void>) {
    setTrouble("");
    try {
      await what();
      again();
      onChanged();
    } catch (error: unknown) {
      setTrouble(error instanceof Error ? error.message : "That did not work.");
    }
  }

  return (
    <section className="flex flex-col gap-4">
      <h2 className="text-[15px] font-semibold">Who administers this installation</h2>

      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      {people.state === "loading" && <p className="text-[15px] text-muted-foreground">Looking...</p>}

      {people.state === "ready" && (
        <TableWrapper>
          <Table>
            <thead>
              <tr>
                <TableHeaderCell>Person</TableHeaderCell>
                <TableHeaderCell> </TableHeaderCell>
              </tr>
            </thead>
            <tbody>
              {people.value.administrators.map((person) => (
                <tr key={person.email}>
                  <TableCell>
                    <span className="flex flex-wrap items-center gap-2">
                      <span className="text-[15px]">{person.displayName || person.email}</span>
                      {person.yourself && <Badge tone="published">You</Badge>}
                    </span>
                    {person.displayName && (
                      <span className="block text-[13px] text-subtle-foreground">{person.email}</span>
                    )}
                  </TableCell>
                  <TableCell>
                    <Button
                      variant="outline"
                      onClick={() => run(() => standDown(person.email))}
                      aria-label={`Stop ${person.email} administering`}
                    >
                      Stop
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
          void run(() => appoint(asked));
        }}
      >
        <div className="min-w-[240px] flex-1">
          <Field
            label="Appoint somebody"
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder="their email address"
          />
        </div>
        <Button type="submit" variant="outline" disabled={!email.trim()}>
          Appoint
        </Button>
      </form>
    </section>
  );
}

/**
 * Accounts is where somebody who has left is stopped. An account belongs to
 * the installation rather than to an organisation, so this is the only place
 * it can be done: an organisation's owner turning one off would lock that
 * person out of every other organisation too.
 */
function Accounts({ onChanged }: { onChanged: () => void }) {
  const [email, setEmail] = useState("");
  const [said, setSaid] = useState("");
  const [trouble, setTrouble] = useState("");

  async function run(what: () => Promise<void>, done: string) {
    setTrouble("");
    setSaid("");
    try {
      await what();
      setSaid(done);
      onChanged();
    } catch (error: unknown) {
      setTrouble(error instanceof Error ? error.message : "That did not work.");
    }
  }

  const asked = email.trim();

  return (
    <section className="flex flex-col gap-4">
      <h2 className="text-[15px] font-semibold">Somebody who has left</h2>

      <p className="text-[15px] text-muted-foreground">
        Turning an account off stops it being used anywhere here, and ends whatever it has open.
        What that person wrote stays theirs.
      </p>

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

      <div className="flex flex-wrap items-end gap-3">
        <div className="min-w-[240px] flex-1">
          <Field
            label="Their email address"
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
          />
        </div>

        <Button
          variant="outline"
          disabled={asked === ""}
          onClick={() => void run(() => turnAccountOff(asked), `${asked} can no longer sign in.`)}
        >
          Turn it off
        </Button>

        <Button
          variant="outline"
          disabled={asked === ""}
          onClick={() => void run(() => turnAccountOn(asked), `${asked} can sign in again.`)}
        >
          Turn it on again
        </Button>
      </div>
    </section>
  );
}

function What({ since }: { since: number }) {
  // since changes whenever something was done on this screen, so that what
  // you just did appears without reloading the page.
  const [records] = useAsync(listRecords, [since]);

  if (records.state === "loading") {
    return <p className="text-[15px] text-muted-foreground">Looking...</p>;
  }
  if (records.state === "failed") {
    return (
      <p role="alert" className="text-sm text-destructive-foreground">
        {records.message}
      </p>
    );
  }
  if (records.value.records.length === 0) {
    return <p className="text-[15px] text-subtle-foreground">Nothing has been done yet.</p>;
  }

  return (
    <ul className="flex flex-col gap-3">
      {records.value.records.map((record, index) => (
        <li key={index} className="border-b border-border-subtle pb-3 last:border-0">
          <p className="text-[15px]">
            {record.who} {record.did}
            {record.to && <span className="font-semibold"> {record.to}</span>}
          </p>
          <p className="text-[13px] text-subtle-foreground">
            {new Date(record.at).toLocaleString()}
            {record.detail?.ownedBy && ` · owned by ${record.detail.ownedBy}`}
            {record.detail?.person && ` · ${record.detail.person}`}
          </p>
        </li>
      ))}
    </ul>
  );
}

export function AdministrationPage() {
  const [done, setDone] = useState(0);
  const somethingWasDone = () => setDone((count) => count + 1);
  const [organisations, again] = useAsync(listEveryOrganisation, [done]);

  return (
    <div className="flex flex-col gap-8">
      <section className="flex flex-col gap-4">
        <h2 className="text-[15px] font-semibold">Organisations</h2>

        {organisations.state === "loading" && (
          <p className="text-[15px] text-muted-foreground">Looking...</p>
        )}

        {organisations.state === "failed" && (
          <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
            {organisations.message}
          </p>
        )}

        {organisations.state === "ready" && (
          <div className="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(min(280px,100%),1fr))]">
            {organisations.value.organisations.map((organisation) => (
              <div key={organisation.slug} className="rounded-xl border border-border bg-card p-4">
                <p className="text-[15px] font-semibold">{organisation.name}</p>
                <p className="text-[13px] text-subtle-foreground">/o/{organisation.slug}</p>
                <p className="mt-2 text-[13px] text-subtle-foreground">
                  {organisation.projects} {organisation.projects === 1 ? "site" : "sites"} ·{" "}
                  {organisation.people} {organisation.people === 1 ? "person" : "people"}
                </p>
              </div>
            ))}

            {organisations.value.organisations.length === 0 && (
              <p className="text-[15px] text-subtle-foreground">There are none yet.</p>
            )}
          </div>
        )}
      </section>

      <NewOrganisation
        onMade={() => {
          again();
          somethingWasDone();
        }}
      />

      <Administrators onChanged={somethingWasDone} />

      <Accounts onChanged={somethingWasDone} />

      <section className="flex flex-col gap-4">
        <h2 className="text-[15px] font-semibold">What has been done</h2>
        <What since={done} />
      </section>
    </div>
  );
}
