import { useState, type FormEvent } from "react";
import { NavLink } from "react-router";

import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { requestPasswordReset } from "../lib/session";

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
      setProblem(
        error instanceof Error && error.message
          ? error.message
          : "We could not reach the server. Please try again.",
      );
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="mx-auto flex min-h-dvh max-w-[420px] flex-col justify-center gap-6 p-6">
      <span className="flex items-center gap-2">
        <img src="/quiqr-logo.svg" alt="Quiqr" className="h-6 w-auto" />
        <span className="text-[15px] text-muted-foreground">ChatCMS</span>
      </span>

      <h1 className="text-[28px] font-semibold tracking-tight">Set a new password</h1>

      {sent ? (
        <>
          <p role="status" className="text-[15px]">
            If that address has an account, a message is on its way with a link in it. The link
            works for an hour.
          </p>
          <p className="text-[15px] text-muted-foreground">
            Nothing arrived? Look in your spam, and check the address you typed.
          </p>
          <NavLink to="/sign-in" className="inline-flex min-h-11 items-center text-[15px] text-primary">
            Back to signing in
          </NavLink>
        </>
      ) : (
        <form onSubmit={onSubmit} className="flex flex-col gap-[22px]">
          <Field
            label="Email"
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
            {busy ? "Sending..." : "Send me a link"}
          </Button>

          <NavLink to="/sign-in" className="inline-flex min-h-11 items-center text-[15px] text-primary">
            Back to signing in
          </NavLink>
        </form>
      )}
    </main>
  );
}
