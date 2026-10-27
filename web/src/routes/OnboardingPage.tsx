import { Check, Copy } from "lucide-react";
import { useState } from "react";
import { useParams } from "react-router";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { RadioCards } from "../components/ui/radio-cards";
import { Segmented } from "../components/ui/segmented";
import { useAsync } from "../lib/useAsync";
import { useHere } from "../lib/whoami";
import {
  defaultPublishBranch,
  importSite,
  makeKey,
  readKey,
  readPublishing,
  readSite,
  setPublishing,
  testConnection,
  type Connection,
  type Publishing,
  type SiteReport,
} from "../lib/tenancy";

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
          {step.done && <span className="sr-only">done</span>}
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
  if (report.kind === "none") return "Nothing has been brought in yet.";
  if (report.ready) {
    return `Hugo site, Quiqr model found: ${report.pages} pages and ${report.collections} collections.`;
  }
  return report.problem ?? "Something is missing.";
}

export function OnboardingPage() {
  const { organisation } = useHere();
  const { project = "" } = useParams();

  const [site, reloadSite] = useAsync(() => readSite(organisation, project), [organisation, project]);
  const [key, reloadKey] = useAsync(() => readKey(organisation, project), [organisation, project]);
  const [publishing, reloadPublishing] = useAsync(
    () => readPublishing(organisation, project),
    [organisation, project],
  );

  const [source, setSource] = useState("repository");
  const [address, setAddress] = useState("");
  const [problem, setProblem] = useState("");
  const [busy, setBusy] = useState(false);
  const [connection, setConnection] = useState<Connection | null>(null);
  const [copied, setCopied] = useState(false);

  async function run(what: () => Promise<unknown>) {
    setProblem("");
    setBusy(true);
    try {
      await what();
    } catch (error) {
      setProblem(error instanceof Error ? error.message : "That did not work. Please try again.");
    } finally {
      setBusy(false);
    }
  }

  const report = site.state === "ready" ? site.value : undefined;
  const hasKey = key.state === "ready" && key.value.publicKey !== "";
  const current = !report || report.kind === "none" ? "site" : !hasKey ? "key" : "publishing";

  const steps: Step[] = [
    { key: "site", label: "Where the site comes from", done: Boolean(report && report.kind !== "none") },
    { key: "key", label: "Key for the git host", done: hasKey },
    { key: "publishing", label: "How it is published", done: publishing.state === "ready" },
    { key: "people", label: "Invite people", done: false },
  ];

  const settings: Publishing =
    publishing.state === "ready" ? publishing.value : { to: "main-line", form: "source" };

  return (
    <div className="flex flex-wrap items-start gap-6">
      <nav aria-label="Setting up this project" className="min-w-0 shrink grow basis-[220px]">
        <Steps steps={steps} current={current} />
      </nav>

      <div className="flex min-w-0 shrink grow-[999] basis-[520px] flex-col gap-4">
        {problem && (
          <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
            {problem}
          </p>
        )}

        <Section title="Where the site comes from" active={current === "site"}>
          <Segmented
            label="Where the site comes from"
            value={source}
            onValueChange={setSource}
            options={[
              { value: "repository", label: "A site already in git" },
              { value: "export", label: "From Quiqr Desktop" },
              { value: "template", label: "Start from a template" },
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
                    ? "Address of the site in git"
                    : source === "export"
                      ? "Folder holding the exported site"
                      : "Name of the template"
                }
                className="font-mono text-sm"
                required
                value={address}
                onChange={(event) => setAddress(event.target.value)}
              />
            </div>
            <Button type="submit" variant="ink" disabled={busy}>
              Bring it in
            </Button>
          </form>

          <p className="mt-3 text-[13px] text-subtle-foreground">
            {site.state === "loading" ? "Looking..." : report ? siteSummary(report) : ""}
          </p>

          {report?.ready && <Badge tone="live">Ready to edit</Badge>}
        </Section>

        <Section title="Key for the git host" active={current === "key"}>
          <p className="text-[15px] text-muted-foreground">
            {key.state === "ready" ? key.value.advice : "Looking..."}
          </p>

          {hasKey && key.state === "ready" && (
            <div className="mt-3 flex flex-wrap items-center gap-2">
              <code className="min-w-0 flex-1 overflow-x-auto rounded-md border border-border bg-muted px-3 py-2 font-mono text-xs">
                {key.value.publicKey}
              </code>
              <Button
                variant="outline"
                size="compact"
                onClick={() => {
                  void navigator.clipboard?.writeText(key.value.publicKey);
                  setCopied(true);
                }}
              >
                <Copy aria-hidden="true" className="size-4" />
                {copied ? "Copied" : "Copy"}
              </Button>
            </div>
          )}

          <div className="mt-3 flex flex-wrap gap-2">
            <Button
              variant="outline"
              size="compact"
              disabled={busy}
              onClick={() => void run(async () => {
                await makeKey(organisation, project);
                reloadKey();
              })}
            >
              {hasKey ? "Make a new key" : "Make a key"}
            </Button>

            <Button
              variant="outline"
              size="compact"
              disabled={busy}
              onClick={() => void run(async () => {
                setConnection(await testConnection(organisation, project));
              })}
            >
              Test the connection
            </Button>
          </div>

          {connection && (
            <p
              role="status"
              className={
                "mt-3 rounded-md px-3 py-2 text-sm " +
                (connection.good
                  ? "bg-success-soft text-success-foreground"
                  : "bg-warning-soft text-warning-foreground")
              }
            >
              {connection.message}
            </p>
          )}
        </Section>

        <Section title="How it is published" active={current === "publishing"}>
          <RadioCards
            label="Where to publish"
            value={settings.to}
            onValueChange={(to) =>
              void run(async () => {
                await setPublishing(organisation, project, {
                  ...settings,
                  to: to as Publishing["to"],
                  branch: to === "its-own-line" ? (settings.branch ?? defaultPublishBranch) : undefined,
                });
                reloadPublishing();
              })
            }
            options={[
              { value: "main-line", title: "The site's main line", description: "Work goes straight to where the site lives." },
              {
                value: "its-own-line",
                title: "A line of its own",
                description: `Work goes to ${settings.branch ?? defaultPublishBranch}, for review before it goes live.`,
              },
            ]}
          />

          <div className="mt-4">
            <RadioCards
              label="What to publish"
              value={settings.form}
              onValueChange={(form) =>
                void run(async () => {
                  await setPublishing(organisation, project, {
                    ...settings,
                    form: form as Publishing["form"],
                  });
                  reloadPublishing();
                })
              }
              options={[
                { value: "source", title: "The site's files", description: "Your own pipeline builds and puts it live." },
                { value: "built", title: "The built site", description: "ChatCMS builds it with Hugo and publishes the result." },
              ]}
            />
          </div>

          <p className="mt-3 text-[13px] text-subtle-foreground">
            Steps of your own will live in quiqr/workflows. An editor for them comes later.
          </p>
        </Section>
      </div>
    </div>
  );
}
