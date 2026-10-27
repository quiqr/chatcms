import { useState } from "react";
import { NavLink } from "react-router";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { Table, TableCell, TableHeaderCell, TableWrapper } from "../components/ui/table";
import { useAsync } from "../lib/useAsync";
import { when } from "../lib/when";
import { plural, say } from "../lib/say";
import {
  appoint,
  listAdministrators,
  listEveryOrganisation,
  listEveryPerson,
  listRecords,
  makeOrganisation,
  makePerson,
  standDown,
} from "../lib/administration";
import { whatWentWrong } from "../lib/api";
import { named } from "../shells/roles";

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
      setTrouble(whatWentWrong(error));
    } finally {
      setBusy(false);
    }
  }

  return (
    <form onSubmit={make} className="flex flex-col gap-4 rounded-xl border border-border bg-card p-4">
      <h2 className="text-[15px] font-semibold">{say("administration.start-an-organisation")}</h2>

      <Field
        label={say("common.name")}
        value={name}
        onChange={(event) => setName(event.target.value)}
        required
        hint={slug ? `Its address will be /o/${slug}` : undefined}
      />

      <Field
        label={say("administration.who-owns-it")}
        type="email"
        value={ownerEmail}
        onChange={(event) => setOwnerEmail(event.target.value)}
        required
        hint={say("administration.their-email-address-they-must-already")}
      />

      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      <Button type="submit" disabled={busy || !name.trim() || !ownerEmail.trim()}>
        {busy ? say("administration.starting") : say("administration.start-it")}
      </Button>

      <p className="text-[13px] text-subtle-foreground">
        {say("administration.you-will-not-be-in-it")}
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
      setTrouble(whatWentWrong(error));
    }
  }

  return (
    <section className="flex flex-col gap-4">
      <h2 className="text-[15px] font-semibold">{say("administration.who-administers-this-installation")}</h2>

      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      {people.state === "loading" && <p className="text-[15px] text-muted-foreground">{say("common.looking")}</p>}

      {people.state === "ready" && (
        <TableWrapper>
          <Table>
            <thead>
              <tr>
                <TableHeaderCell>{say("common.person")}</TableHeaderCell>
                <TableHeaderCell> </TableHeaderCell>
              </tr>
            </thead>
            <tbody>
              {people.value.administrators.map((person) => (
                <tr key={person.email}>
                  <TableCell>
                    <span className="flex flex-wrap items-center gap-2">
                      <span className="text-[15px]">{person.displayName || person.email}</span>
                      {person.yourself && <Badge tone="published">{say("administration.you")}</Badge>}
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
                      {say("administration.stop")}
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
            label={say("administration.appoint-somebody")}
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder={say("common.their-email-address")}
          />
        </div>
        <Button type="submit" variant="outline" disabled={!email.trim()}>
          {say("administration.appoint")}
        </Button>
      </form>
    </section>
  );
}

/**
 * People is everybody on the installation. It answers "who is this address,
 * and what can they reach" without a database client, and is the way to one
 * person's own screen, where their account is turned off.
 */
function People() {
  const [search, setSearch] = useState("");
  const [asked, setAsked] = useState("");
  const [people] = useAsync(() => listEveryPerson(asked), [asked]);

  return (
    <section className="flex flex-col gap-4">
      <h2 className="text-[15px] font-semibold">{say("administration.people")}</h2>

      <form
        className="flex flex-wrap items-end gap-3"
        onSubmit={(event) => {
          event.preventDefault();
          setAsked(search);
        }}
      >
        <div className="min-w-[240px] flex-1">
          <Field
            label={say("administration.find-somebody")}
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder={say("administration.part-of-a-name-or-an")}
          />
        </div>
        <Button type="submit" variant="outline">
          {say("administration.find")}
        </Button>
        {asked !== "" && (
          <Button
            variant="ghost"
            onClick={() => {
              setSearch("");
              setAsked("");
            }}
          >
            {say("administration.everybody")}
          </Button>
        )}
      </form>

      {people.state === "loading" && <p className="text-[15px] text-muted-foreground">{say("common.looking")}</p>}

      {people.state === "failed" && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {people.message}
        </p>
      )}

      {people.state === "ready" && people.value.people.length === 0 && (
        <p className="text-[15px] text-subtle-foreground">{say("administration.nobody-here-matches-that")}</p>
      )}

      {people.state === "ready" && people.value.people.length > 0 && (
        <>
          <TableWrapper>
            <Table>
              <thead>
                <tr>
                  <TableHeaderCell>{say("common.person")}</TableHeaderCell>
                  <TableHeaderCell>{say("administration.where-they-belong")}</TableHeaderCell>
                  <TableHeaderCell>{say("administration.last-here")}</TableHeaderCell>
                </tr>
              </thead>
              <tbody>
                {people.value.people.map((person) => (
                  <tr key={person.personId}>
                    <TableCell>
                      <NavLink
                        to={`/administration/people/${encodeURIComponent(person.email)}`}
                        className="inline-flex min-h-11 items-center text-[15px] font-semibold underline-offset-2 hover:underline"
                      >
                        {person.displayName || person.email}
                      </NavLink>
                      <span className="block text-[13px] text-subtle-foreground">{person.email}</span>
                      <span className="mt-1 flex flex-wrap items-center gap-2">
                        {person.administers && <Badge tone="published">{say("administration.administers")}</Badge>}
                        {!person.accountOn && <Badge tone="changed">{say("administration.account-off")}</Badge>}
                        {person.stillToAccept && <Badge tone="changed">{say("common.still-to-accept")}</Badge>}
                      </span>
                    </TableCell>
                    <TableCell>
                      {person.organisations.length === 0 && person.projects.length === 0 ? (
                        <span className="text-[13px] text-subtle-foreground">{say("administration.nowhere-yet")}</span>
                      ) : (
                        <span className="text-[13px] text-subtle-foreground">
                          {person.organisations.map((where) => `${where.name} (${named(where.role)})`).join(", ")}
                          {person.organisations.length > 0 && person.projects.length > 0 && " · "}
                          {person.projects.map((site) => `${site.name} (${named(site.role)})`).join(", ")}
                        </span>
                      )}
                    </TableCell>
                    <TableCell>
                      <span className="text-[13px] text-subtle-foreground">
                        {person.lastSeen ? when(person.lastSeen) : say("administration.never")}
                      </span>
                    </TableCell>
                  </tr>
                ))}
              </tbody>
            </Table>
          </TableWrapper>

          {people.value.shown < people.value.altogether && (
            <p className="text-[13px] text-subtle-foreground">
              {say("administration.showing-of", {
                shown: people.value.shown,
                altogether: people.value.altogether,
              })}
            </p>
          )}
        </>
      )}
    </section>
  );
}

function NewPerson({ onMade }: { onMade: () => void }) {
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [said, setSaid] = useState("");
  const [trouble, setTrouble] = useState("");
  const [busy, setBusy] = useState(false);

  return (
    <form
      className="flex flex-col gap-4 rounded-xl border border-border bg-card p-4"
      onSubmit={(event) => {
        event.preventDefault();
        const asked = email.trim();
        setBusy(true);
        setTrouble("");
        setSaid("");
        void makePerson({ email: asked, displayName: name.trim() })
          .then(() => {
            setEmail("");
            setName("");
            setSaid(`${asked} has been sent a way in.`);
            onMade();
          })
          .catch((error: unknown) => {
            setTrouble(whatWentWrong(error));
          })
          .finally(() => setBusy(false));
      }}
    >
      <h2 className="text-[15px] font-semibold">{say("administration.make-a-person")}</h2>

      <Field
        label={say("common.email")}
        type="email"
        value={email}
        onChange={(event) => setEmail(event.target.value)}
        required
      />

      <Field
        label={say("administration.their-name")}
        value={name}
        onChange={(event) => setName(event.target.value)}
        hint={say("administration.they-can-change-it-themselves-afterwards")}
      />

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

      <Button type="submit" disabled={busy || email.trim() === ""}>
        {busy ? say("administration.making") : say("administration.make-them")}
      </Button>

      <p className="text-[13px] text-subtle-foreground">
        {say("administration.they-choose-their-own-password-from")}
      </p>
    </form>
  );
}

function What({ since }: { since: number }) {
  // since changes whenever something was done on this screen, so that what
  // you just did appears without reloading the page.
  const [records] = useAsync(listRecords, [since]);

  if (records.state === "loading") {
    return <p className="text-[15px] text-muted-foreground">{say("common.looking")}</p>;
  }
  if (records.state === "failed") {
    return (
      <p role="alert" className="text-sm text-destructive-foreground">
        {records.message}
      </p>
    );
  }
  if (records.value.records.length === 0) {
    return <p className="text-[15px] text-subtle-foreground">{say("administration.nothing-has-been-done-yet")}</p>;
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
            {when(record.at)}
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
        <h2 className="text-[15px] font-semibold">{say("administration.organisations")}</h2>

        {organisations.state === "loading" && (
          <p className="text-[15px] text-muted-foreground">{say("common.looking")}</p>
        )}

        {organisations.state === "failed" && (
          <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
            {organisations.message}
          </p>
        )}

        {organisations.state === "ready" && (
          <div className="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(min(280px,100%),1fr))]">
            {organisations.value.organisations.map((organisation) => (
              <NavLink
                key={organisation.slug}
                to={`/administration/o/${organisation.slug}`}
                className="rounded-xl border border-border bg-card p-4 hover:border-input"
              >
                <p className="text-[15px] font-semibold">{organisation.name}</p>
                {/* chatcms-not-a-string: an address */}
                <p className="text-[13px] text-subtle-foreground">/o/{organisation.slug}</p>
                <p className="mt-2 text-[13px] text-subtle-foreground">
                  {plural(organisation.projects, "common.site-one", "common.site-many")} ·{" "}
                  {plural(organisation.people, "common.person-one", "common.person-many")}
                </p>
              </NavLink>
            ))}

            {organisations.value.organisations.length === 0 && (
              <p className="text-[15px] text-subtle-foreground">{say("administration.there-are-none-yet")}</p>
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

      <People key={done} />

      <NewPerson onMade={somethingWasDone} />

      <section className="flex flex-col gap-4">
        <h2 className="text-[15px] font-semibold">{say("administration.what-has-been-done")}</h2>
        <What since={done} />
      </section>
    </div>
  );
}
