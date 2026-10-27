import { CheckCircle2, Clock, RotateCcw } from "lucide-react";
import { useState } from "react";
import { useParams } from "react-router";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Checkbox } from "../components/ui/field";
import { useAsync } from "../lib/useAsync";
import { when } from "../lib/when";
import { useHere } from "../lib/whoami";
import { plural, say } from "../lib/say";
import {
  beforeGoingBack,
  goBackTo,
  readVersion,
  readTimeline,
  type GoBackWarning,
  type Review,
  type Version,
  type VersionTimeline,
  type WentBack,
} from "../lib/tenancy";

function changeCount(count: number) {
  if (count === 0) return say("common.no-changes");
  return plural(count, "common.change-one", "common.change-many");
}

function Contents({ contents }: { contents: Review }) {
  if (contents.changes.length === 0) {
    return <p className="text-sm text-subtle-foreground">{say("history.nothing-was-changed-in-this-version")}</p>;
  }

  return (
    <ul className="flex flex-col gap-2">
      {contents.changes.map((change) => (
        <li key={`${change.section}/${change.page}`} className="text-sm">
          <span className="font-medium">{change.page}</span>
          <span className="text-subtle-foreground"> · {change.section}</span>
          {change.fields && change.fields.length > 0 && (
            <span className="text-subtle-foreground">
              {" "}
              · {change.fields.map((field) => field.name).join(", ")}
            </span>
          )}
        </li>
      ))}
    </ul>
  );
}

function Confirm({
  warning,
  busy,
  onGoBack,
  onCancel,
}: {
  warning: GoBackWarning;
  busy: boolean;
  onGoBack: (publishRightAway: boolean) => void;
  onCancel: () => void;
}) {
  const [publishRightAway, setPublishRightAway] = useState(true);

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="go-back-title"
      className="flex flex-col gap-4 rounded-lg border border-border bg-card p-4"
    >
      <h3 id="go-back-title" className="text-base font-semibold">
        {say("history.go-back-to-version", { title: warning.title })}
      </h3>
      <p className="text-[15px]">
        {say("history.your-site-will-look-as-it-did", {
          when: when(warning.when) || say("history.that-day"),
        })}
      </p>
      {warning.setAside > 0 && (
        <p className="rounded-md bg-warning-soft p-3 text-[15px] text-warning-foreground">
          {warning.reassurance}
        </p>
      )}
      {warning.setAside === 0 && (
        <p className="text-[15px] text-subtle-foreground">{warning.reassurance}</p>
      )}
      <Checkbox
        label={say("history.publish-right-away")}
        checked={publishRightAway}
        onChange={(event) => setPublishRightAway(event.target.checked)}
      />
      <div className="flex flex-wrap gap-2">
        <Button onClick={() => onGoBack(publishRightAway)} disabled={busy}>
          {busy ? say("history.going-back") : say("history.yes-go-back")}
        </Button>
        <Button variant="ghost" onClick={onCancel} disabled={busy}>
          {say("history.keep-things-as-they-are")}
        </Button>
      </div>
    </div>
  );
}

function Entry({
  entry,
  organisation,
  project,
  onGoBack,
}: {
  entry: Version;
  organisation: string;
  project: string;
  onGoBack: () => void;
}) {
  const [open, setOpen] = useState(false);
  const [contents] = useAsync<Review | null>(
    () => (open ? readVersion(organisation, project, entry.version) : Promise.resolve(null)),
    [organisation, project, entry.version, open],
  );

  return (
    <li className="flex flex-col gap-2 border-b border-border py-3 last:border-b-0">
      <div className="flex flex-wrap items-center gap-2.5">
        <span className="text-[15px] font-medium">{entry.title}</span>
        {entry.live && <Badge tone="live">{say("history.live-now")}</Badge>}
      </div>
      <p className="text-sm text-subtle-foreground">
        {when(entry.at)}
        {entry.by ? ` · ${entry.by}` : ""} · {changeCount(entry.changes)}
      </p>
      <div className="flex flex-wrap gap-2">
        <button
          type="button"
          aria-expanded={open}
          onClick={() => setOpen(!open)}
          className="flex min-h-11 items-center text-sm font-medium text-primary"
        >
          {open ? say("history.hide-what-is-in-it") : say("history.see-what-is-in-it")}
        </button>
        {!entry.live && (
          <button
            type="button"
            onClick={onGoBack}
            className="flex min-h-11 items-center gap-1.5 text-sm font-medium text-primary"
          >
            <RotateCcw aria-hidden="true" className="size-4" />
            {say("history.go-back-to-this-version")}
          </button>
        )}
      </div>
      {open && contents.state === "ready" && contents.value && <Contents contents={contents.value} />}
      {open && contents.state === "loading" && (
        <p className="text-sm text-subtle-foreground">{say("history.looking")}</p>
      )}
    </li>
  );
}

export function HistoryPage() {
  const { organisation } = useHere();
  const { project = "" } = useParams();
  const [refresh, setRefresh] = useState(0);
  const [going, setGoing] = useState<{ version: string; warning: GoBackWarning } | null>(null);
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState<WentBack | null>(null);

  const [timeline] = useAsync<VersionTimeline>(
    () => readTimeline(organisation, project),
    [organisation, project, refresh],
  );

  async function ask(version: string) {
    setDone(null);
    setGoing({ version, warning: await beforeGoingBack(organisation, project, version) });
  }

  async function go(publishRightAway: boolean) {
    if (!going) return;
    setBusy(true);
    try {
      setDone(await goBackTo(organisation, project, going.version, publishRightAway));
      setGoing(null);
      setRefresh(refresh + 1);
    } finally {
      setBusy(false);
    }
  }

  if (timeline.state === "loading") {
    return <p className="text-[15px] text-subtle-foreground">{say("history.looking-at-your-history")}</p>;
  }
  if (timeline.state === "failed") {
    return <p className="text-[15px]">{timeline.message}</p>;
  }

  return (
    <section className="flex flex-col gap-5">
      {done && (
        <p
          aria-live="polite"
          className="flex items-center gap-2.5 rounded-md bg-success-soft p-3 text-[15px] text-success-foreground"
        >
          <CheckCircle2 aria-hidden="true" className="size-5" />
          {done.published
            ? say("history.back-to-and-live", { title: done.title })
            : say("history.back-to-and-waiting", { title: done.title })}
        </p>
      )}

      {going && (
        <Confirm warning={going.warning} busy={busy} onGoBack={go} onCancel={() => setGoing(null)} />
      )}

      {timeline.value.unpublished > 0 && (
        <div className="flex items-center gap-2.5 rounded-lg border border-dashed border-border p-3">
          <Clock aria-hidden="true" className="size-5 text-subtle-foreground" />
          <span className="text-[15px]">
            {say("history.not-published-yet-count", { changes: changeCount(timeline.value.unpublished) })}
          </span>
        </div>
      )}

      {timeline.value.groups.length === 0 ? (
        <p className="text-[15px] text-subtle-foreground">
          {say("history.nothing-has-been-saved-on-this")}
        </p>
      ) : (
        timeline.value.groups.map((group) => (
          <div key={group.period} className="flex flex-col gap-1">
            <h3 className="text-xs font-medium uppercase tracking-wide text-subtle-foreground">
              {group.period}
            </h3>
            <ul className="flex flex-col">
              {group.entries.map((entry) => (
                <Entry
                  key={entry.version}
                  entry={entry}
                  organisation={organisation}
                  project={project}
                  onGoBack={() => ask(entry.version)}
                />
              ))}
            </ul>
          </div>
        ))
      )}
    </section>
  );
}
