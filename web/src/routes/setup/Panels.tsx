import { Copy } from "lucide-react";
import { useState } from "react";

import { Button } from "../../components/ui/button";
import { RadioCards } from "../../components/ui/radio-cards";
import { useAsync } from "../../lib/useAsync";
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
      onTrouble(error instanceof Error ? error.message : "That did not work. Please try again.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
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
      onTrouble(error instanceof Error ? error.message : "That did not work. Please try again.");
    } finally {
      setBusy(false);
    }
  }

  void busy;

  return (
    <>
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
    </>
  );
}
