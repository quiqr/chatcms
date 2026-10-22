import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router";

import { Button } from "../components/ui/button";
import { Checkbox, Field } from "../components/ui/field";
import { RequestFailed } from "../lib/api";
import { signIn } from "../lib/session";

export function SignInPage() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [remember, setRemember] = useState(false);
  const [problem, setProblem] = useState("");
  const [busy, setBusy] = useState(false);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setProblem("");
    setBusy(true);

    try {
      await signIn(email, password, remember);
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

  return (
    <main className="flex min-h-dvh flex-wrap">
      <div className="flex min-w-0 shrink grow basis-[420px] items-center justify-center p-6 sm:p-8">
        <div className="w-full max-w-[380px]">
          <span className="flex items-center gap-2">
            <img src="/quiqr-logo.svg" alt="Quiqr" className="h-6 w-auto" />
            <span className="text-[15px] text-muted-foreground">ChatCMS</span>
          </span>

          <h1 className="mt-8 text-[28px] font-semibold tracking-tight">Sign in</h1>

          <form onSubmit={onSubmit} className="mt-6 flex flex-col gap-[22px]">
            <Field
              label="Email"
              type="email"
              name="email"
              autoComplete="email"
              required
              value={email}
              onChange={(event) => setEmail(event.target.value)}
            />

            <Field
              label="Password"
              type="password"
              name="password"
              autoComplete="current-password"
              required
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              action={
                <a
                  href="/forgot-password"
                  className="-my-3 inline-flex min-h-11 items-center text-[13px] text-primary"
                >
                  Forgot password?
                </a>
              }
            />

            <Checkbox
              label="Keep me signed in on this device"
              checked={remember}
              onChange={(event) => setRemember(event.target.checked)}
            />

            {problem && (
              <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
                {problem}
              </p>
            )}

            <Button type="submit" disabled={busy}>
              {busy ? "Signing in..." : "Sign in"}
            </Button>
          </form>

          <p className="mt-6 text-[13px] text-subtle-foreground">
            Got an invitation email? Use the link in it to set your password first.
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
