import { AlertTriangle, CheckCircle2, ChevronDown, FileText, Image, Trash2 } from "lucide-react";
import { useState } from "react";
import { useParams } from "react-router";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { useAsync } from "../lib/useAsync";
import { onADay } from "../lib/when";
import { useHere } from "../lib/whoami";
import { plural, say } from "../lib/say";
import {
  publishNow,
  readReview,
  type Change,
  type ChangeKind,
  type FieldChange,
  type Publish,
  type Review,
  type Word,
} from "../lib/tenancy";

const badges: Record<ChangeKind, { label: string; tone: "changed" | "new" | "attention" | "neutral" }> = {
  "changed-page": { label: "Changed", tone: "changed" },
  "new-page": { label: say("publish.new-page"), tone: "new" },
  "removed-page": { label: "Removed", tone: "attention" },
  "new-image": { label: say("publish.new-image"), tone: "new" },
  "removed-image": { label: "Removed", tone: "attention" },
  "other-file": { label: "Changed", tone: "neutral" },
};

function iconFor(kind: ChangeKind) {
  if (kind === "new-image" || kind === "removed-image") {
    return <Image aria-hidden="true" className="size-4 text-muted-foreground" />;
  }
  if (kind === "removed-page") {
    return <Trash2 aria-hidden="true" className="size-4 text-muted-foreground" />;
  }
  return <FileText aria-hidden="true" className="size-4 text-muted-foreground" />;
}

function Words({ words }: { words: Word[] }) {
  return (
    <p className="text-[15px] leading-relaxed">
      {words.map((word, index) => {
        if (word.part === "gone") {
          return (
            <del key={index} className="bg-danger-soft text-danger-foreground line-through decoration-1">
              {word.text}
            </del>
          );
        }
        if (word.part === "added") {
          return (
            <ins key={index} className="bg-success-soft text-success-foreground no-underline">
              {word.text}
            </ins>
          );
        }
        return <span key={index}>{word.text}</span>;
      })}
    </p>
  );
}

function Fields({ fields }: { fields: FieldChange[] }) {
  return (
    <dl className="flex flex-col gap-3 border-l-2 border-border pl-3">
      {fields.map((field) => (
        <div key={field.name} className="flex flex-col gap-1">
          <dt className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            {field.name}
          </dt>
          <dd>{field.words?.length ? <Words words={field.words} /> : <p className="text-[15px]">{say("publish.changed")}</p>}</dd>
        </div>
      ))}
    </dl>
  );
}

function ChangeRow({ change }: { change: Change }) {
  const [open, setOpen] = useState(false);
  const badge = badges[change.kind];
  const fields = change.fields ?? [];

  return (
    <li className="flex flex-col gap-2 border-b border-border py-3 last:border-b-0">
      <div className="flex flex-wrap items-center gap-2.5">
        {iconFor(change.kind)}
        <Badge tone={badge.tone}>{badge.label}</Badge>
        <span className="text-[15px] font-medium">{change.page}</span>
        <span className="text-sm text-muted-foreground">
          {change.section}
          {change.by ? ` · ${change.by}` : ""}
          {onADay(change.at) ? ` · ${onADay(change.at)}` : ""}
        </span>
      </div>

      {fields.length > 0 && (
        <>
          <p className="text-sm text-muted-foreground">
            {fields.map((field) => field.name).join(", ")}
          </p>
          <button
            type="button"
            aria-expanded={open}
            onClick={() => setOpen(!open)}
            className="flex min-h-11 w-fit items-center gap-1.5 text-sm font-medium text-accent-foreground"
          >
            <ChevronDown aria-hidden="true" className={"size-4 transition-transform " + (open ? "rotate-180" : "")} />
            {open ? say("publish.hide-changes") : say("publish.show-changes")}
          </button>
          {open && <Fields fields={fields} />}
        </>
      )}
    </li>
  );
}

function Progress({ publish }: { publish: Publish }) {
  const steps = publish.steps ?? [];

  return (
    <ol className="flex flex-col gap-2">
      {steps.map((step) => (
        <li key={step.title} className="flex min-h-11 items-center gap-2.5 text-[15px]">
          {step.state === "succeeded" ? (
            <CheckCircle2 aria-hidden="true" className="size-5 text-success-foreground" />
          ) : (
            <AlertTriangle aria-hidden="true" className="size-5 text-danger-foreground" />
          )}
          {step.title}
          <span className="sr-only">
            {step.state === "succeeded" ? say("onboarding.done") : say("publish.did-not-finish")}
          </span>
        </li>
      ))}
    </ol>
  );
}

function Done({ publish, onKeepEditing }: { publish: Publish; onKeepEditing: () => void }) {
  return (
    <section className="flex flex-col gap-4" aria-live="polite">
      <div className="flex items-center gap-2.5">
        <CheckCircle2 aria-hidden="true" className="size-6 text-success-foreground" />
        <h2 className="text-lg font-semibold">{say("publish.your-site-is-up-to-date")}</h2>
      </div>
      <p className="text-[15px] text-muted-foreground">
        {publish.changeCount === 0
          ? say("publish.there-was-nothing-waiting-so-nothing")
          : plural(publish.changeCount, "publish.change-is-live-one", "publish.change-is-live-many")}
      </p>
      <Progress publish={publish} />
      <div className="flex flex-wrap gap-2">
        <Button onClick={onKeepEditing}>{say("publish.back-to-content")}</Button>
      </div>
    </section>
  );
}

function Trouble({ publish, onTryAgain }: { publish: Publish; onTryAgain: () => void }) {
  const [showDetail, setShowDetail] = useState(false);

  return (
    <section className="flex flex-col gap-4" aria-live="polite">
      <div className="flex items-center gap-2.5">
        <AlertTriangle aria-hidden="true" className="size-6 text-danger-foreground" />
        <h2 className="text-lg font-semibold">{say("publish.we-couldn-t-publish-this-time")}</h2>
      </div>
      <p className="text-[15px]">{publish.trouble}</p>
      <Progress publish={publish} />
      <div className="flex flex-wrap gap-2">
        <Button onClick={onTryAgain}>{say("publish.try-again")}</Button>
      </div>

      {publish.detail && (
        // chatcms-audience: implementer
        <div className="flex flex-col gap-2">
          <button
            type="button"
            aria-expanded={showDetail}
            onClick={() => setShowDetail(!showDetail)}
            className="flex min-h-11 w-fit items-center gap-1.5 text-sm text-muted-foreground"
          >
            <ChevronDown aria-hidden="true" className={"size-4 transition-transform " + (showDetail ? "rotate-180" : "")} />
            {say("publish.technical-details-for-implementers")}
          </button>
          {showDetail && (
            <pre className="max-h-80 overflow-auto rounded-lg bg-muted p-3 text-xs">{publish.detail}</pre>
          )}
        </div>
      )}
    </section>
  );
}

export function PublishPage() {
  const { organisation } = useHere();
  const { project = "" } = useParams();
  const [note, setNote] = useState("");
  const [publish, setPublish] = useState<Publish | null>(null);
  const [sending, setSending] = useState(false);
  const [refresh, setRefresh] = useState(0);

  const [review] = useAsync<Review>(
    () => readReview(organisation, project),
    [organisation, project, refresh],
  );

  async function send() {
    setSending(true);
    try {
      setPublish(await publishNow(organisation, project, note));
      setNote("");
    } finally {
      setSending(false);
    }
  }

  function again() {
    setPublish(null);
    setRefresh(refresh + 1);
  }

  if (publish?.state === "succeeded") {
    return <Done publish={publish} onKeepEditing={again} />;
  }
  if (publish) {
    return <Trouble publish={publish} onTryAgain={again} />;
  }

  if (review.state === "loading") {
    return <p className="text-[15px] text-muted-foreground">{say("publish.looking-at-what-changed")}</p>;
  }
  if (review.state === "failed") {
    return <p className="text-[15px]">{review.message}</p>;
  }

  const waiting = review.value.changes ?? [];
  const count = review.value.count;

  return (
    <section className="flex flex-col gap-5">
      <header className="flex flex-col gap-1">
        <h2 className="text-lg font-semibold">
          {count === 0
            ? say("publish.nothing-is-waiting")
            : `Ready to publish ${count} ${count === 1 ? "change" : "changes"}`}
        </h2>
        {review.value.lastPublished && (
          <p className="text-sm text-muted-foreground">
            {say("publish.last-published")} {onADay(review.value.lastPublished)}
          </p>
        )}
      </header>

      {count === 0 ? (
        <p className="text-[15px] text-muted-foreground">
          {say("publish.everything-you-have-written-is-already")}
        </p>
      ) : (
        <ul className="flex flex-col">
          {waiting.map((change) => (
            <ChangeRow key={`${change.section}/${change.page}`} change={change} />
          ))}
        </ul>
      )}

      <Field
        label={say("publish.what-s-in-this-update")}
        hint={say("publish.optional-people-see-this-later-in")}
        type="text"
        value={note}
        onChange={(event) => setNote(event.target.value)}
        placeholder={say("publish.spring-prices")}
      />

      <div className="sticky bottom-0 -mx-4 flex flex-col gap-2 border-t border-border bg-card px-4 py-3 sm:static sm:mx-0 sm:border-0 sm:bg-transparent sm:px-0 sm:py-0">
        <Button onClick={send} disabled={sending || count === 0} className="w-full sm:w-fit">
          {sending ? say("publish.publishing") : `Publish ${count} ${count === 1 ? "change" : "changes"}`}
        </Button>
        <p className="text-sm text-muted-foreground">{say("publish.you-can-keep-editing-while-this")}</p>
      </div>
    </section>
  );
}
