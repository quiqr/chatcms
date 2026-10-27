import { useState, type FormEvent } from "react";
import { NavLink } from "react-router";

import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { requestPasswordReset } from "../lib/session";
import { say } from "../lib/say";
import { whatWentWrong } from "../lib/api";

/**
 * ForgotPasswordPage asks for an address and then says the same thing
 * whatever the answer was. Telling somebody that an address is unknown would
 * turn this screen into a way of finding out who has an account.
 */
export function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [busy, setBusy] = useState(false);
  const [sent, setSent] = useState(false);
  const [problem, setProblem] = useState("");

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setProblem("");
    setBusy(true);

    try {
      await requestPasswordReset(email);
      setSent(true);
    } catch (error) {
      setProblem(whatWentWrong(error, "common.cannot-reach-the-server-again"));
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="mx-auto flex min-h-dvh max-w-[420px] flex-col justify-center gap-6 p-6">
      <span className="flex items-center gap-2">
        <img src="/quiqr-logo.svg" alt={say("common.quiqr")} className="h-6 w-auto" />
        <span className="text-[15px] text-muted-foreground">{say("common.chatcms")}</span>
      </span>

      <h1 className="text-[28px] font-semibold tracking-tight">{say("forgot-password.set-a-new-password")}</h1>

      {sent ? (
        <>
          <p role="status" className="text-[15px]">
            {say("forgot-password.if-that-address-has-an-account")}
          </p>
          <p className="text-[15px] text-muted-foreground">
            {say("forgot-password.nothing-arrived-look-in-your-spam")}
          </p>
          <NavLink to="/sign-in" className="inline-flex min-h-11 items-center text-[15px] text-primary">
            {say("forgot-password.back-to-signing-in")}
          </NavLink>
        </>
      ) : (
        <form onSubmit={onSubmit} className="flex flex-col gap-[22px]">
          <Field
            label={say("common.email")}
            type="email"
            name="email"
            autoComplete="email"
            required
            value={email}
            onChange={(event) => setEmail(event.target.value)}
          />

          {problem && (
            <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
              {problem}
            </p>
          )}

          <Button type="submit" disabled={busy || email === ""}>
            {busy ? say("common.sending") : say("forgot-password.send-me-a-link")}
          </Button>

          <NavLink to="/sign-in" className="inline-flex min-h-11 items-center text-[15px] text-primary">
            {say("forgot-password.back-to-signing-in")}
          </NavLink>
        </form>
      )}
    </main>
  );
}
