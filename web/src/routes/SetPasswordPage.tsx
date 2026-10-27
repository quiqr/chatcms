import { useState, type FormEvent } from "react";
import { useNavigate, useSearchParams } from "react-router";

import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { whatWentWrong } from "../lib/api";
import { setPasswordWithToken } from "../lib/session";
import { say } from "../lib/say";

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
      setProblem(whatWentWrong(error, "common.cannot-reach-the-server"));
    } finally {
      setBusy(false);
    }
  }

  if (!token) {
    return (
      <main className="mx-auto flex min-h-dvh max-w-xl flex-col justify-center gap-4 p-6">
        <h1 className="text-[28px] font-semibold tracking-tight">{say("set-password.this-link-is-not-complete")}</h1>
        <p className="text-[15px] text-muted-foreground">
          {say("set-password.open-the-link-from-your-email")}
        </p>
      </main>
    );
  }

  return (
    <main className="flex min-h-dvh flex-wrap">
      <div className="flex min-w-0 shrink grow basis-[420px] items-center justify-center p-6 sm:p-8">
        <div className="w-full max-w-[380px]">
          <span className="flex items-center gap-2">
            <img src="/quiqr-logo.svg" alt={say("common.quiqr")} className="h-6 w-auto" />
            <span className="text-[15px] text-muted-foreground">{say("common.chatcms")}</span>
          </span>

          <h1 className="mt-8 text-[28px] font-semibold tracking-tight">{say("set-password.choose-a-password")}</h1>
          <p className="mt-2 text-[15px] text-muted-foreground">
            {say("set-password.then-you-are-in-you-can")}
          </p>

          <form onSubmit={onSubmit} className="mt-6 flex flex-col gap-[22px]">
            <Field
              label={say("set-password.new-password")}
              type="password"
              name="new-password"
              autoComplete="new-password"
              required
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              hint={say("set-password.at-least-twelve-characters-a-few")}
            />

            {problem && (
              <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
                {problem}
              </p>
            )}

            <Button type="submit" disabled={busy || password === ""}>
              {busy ? say("common.saving") : say("set-password.set-my-password")}
            </Button>
          </form>

          <p className="mt-6 text-[13px] text-subtle-foreground">
            {say("set-password.links-run-out-if-this-one")}
          </p>
        </div>
      </div>

      <div className="flex min-w-0 shrink grow basis-[420px] items-center justify-center bg-primary-soft p-8">
        <p className="max-w-[320px] text-[17px] text-primary-soft-foreground">
          {say("set-password.fill-in-a-form-or-just")}
        </p>
      </div>
    </main>
  );
}
