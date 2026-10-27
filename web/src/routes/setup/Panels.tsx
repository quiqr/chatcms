import { Copy } from "lucide-react";
import { useState } from "react";

import { Button } from "../../components/ui/button";
import { RadioCards } from "../../components/ui/radio-cards";
import { useAsync } from "../../lib/useAsync";
import { say } from "../../lib/say";
import {
  defaultPublishBranch,
  makeKey,
  readKey,
  readPublishing,
  setPublishing,
  testConnection,
  type Connection,
  type Publishing,
} from "../../lib/tenancy";
import { whatWentWrong } from "../../lib/api";

/**
 * These are the panels a project is set up with. The settings screen shows
 * the same ones afterwards, so they live here rather than in either screen:
 * a copy would let the two drift, and what setup decided is exactly what
 * settings changes.
 */

export function KeyPanel({
  organisation,
  project,
  onTrouble,
}: {
  organisation: string;
  project: string;
  onTrouble: (said: string) => void;
}) {
  const [key, reloadKey] = useAsync(() => readKey(organisation, project), [organisation, project]);
  const [connection, setConnection] = useState<Connection | null>(null);
  const [copied, setCopied] = useState(false);
  const [busy, setBusy] = useState(false);

  const hasKey = key.state === "ready" && key.value.publicKey !== "";

  async function run(what: () => Promise<unknown>) {
    setBusy(true);
    onTrouble("");
    try {
      await what();
    } catch (error) {
      onTrouble(whatWentWrong(error, "common.that-did-not-work-try-again"));
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
          <p className="text-[15px] text-muted-foreground">
            {key.state === "ready" ? key.value.advice : say("common.looking")}
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
                {copied ? say("panels.copied") : say("panels.copy")}
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
              {hasKey ? say("panels.make-a-new-key") : say("panels.make-a-key")}
            </Button>

            <Button
              variant="outline"
              size="compact"
              disabled={busy}
              onClick={() => void run(async () => {
                setConnection(await testConnection(organisation, project));
              })}
            >
              {say("panels.test-the-connection")}
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
    </>
  );
}

export function PublishingPanel({
  organisation,
  project,
  onTrouble,
}: {
  organisation: string;
  project: string;
  onTrouble: (said: string) => void;
}) {
  const [publishing, reloadPublishing] = useAsync(
    () => readPublishing(organisation, project),
    [organisation, project],
  );
  const [busy, setBusy] = useState(false);

  const settings: Publishing =
    publishing.state === "ready" ? publishing.value : { to: "main-line", form: "source" };

  async function run(what: () => Promise<unknown>) {
    setBusy(true);
    onTrouble("");
    try {
      await what();
    } catch (error) {
      onTrouble(whatWentWrong(error, "common.that-did-not-work-try-again"));
    } finally {
      setBusy(false);
    }
  }

  void busy;

  return (
    <>
          <RadioCards
            label={say("panels.where-to-publish")}
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
              { value: "main-line", title: say("panels.the-site-s-main-line"), description: say("panels.work-goes-straight-to-where-the-site-lives") },
              {
                value: "its-own-line",
                title: say("panels.a-line-of-its-own"),
                description: `Work goes to ${settings.branch ?? defaultPublishBranch}, for review before it goes live.`,
              },
            ]}
          />

          <div className="mt-4">
            <RadioCards
              label={say("panels.what-to-publish")}
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
                { value: "source", title: say("panels.the-site-s-files"), description: say("panels.your-own-pipeline-builds-and-puts-it-live") },
                { value: "built", title: say("panels.the-built-site"), description: say("panels.chatcms-builds-it-with-hugo-and-publishes-the-re") },
              ]}
            />
          </div>

          <p className="mt-3 text-[13px] text-subtle-foreground">
            {say("panels.steps-of-your-own-will-live")}
          </p>
    </>
  );
}
