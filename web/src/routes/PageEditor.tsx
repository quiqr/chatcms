import { useCallback, useEffect, useRef, useState } from "react";
import { useParams } from "react-router";

import { Button } from "../components/ui/button";
import { BodyEditor } from "../fields/BodyEditor";
import { Draw, withPicking } from "../fields/Fields";
import { MediaPage } from "./MediaPage";
import { readPage, savePage, type Control, type Page } from "../lib/tenancy";

type Saving = "saved" | "saving" | "trouble";

const quietFor = 800;

/**
 * withValue replaces one control's value wherever it sits, so that what is
 * drawn always comes from one place rather than from a copy that can drift.
 */
function withValue(controls: Control[], path: string, value: unknown): Control[] {
  return controls.map((control) => {
    if (control.path === path) {
      return { ...control, value };
    }
    if (control.fields && control.type !== "accordion") {
      return { ...control, fields: withValue(control.fields, path, value) };
    }
    return control;
  });
}

function Status({ state }: { state: Saving }) {
  const said = {
    saved: "All changes saved",
    saving: "Saving…",
    trouble: "Not saved yet. We will keep trying.",
  }[state];

  return (
    <p aria-live="polite" className="text-sm text-subtle-foreground">
      {said}
    </p>
  );
}

export function PageEditor({ organisation }: { organisation: string }) {
  const { project = "", page = "" } = useParams();
  const [loaded, setLoaded] = useState<Page | null>(null);
  const [trouble, setTrouble] = useState("");
  const [state, setState] = useState<Saving>("saved");
  // choosing holds what to do with the picture somebody picks, which is how
  // a field reaches the library without the library knowing about fields.
  const [choosing, setChoosing] = useState<((path: string) => void) | null>(null);

  // waiting holds what has changed but has not been written yet, so that
  // several quick edits become one save rather than one each.
  const waiting = useRef<{ values: Record<string, unknown>; body?: string }>({ values: {} });
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    let cancelled = false;
    readPage(organisation, project, page)
      .then((answer) => {
        if (!cancelled) setLoaded(answer);
      })
      .catch((error: unknown) => {
        if (!cancelled) setTrouble(error instanceof Error ? error.message : "We could not open this page.");
      });
    return () => {
      cancelled = true;
    };
  }, [organisation, project, page]);

  const write = useCallback(async () => {
    const pending = waiting.current;
    waiting.current = { values: {} };
    if (Object.keys(pending.values).length === 0 && pending.body === undefined) {
      return;
    }

    setState("saving");
    try {
      await savePage(organisation, project, page, {
        values: pending.values,
        body: pending.body,
      });
      setState("saved");
    } catch (error: unknown) {
      setState("trouble");
      setTrouble(error instanceof Error ? error.message : "We could not save that.");
    }
  }, [organisation, project, page]);

  const later = useCallback(() => {
    if (timer.current) clearTimeout(timer.current);
    timer.current = setTimeout(() => void write(), quietFor);
  }, [write]);

  // What is typed last still has to reach the server, so anything waiting is
  // written when the screen goes away.
  useEffect(() => {
    return () => {
      if (timer.current) clearTimeout(timer.current);
      void write();
    };
  }, [write]);

  function change(path: string, value: unknown) {
    setLoaded((held) => (held ? { ...held, fields: withValue(held.fields, path, value) } : held));
    waiting.current.values[path] = value;
    later();
  }

  function changeBody(body: string) {
    setLoaded((held) => (held ? { ...held, body } : held));
    waiting.current.body = body;
    later();
  }

  if (trouble && !loaded) {
    return <p className="text-[15px]">{trouble}</p>;
  }
  if (!loaded) {
    return <p className="text-[15px] text-subtle-foreground">Opening this page…</p>;
  }

  withPicking((choose) => setChoosing(() => choose));

  if (choosing) {
    return (
      <section className="flex flex-col gap-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-lg font-semibold">Choose a picture</h2>
          <Button variant="ghost" onClick={() => setChoosing(null)}>
            Back to the page
          </Button>
        </div>
        <MediaPage
          organisation={organisation}
          picking
          onUse={(image) => {
            choosing("/" + image.path.replace(/^static\//, ""));
            setChoosing(null);
          }}
        />
      </section>
    );
  }

  return (
    <section className="flex max-w-[760px] flex-col gap-5">
      <header className="flex flex-wrap items-baseline justify-between gap-2">
        <div className="flex flex-col gap-0.5">
          <p className="text-sm text-subtle-foreground">{loaded.section}</p>
          <h2 className="text-lg font-semibold">{loaded.title}</h2>
        </div>
        <Status state={state} />
      </header>

      {loaded.fields.map((control) => (
        <Draw key={control.path} control={control} onChange={change} />
      ))}

      {loaded.hasBody ? (
        <BodyEditor value={loaded.body} onChange={changeBody} />
      ) : (
        loaded.body.trim() !== "" && (
          <div className="flex flex-col gap-1.5">
            <span className="text-sm font-medium">Page text</span>
            <p className="whitespace-pre-wrap rounded-md bg-muted p-3 text-[15px] text-subtle-foreground">
              {loaded.body}
            </p>
            <p className="text-[13px] text-subtle-foreground">
              This page&rsquo;s text is not set up to be edited here.
            </p>
          </div>
        )
      )}
    </section>
  );
}
