import { useState, type FormEvent } from "react";
import { useNavigate, useSearchParams } from "react-router";

import { Button } from "../components/ui/button";
import { Checkbox, Field } from "../components/ui/field";
import { whatWentWrong } from "../lib/api";
import { signIn } from "../lib/session";
import { say } from "../lib/say";

export function SignInPage() {
  const navigate = useNavigate();
  const [address] = useSearchParams();
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
      // Back where the session ran out, when that is where somebody came
      // from. Only within this site: an address from elsewhere in a link is
      // how a person is sent somewhere they did not mean to go.
      const then = address.get("then");
      navigate(then && then.startsWith("/") && !then.startsWith("//") ? then : "/projects");
    } catch (error) {
      setProblem(whatWentWrong(error, "common.cannot-reach-the-server"));
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="flex min-h-dvh flex-wrap">
      <div className="flex min-w-0 shrink grow basis-[420px] items-center justify-center p-6 sm:p-8">
        <div className="w-full max-w-[380px]">
          <span className="flex items-center gap-2">
            <img src="/quiqr-logo.svg" alt={say("common.quiqr")} className="h-6 w-auto" />
            <span className="text-[15px] text-muted-foreground">{say("common.chatcms")}</span>
          </span>

          <h1 className="mt-8 text-[28px] font-semibold tracking-tight">{say("sign-in.sign-in")}</h1>

          {address.get("signed-out") ? (
            // Said out loud, because a session that merely ran out ends on
            // this same screen and the two should not look alike.
            <p role="status" className="mt-2 text-[15px] text-muted-foreground">
              {say("sign-in.you-have-signed-out")}
            </p>
          ) : (
            <p className="mt-2 text-[15px] text-muted-foreground">
              {say("sign-in.edit-and-publish-your-website-no")}
            </p>
          )}

          <form onSubmit={onSubmit} className="mt-6 flex flex-col gap-[22px]">
            <Field
              label={say("common.email")}
              type="email"
              name="email"
              autoComplete="email"
              required
              value={email}
              onChange={(event) => setEmail(event.target.value)}
            />

            <Field
              label={say("sign-in.password")}
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
                  {say("sign-in.forgot-password")}
                </a>
              }
            />

            <Checkbox
              label={say("sign-in.keep-me-signed-in-on-this")}
              checked={remember}
              onChange={(event) => setRemember(event.target.checked)}
            />

            {problem && (
              <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
                {problem}
              </p>
            )}

            <Button type="submit" disabled={busy}>
              {busy ? say("sign-in.signing-in") : say("sign-in.sign-in")}
            </Button>
          </form>

          <p className="mt-6 text-[13px] text-subtle-foreground">
            {say("sign-in.got-an-invitation-email-use-the")}
          </p>

          <p className="mt-10 flex flex-wrap gap-4 text-[13px] text-subtle-foreground">
            <a className="inline-flex min-h-11 items-center" href="https://chatcms.eu">
              {say("sign-in.chatcms-eu")}
            </a>
            <a className="inline-flex min-h-11 items-center" href="/privacy">
              {say("sign-in.privacy")}
            </a>
            <a className="inline-flex min-h-11 items-center" href="/help">
              {say("common.help")}
            </a>
          </p>
        </div>
      </div>

      <div className="flex min-w-0 shrink grow basis-[420px] items-center justify-center bg-primary-soft p-8">
        <p className="max-w-[320px] text-[17px] text-primary-soft-foreground">
          {say("sign-in.fill-in-a-form-or-just")}
        </p>
      </div>
    </main>
  );
}
