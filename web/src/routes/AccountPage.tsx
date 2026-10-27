import { useState } from "react";

import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { changeOwnName, changeOwnPassword } from "../lib/session";
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

      <section className="flex flex-col gap-2">
        <h2 className="text-[15px] font-semibold">Your email address</h2>
        <p className="text-[15px] text-muted-foreground">
          {who.person.email}
        </p>
        <p className="text-[13px] text-subtle-foreground">
          Changing this means proving the new address, which this version cannot do yet. Ask
          whoever runs this installation.
        </p>
      </section>
    </div>
  );
}
