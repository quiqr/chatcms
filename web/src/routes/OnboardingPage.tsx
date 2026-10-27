import { Check } from "lucide-react";
import { useState } from "react";
import { useParams } from "react-router";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { Segmented } from "../components/ui/segmented";
import { useAsync } from "../lib/useAsync";
import { KeyPanel, PublishingPanel } from "./setup/Panels";
import { useHere } from "../lib/whoami";
import { say } from "../lib/say";
import {
  importSite,
  readKey,
  readPublishing,
  readSite,
  type SiteReport,
} from "../lib/tenancy";
import { whatWentWrong } from "../lib/api";

type Step = { key: string; label: string; done: boolean };

function Steps({ steps, current }: { steps: Step[]; current: string }) {
  return (
    <ol className="flex flex-col gap-1">
      {steps.map((step) => (
        <li
          key={step.key}
          aria-current={step.key === current ? "step" : undefined}
          className={
            "flex min-h-11 items-center gap-2.5 rounded-lg px-3 text-[15px] " +
            (step.key === current ? "bg-muted font-medium" : "text-muted-foreground")
          }
        >
          <span
            aria-hidden="true"
            className={
              "inline-flex size-5 items-center justify-center rounded-full text-xs " +
              (step.done ? "bg-success-soft text-success-foreground" : "bg-muted")
            }
          >
            {step.done ? <Check className="size-3" /> : ""}
          </span>
          {step.label}
          {step.done && <span className="sr-only">{say("onboarding.done")}</span>}
        </li>
      ))}
    </ol>
  );
}

function Section({
  title,
  active,
  children,
}: {
  title: string;
  active?: boolean;
  children: React.ReactNode;
}) {
  return (
    <section
      aria-label={title}
      className={
        "rounded-xl border bg-card p-5 " +
        (active ? "border-[1.5px] border-primary ring-[3px] ring-primary-soft" : "border-border")
      }
    >
      <h3 className="text-[15px] font-semibold">{title}</h3>
      <div className="mt-3">{children}</div>
    </section>
  );
}

function siteSummary(report: SiteReport) {
  if (report.kind === "none") return say("onboarding.nothing-has-been-brought-in-yet");
  if (report.ready) {
    return `Hugo site, Quiqr model found: ${report.pages} pages and ${report.collections} collections.`;
  }
  return report.problem ?? "Something is missing.";
}

export function OnboardingPage() {
  const { organisation } = useHere();
  const { project = "" } = useParams();

  const [site, reloadSite] = useAsync(() => readSite(organisation, project), [organisation, project]);
  // Read here only to say which step is done; the panels load their own.
  const [key] = useAsync(() => readKey(organisation, project), [organisation, project]);
  const [publishing] = useAsync(
    () => readPublishing(organisation, project),
    [organisation, project],
  );

  const [source, setSource] = useState("repository");
  const [address, setAddress] = useState("");
  const [problem, setProblem] = useState("");
  const [busy, setBusy] = useState(false);

  async function run(what: () => Promise<unknown>) {
    setProblem("");
    setBusy(true);
    try {
      await what();
    } catch (error) {
      setProblem(whatWentWrong(error, "common.that-did-not-work-try-again"));
    } finally {
      setBusy(false);
    }
  }

  const report = site.state === "ready" ? site.value : undefined;
  const hasKey = key.state === "ready" && key.value.publicKey !== "";
  const current = !report || report.kind === "none" ? "site" : !hasKey ? "key" : "publishing";

  const steps: Step[] = [
    { key: "site", label: say("onboarding.where-the-site-comes-from"), done: Boolean(report && report.kind !== "none") },
    { key: "key", label: say("onboarding.key-for-the-git-host"), done: hasKey },
    { key: "publishing", label: say("project-settings.how-it-is-published"), done: publishing.state === "ready" },
    { key: "people", label: say("onboarding.invite-people"), done: false },
  ];

  return (
    <div className="flex flex-wrap items-start gap-6">
      <nav aria-label={say("onboarding.setting-up-this-project")} className="min-w-0 shrink grow basis-[220px]">
        <Steps steps={steps} current={current} />
      </nav>

      <div className="flex min-w-0 shrink grow-[999] basis-[520px] flex-col gap-4">
        {problem && (
          <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
            {problem}
          </p>
        )}

        <Section title={say("onboarding.where-the-site-comes-from")} active={current === "site"}>
          <Segmented
            label={say("onboarding.where-the-site-comes-from")}
            value={source}
            onValueChange={setSource}
            options={[
              { value: "repository", label: say("onboarding.a-site-already-in-git") },
              { value: "export", label: say("onboarding.from-quiqr-desktop") },
              { value: "template", label: say("onboarding.start-from-a-template") },
            ]}
          />

          <form
            className="mt-4 flex flex-wrap items-end gap-3"
            onSubmit={(event) => {
              event.preventDefault();
              void run(async () => {
                await importSite(organisation, project, { kind: source, address });
                reloadSite();
              });
            }}
          >
            <div className="min-w-[240px] flex-1">
              <Field
                label={
                  source === "repository"
                    ? say("onboarding.address-of-the-site-in-git")
                    : source === "export"
                      ? say("onboarding.folder-holding-the-exported-site")
                      : say("onboarding.name-of-the-template")
                }
                className="font-mono text-sm"
                required
                value={address}
                onChange={(event) => setAddress(event.target.value)}
              />
            </div>
            <Button type="submit" variant="ink" disabled={busy}>
              {say("onboarding.bring-it-in")}
            </Button>
          </form>

          <p className="mt-3 text-[13px] text-subtle-foreground">
            {site.state === "loading" ? say("common.looking") : report ? siteSummary(report) : ""}
          </p>

          {report?.ready && <Badge tone="live">{say("onboarding.ready-to-edit")}</Badge>}
        </Section>

        <Section title={say("onboarding.key-for-the-git-host")} active={current === "key"}>
          <KeyPanel organisation={organisation} project={project} onTrouble={setProblem} />
        </Section>

        <Section title={say("onboarding.how-it-is-published")} active={current === "publishing"}>
          <PublishingPanel organisation={organisation} project={project} onTrouble={setProblem} />
        </Section>

      </div>
    </div>
  );
}
