import { useState } from "react";

import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { Avatar } from "../components/ui/avatar";
import { useAsync } from "../lib/useAsync";
import { when } from "../lib/when";
import {
  askForANewAddress,
  changeOwnName,
  changeOwnPassword,
  endSession,
  removePicture,
  setPicture,
  whereSignedIn,
} from "../lib/session";
import { useHere } from "../lib/whoami";

export function AccountPage() {
  const { who, again } = useHere();

  const [name, setName] = useState(who.person.displayName);
  const [savingName, setSavingName] = useState(false);
  const [nameSaved, setNameSaved] = useState(false);
  const [nameTrouble, setNameTrouble] = useState("");

  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [savingPassword, setSavingPassword] = useState(false);
  const [passwordSaved, setPasswordSaved] = useState(false);
  const [passwordTrouble, setPasswordTrouble] = useState("");

  async function saveName(event: React.FormEvent) {
    event.preventDefault();
    setSavingName(true);
    setNameTrouble("");
    setNameSaved(false);
    try {
      await changeOwnName(name.trim());
      again();
      setNameSaved(true);
    } catch (error: unknown) {
      setNameTrouble(error instanceof Error ? error.message : "That did not work.");
    } finally {
      setSavingName(false);
    }
  }

  async function savePassword(event: React.FormEvent) {
    event.preventDefault();
    setSavingPassword(true);
    setPasswordTrouble("");
    setPasswordSaved(false);
    try {
      await changeOwnPassword(current, next);
      setCurrent("");
      setNext("");
      setPasswordSaved(true);
    } catch (error: unknown) {
      setPasswordTrouble(error instanceof Error ? error.message : "That did not work.");
    } finally {
      setSavingPassword(false);
    }
  }

  return (
    <div className="flex max-w-[520px] flex-col gap-8">
      <YourPicture />

      <form onSubmit={saveName} className="flex flex-col gap-4">
        <h2 className="text-[15px] font-semibold">Your name</h2>

        <Field
          label="Name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          required
          hint="What other people see beside the changes you make."
        />

        {nameTrouble && (
          <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
            {nameTrouble}
          </p>
        )}
        {nameSaved && (
          <p role="status" className="text-[15px] text-muted-foreground">
            Saved.
          </p>
        )}

        <div>
          <Button type="submit" disabled={savingName || name.trim() === ""}>
            {savingName ? "Saving..." : "Save"}
          </Button>
        </div>
      </form>

      <form onSubmit={savePassword} className="flex flex-col gap-4">
        <h2 className="text-[15px] font-semibold">Your password</h2>

        <Field
          label="Your password now"
          type="password"
          autoComplete="current-password"
          value={current}
          onChange={(event) => setCurrent(event.target.value)}
          required
          hint="Asked for even though you are signed in, so that a screen left open is not enough to take your account over."
        />

        <Field
          label="New password"
          type="password"
          autoComplete="new-password"
          value={next}
          onChange={(event) => setNext(event.target.value)}
          required
        />

        {passwordTrouble && (
          <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
            {passwordTrouble}
          </p>
        )}
        {passwordSaved && (
          <p role="status" className="text-[15px] text-muted-foreground">
            Your password is changed.
          </p>
        )}

        <div>
          <Button type="submit" disabled={savingPassword || current === "" || next === ""}>
            {savingPassword ? "Saving..." : "Change my password"}
          </Button>
        </div>
      </form>

      <YourAddress now={who.person.email} />

      <WhereYouAreSignedIn />
    </div>
  );
}

/**
 * YourPicture is the one picture an account has. It is kept with the account
 * rather than in any site, because it crosses organisations and must never
 * be published with somebody's content.
 */
function YourPicture() {
  const { who, again } = useHere();
  const [trouble, setTrouble] = useState("");
  const [busy, setBusy] = useState(false);
  const [shown, setShown] = useState(0);

  async function run(what: () => Promise<void>) {
    setBusy(true);
    setTrouble("");
    try {
      await what();
      setShown((count) => count + 1);
      again();
    } catch (error: unknown) {
      setTrouble(error instanceof Error ? error.message : "That did not work.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="flex flex-col gap-4">
      <h2 className="text-[15px] font-semibold">Your picture</h2>

      <div className="flex flex-wrap items-center gap-4">
        <Avatar
          key={shown}
          changed={shown}
          name={who.person.displayName || who.person.email}
          personId={who.person.id}
          className="size-16"
        />

        <label className="inline-flex min-h-11 cursor-pointer items-center rounded-md border border-input px-3 text-sm font-semibold hover:bg-muted">
          Choose a picture
          <input
            type="file"
            accept="image/png,image/jpeg,image/gif,image/webp"
            className="sr-only"
            onChange={(event) => {
              const chosen = event.target.files?.[0];
              event.target.value = "";
              if (chosen) void run(() => setPicture(chosen));
            }}
          />
        </label>

        <Button variant="outline" disabled={busy} onClick={() => void run(removePicture)}>
          Use my initials
        </Button>
      </div>

      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}
    </section>
  );
}

/**
 * YourAddress asks for a new one and says that nothing changes until the
 * link in the message sent to it is followed. A typo must not be able to
 * lock somebody out of their own account.
 */
function YourAddress({ now }: { now: string }) {
  const [wanted, setWanted] = useState("");
  const [sent, setSent] = useState(false);
  const [trouble, setTrouble] = useState("");
  const [busy, setBusy] = useState(false);

  async function ask(event: React.FormEvent) {
    event.preventDefault();
    setBusy(true);
    setTrouble("");
    setSent(false);
    try {
      await askForANewAddress(wanted.trim());
      setWanted("");
      setSent(true);
    } catch (error: unknown) {
      setTrouble(error instanceof Error ? error.message : "That did not work.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form onSubmit={ask} className="flex flex-col gap-4">
      <h2 className="text-[15px] font-semibold">Your email address</h2>

      <p className="text-[15px] text-muted-foreground">
        You sign in with <strong>{now}</strong>.
      </p>

      <Field
        label="New address"
        type="email"
        value={wanted}
        onChange={(event) => setWanted(event.target.value)}
        hint="We send a message there. Nothing changes until you follow the link in it."
      />

      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}
      {sent && (
        <p role="status" className="text-[15px] text-muted-foreground">
          Check that address for a message. You sign in with {now} until you follow the link.
        </p>
      )}

      <div>
        <Button type="submit" disabled={busy || wanted.trim() === ""}>
          {busy ? "Sending..." : "Send me the link"}
        </Button>
      </div>
    </form>
  );
}

/** WhereYouAreSignedIn is for the laptop somebody left at the office. */
function WhereYouAreSignedIn() {
  const [places, again] = useAsync(whereSignedIn, []);
  const [trouble, setTrouble] = useState("");

  async function end(id: string) {
    setTrouble("");
    try {
      await endSession(id);
      again();
    } catch (error: unknown) {
      setTrouble(error instanceof Error ? error.message : "That did not work.");
    }
  }

  return (
    <section className="flex flex-col gap-4">
      <h2 className="text-[15px] font-semibold">Where you are signed in</h2>

      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      {places.state === "loading" && (
        <p className="text-[15px] text-muted-foreground">Looking...</p>
      )}

      {places.state === "ready" && (
        <ul className="flex flex-col gap-3">
          {places.value.sessions.map((place) => (
            <li key={place.id} className="flex flex-wrap items-center gap-3 border-b border-border-subtle pb-3 last:border-0">
              <span className="flex-1 text-[15px]">
                {place.thisOne ? "This screen" : "Another screen"}
                <span className="block text-[13px] text-subtle-foreground">
                  Last used {when(place.lastSeen)}
                </span>
              </span>

              {!place.thisOne && (
                <Button variant="outline" onClick={() => void end(place.id)}>
                  End it
                </Button>
              )}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
