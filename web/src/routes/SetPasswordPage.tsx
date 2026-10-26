import { useState, type FormEvent } from "react";
import { useNavigate, useSearchParams } from "react-router";

import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { RequestFailed } from "../lib/api";
import { setPasswordWithToken } from "../lib/session";

/**
 * SetPasswordPage is where an invitation and a password reset both land. It
 * is the first screen somebody invited ever sees, so it says what is
 * happening rather than only asking for a password.
 */
export function SetPasswordPage() {
  const navigate = useNavigate();
  const [address] = useSearchParams();
  const token = address.get("token") ?? "";

  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [problem, setProblem] = useState("");

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setProblem("");
    setBusy(true);

    try {
      await setPasswordWithToken(token, password);
      navigate("/projects");
    } catch (error) {
      setProblem(
        error instanceof RequestFailed
          ? error.message
          : "We could not reach the server. Nothing is lost; please try again.",
      );
    } finally {
      setBusy(false);
    }
  }

  if (!token) {
    return (
      <main className="mx-auto flex min-h-dvh max-w-xl flex-col justify-center gap-4 p-6">
        <h1 className="text-[28px] font-semibold tracking-tight">This link is not complete</h1>
        <p className="text-[15px] text-muted-foreground">
          Open the link from your email again, or ask whoever invited you to send another.
        </p>
      </main>
    );
  }

  return (
    <main className="flex min-h-dvh flex-wrap">
      <div className="flex min-w-0 shrink grow basis-[420px] items-center justify-center p-6 sm:p-8">
        <div className="w-full max-w-[380px]">
          <span className="flex items-center gap-2">
            <img src="/quiqr-logo.svg" alt="Quiqr" className="h-6 w-auto" />
            <span className="text-[15px] text-muted-foreground">ChatCMS</span>
          </span>

          <h1 className="mt-8 text-[28px] font-semibold tracking-tight">Choose a password</h1>
          <p className="mt-2 text-[15px] text-muted-foreground">
            Then you are in. You can change it later.
          </p>

          <form onSubmit={onSubmit} className="mt-6 flex flex-col gap-[22px]">
            <Field
              label="New password"
              type="password"
              name="new-password"
              autoComplete="new-password"
              required
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              hint="At least twelve characters. A few words you will remember beats a short tangle."
            />

            {problem && (
              <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
                {problem}
              </p>
            )}

            <Button type="submit" disabled={busy || password === ""}>
              {busy ? "Saving..." : "Set my password"}
            </Button>
          </form>

          <p className="mt-6 text-[13px] text-subtle-foreground">
            Links run out. If this one has, ask whoever invited you for another.
          </p>
        </div>
      </div>

      <div className="flex min-w-0 shrink grow basis-[420px] items-center justify-center bg-primary-soft p-8">
        <p className="max-w-[320px] text-[17px] text-primary-soft-foreground">
          Fill in a form, or just ask. Your site updates as you go.
        </p>
      </div>
    </main>
  );
}
