import { AlertTriangle, CloudOff, RefreshCw } from "lucide-react";
import { useState } from "react";
import { useParams } from "react-router";

import { Button } from "../components/ui/button";
import { useAsync } from "../lib/useAsync";
import { useHere } from "../lib/whoami";
import { plural, say } from "../lib/say";
import {
  abandonChanges,
  answerChanges,
  bringChanges,
  readStanding,
  type Answer,
  type FieldConflict,
  type PageConflict,
  type Standing,
  type Word,
} from "../lib/tenancy";
import { whatWentWrong } from "../lib/api";

function Words({ words }: { words: Word[] }) {
  return (
    <p className="text-[15px] leading-relaxed">
      {words.map((word, index) => {
        if (word.part === "gone") {
          return (
            <del key={index} className="bg-destructive-soft text-destructive-foreground line-through decoration-1">
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

function Question({
  page,
  field,
  who,
  chosen,
  onChoose,
}: {
  page: PageConflict;
  field: FieldConflict;
  who: string;
  chosen: "mine" | "theirs";
  onChoose: (choice: "mine" | "theirs") => void;
}) {
  const [comparing, setComparing] = useState(false);

  return (
    <li className="flex flex-col gap-2 rounded-md border border-border p-3">
      <p className="text-[15px]">
        {say("sync.also-changed-this-field", { who, field: field.name.toLowerCase() })}
      </p>
      <div className="flex flex-wrap gap-2">
        <Button variant={chosen === "mine" ? "primary" : "ghost"} onClick={() => onChoose("mine")}>
          {say("sync.keep-mine")}
        </Button>
        <Button variant={chosen === "theirs" ? "primary" : "ghost"} onClick={() => onChoose("theirs")}>
          {say("sync.use-theirs-named", { who })}
        </Button>
        <Button variant="ghost" onClick={() => setComparing(!comparing)}>
          {comparing ? say("sync.hide") : say("sync.compare")}
        </Button>
      </div>
      {comparing && (
        <div className="flex flex-col gap-2">
          {field.words && field.words.length > 0 ? (
            <Words words={field.words} />
          ) : (
            <>
              <p className="text-[15px]">
                <span className="text-sm text-subtle-foreground">{say("sync.mine")} </span>
                {field.mine || say("sync.nothing")}
              </p>
              <p className="text-[15px]">
                <span className="text-sm text-subtle-foreground">{who}: </span>
                {field.theirs || "nothing"}
              </p>
            </>
          )}
        </div>
      )}
      <span className="sr-only">{`${page.title}, ${field.name}`}</span>
    </li>
  );
}

export function SyncPage() {
  const { organisation } = useHere();
  const { project = "" } = useParams();
  const [choices, setChoices] = useState<Record<string, Record<string, "mine" | "theirs">>>({});
  const [busy, setBusy] = useState(false);
  const [trouble, setTrouble] = useState("");
  const [showTechnical, setShowTechnical] = useState(false);

  const [standing] = useAsync<Standing>(
    () => readStanding(organisation, project),
    [organisation, project],
  );

  // What an action answered. Asking again afterwards would lose it: by then
  // the work has arrived and the site is simply saved, so the person would
  // never be told that somebody else had changed it.
  const [after, setAfter] = useState<Standing | null>(null);

  async function run(what: () => Promise<Standing>) {
    setBusy(true);
    setTrouble("");
    try {
      setAfter(await what());
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error));
    } finally {
      setBusy(false);
    }
  }

  function answers(conflicts: PageConflict[]): Answer[] {
    return conflicts.map((page) => {
      if (page.wholeFile) {
        return { file: page.file, whole: choices[page.file]?.["whole"] ?? "mine" };
      }
      return { file: page.file, fields: choices[page.file] ?? {} };
    });
  }

  if (standing.state === "loading") {
    return <p className="text-[15px] text-subtle-foreground">{say("sync.checking")}</p>;
  }
  if (standing.state === "failed") {
    return <p className="text-[15px]">{standing.message}</p>;
  }

  const now = after ?? standing.value;
  const conflicts = now.conflicts ?? [];
  const who = (now.arrivedFrom ?? []).join(" and ") || "Someone else";

  return (
    <section className="flex flex-col gap-5">
      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      <div className="flex flex-wrap items-center gap-3">
        {now.state === "offline" ? (
          <CloudOff aria-hidden="true" className="size-5 text-warning-foreground" />
        ) : now.state === "asking" ? (
          <AlertTriangle aria-hidden="true" className="size-5 text-warning-foreground" />
        ) : (
          <RefreshCw aria-hidden="true" className={"size-5 " + (now.state === "saving" ? "animate-spin" : "")} />
        )}
        <p aria-live="polite" className="flex-1 text-[15px]">
          {now.said}
        </p>
        <Button variant="ghost" onClick={() => run(() => bringChanges(organisation, project))} disabled={busy}>
          {say("sync.check-for-updates")}
        </Button>
      </div>

      {now.unpublished > 0 && (
        <p className="text-[15px] text-subtle-foreground">
          {plural(now.unpublished, "sync.waiting-one", "sync.waiting-many")}
        </p>
      )}

      {(now.arrivedFrom ?? []).length > 0 && now.state === "arrived" && (
        <p className="rounded-md bg-info-soft px-3 py-2 text-[15px] text-info-foreground">
          {say("sync.updated-this-site", { who })}
        </p>
      )}

      {conflicts.map((page) => (
        <div key={page.file} className="flex flex-col gap-2">
          <h3 className="text-[15px] font-semibold">{page.title}</h3>
          {page.wholeFile ? (
            <div className="flex flex-col gap-2 rounded-md border border-border p-3">
              <p className="text-[15px]">
                {say("sync.also-changed-whole-page", {
                  who: page.who || say("sync.someone-else"),
                })}
              </p>
              <div className="flex flex-wrap gap-2">
                <Button
                  variant={(choices[page.file]?.["whole"] ?? "mine") === "mine" ? "primary" : "ghost"}
                  onClick={() => setChoices({ ...choices, [page.file]: { whole: "mine" } })}
                >
                  {say("sync.keep-mine")}
                </Button>
                <Button
                  variant={choices[page.file]?.["whole"] === "theirs" ? "primary" : "ghost"}
                  onClick={() => setChoices({ ...choices, [page.file]: { whole: "theirs" } })}
                >
                  {say("sync.use-theirs")}
                </Button>
              </div>
            </div>
          ) : (
            <ul className="flex flex-col gap-2">
              {(page.fields ?? []).map((field) => (
                <Question
                  key={field.path}
                  page={page}
                  field={field}
                  who={page.who || "Someone else"}
                  chosen={choices[page.file]?.[field.path] ?? "mine"}
                  onChoose={(choice) =>
                    setChoices({
                      ...choices,
                      [page.file]: { ...(choices[page.file] ?? {}), [field.path]: choice },
                    })
                  }
                />
              ))}
            </ul>
          )}
        </div>
      ))}

      {conflicts.length > 0 && (
        <div className="flex flex-wrap gap-2">
          <Button
            onClick={() => {
              setChoices({});
              void run(() => answerChanges(organisation, project, answers(conflicts)));
            }}
            disabled={busy}
          >
            {busy ? say("sync.saving") : say("sync.save-these-answers")}
          </Button>
          <Button variant="ghost" onClick={() => run(() => abandonChanges(organisation, project))} disabled={busy}>
            {say("sync.start-again")}
          </Button>
        </div>
      )}

      {now.technical && (
        // chatcms-audience: implementer
        <div className="flex flex-col gap-2">
          <button
            type="button"
            aria-expanded={showTechnical}
            onClick={() => setShowTechnical(!showTechnical)}
            className="flex min-h-11 w-fit items-center text-sm text-subtle-foreground"
          >
            {say("sync.technical-details-for-implementers")}
          </button>
          {showTechnical && (
            <dl className="rounded-md bg-muted p-3 font-mono text-xs">
              {/* chatcms-not-a-string: what the technical detail is called, shown as it is */}
              <div className="flex gap-2">
                <dt>branch</dt>
                <dd>{now.technical.branch}</dd>
              </div>
              {/* chatcms-not-a-string: the same */}
              <div className="flex gap-2">
                <dt>remote</dt>
                <dd className="break-all">{now.technical.remote}</dd>
              </div>
              {/* chatcms-not-a-string: the same */}
              <div className="flex gap-2">
                <dt>head</dt>
                <dd>{now.technical.head}</dd>
              </div>
            </dl>
          )}
        </div>
      )}
    </section>
  );
}
