import { useState } from "react";
import { useParams } from "react-router";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import {
  actAs,
  readPerson,
  turnAccountOff,
  turnAccountOn,
  type Belonging,
} from "../lib/administration";
import { useAsync } from "../lib/useAsync";
import { when } from "../lib/when";
import { say } from "../lib/say";
import { whatWentWrong } from "../lib/api";
import { named } from "../shells/roles";

function Places({ heading, places, nothing }: { heading: string; places: Belonging[]; nothing: string }) {
  return (
    <section className="flex flex-col gap-3">
      <h2 className="text-[15px] font-semibold">{heading}</h2>

      {places.length === 0 ? (
        <p className="text-[15px] text-subtle-foreground">{nothing}</p>
      ) : (
        <ul className="flex flex-col gap-2">
          {places.map((place) => (
            <li key={`${place.organisation ?? ""}/${place.slug}`} className="text-[15px]">
              {place.name || place.slug}
              <span className="block text-[13px] text-subtle-foreground">
                {place.organisation ? `/o/${place.organisation}/p/${place.slug}` : `/o/${place.slug}`}
                {" · "}
                {named(place.role)}
              </span>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

/**
 * One person as whoever runs the installation sees them: everywhere they can
 * reach, and whether their account is on. It answers "who is this address,
 * and what can they get at", which until now took a database client.
 */
export function AdministeredPersonPage() {
  const { email = "" } = useParams();
  const [person, again] = useAsync(() => readPerson(email), [email]);

  const [trouble, setTrouble] = useState("");
  const [busy, setBusy] = useState(false);

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

  if (person.state === "loading") {
    return <p className="text-[15px] text-muted-foreground">{say("common.looking")}</p>;
  }
  if (person.state === "failed") {
    return (
      <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
        {person.message}
      </p>
    );
  }

  const here = person.value;

  return (
    <div className="flex max-w-[680px] flex-col gap-8">
      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      <section className="flex flex-col gap-2">
        <h1 className="text-xl font-semibold">{here.displayName || here.email}</h1>
        <p className="text-[15px] text-muted-foreground">{here.email}</p>

        <p className="flex flex-wrap items-center gap-2">
          {here.administers && <Badge tone="published">{say("administered-person.administers-this-installation")}</Badge>}
          {!here.accountOn && <Badge tone="changed">{say("administered-person.account-off")}</Badge>}
          {here.stillToAccept && <Badge tone="changed">{say("common.still-to-accept")}</Badge>}
        </p>

        <p className="text-[13px] text-subtle-foreground">
          {here.lastSeen ? `Last here ${when(here.lastSeen)}` : say("administered-person.has-never-been-here")}
        </p>
      </section>

      <Places
        heading={say("administered-person.organisations")}
        places={here.organisations}
        nothing={say("administered-person.in-none")}
      />

      <Places heading={say("administered-person.sites")} places={here.projects} nothing={say("administered-person.a-role-on-none")} />

      {!here.administers && here.accountOn && (
        <section className="flex flex-col gap-3">
          <h2 className="text-[15px] font-semibold">{say("administered-person.see-what-they-see")}</h2>

          <p className="text-[15px] text-muted-foreground">
            {say("administered-person.you-will-be-this-person-until")}
          </p>

          <div>
            <Button
              variant="outline"
              disabled={busy}
              onClick={() => {
                setBusy(true);
                void actAs(here.email)
                  .then(() => window.location.assign("/projects"))
                  .catch((error: unknown) => {
                    setTrouble(whatWentWrong(error));
                    setBusy(false);
                  });
              }}
            >
              {say("administered-person.act-as")} {here.displayName || here.email}
            </Button>
          </div>
        </section>
      )}

      <section className="flex flex-col gap-3">
        <h2 className="text-[15px] font-semibold">{say("administered-person.their-account")}</h2>

        <p className="text-[15px] text-muted-foreground">
          {say("administered-person.turning-an-account-off-stops-it")}
        </p>

        <div className="flex flex-wrap gap-2">
          {here.accountOn ? (
            <Button variant="outline" disabled={busy} onClick={() => void run(() => turnAccountOff(here.email))}>
              {say("administered-person.turn-it-off")}
            </Button>
          ) : (
            <Button variant="outline" disabled={busy} onClick={() => void run(() => turnAccountOn(here.email))}>
              {say("administered-person.turn-it-on-again")}
            </Button>
          )}
        </div>
      </section>
    </div>
  );
}
