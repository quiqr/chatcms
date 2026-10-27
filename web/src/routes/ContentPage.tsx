import { Copy, FileText, MoreHorizontal, Pencil, Plus, Trash2 } from "lucide-react";
import { useState } from "react";
import { useNavigate, useParams } from "react-router";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { useAsync } from "../lib/useAsync";
import { onADay } from "../lib/when";
import { useHere } from "../lib/whoami";
import {
  createPage,
  deletePage,
  readContent,
  renamePage,
  type ContentItem,
  type ContentSection,
} from "../lib/tenancy";

const atATime = 20;

type Asking =
  | { what: "new"; section: string; itemName: string }
  | { what: "rename" | "copy"; page: string; title: string }
  | { what: "delete"; page: string; title: string };

function Ask({
  asking,
  busy,
  onDo,
  onCancel,
}: {
  asking: Asking;
  busy: boolean;
  onDo: (title: string) => void;
  onCancel: () => void;
}) {
  const [title, setTitle] = useState(asking.what === "new" ? "" : asking.title);

  if (asking.what === "delete") {
    return (
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="ask-title"
        className="flex flex-col gap-4 rounded-lg border border-border bg-card p-4"
      >
        <h3 id="ask-title" className="text-base font-semibold">
          Remove &ldquo;{asking.title}&rdquo;?
        </h3>
        <p className="text-[15px]">
          It will go from your site the next time you publish. Until then nothing changes for
          anybody reading it, and you can bring it back from your history.
        </p>
        <div className="flex flex-wrap gap-2">
          <Button onClick={() => onDo(asking.title)} disabled={busy}>
            {busy ? "Removing…" : "Yes, remove it"}
          </Button>
          <Button variant="ghost" onClick={onCancel} disabled={busy}>
            Keep it
          </Button>
        </div>
      </div>
    );
  }

  const heading =
    asking.what === "new"
      ? `New ${asking.itemName}`
      : asking.what === "copy"
        ? `Make a copy of “${asking.title}”`
        : `Rename “${asking.title}”`;

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="ask-title"
      className="flex flex-col gap-4 rounded-lg border border-border bg-card p-4"
    >
      <h3 id="ask-title" className="text-base font-semibold">
        {heading}
      </h3>
      <Field
        label="Name"
        hint="This is what it is called, and what its address is made from."
        type="text"
        value={title}
        autoFocus
        onChange={(event) => setTitle(event.target.value)}
      />
      <div className="flex flex-wrap gap-2">
        <Button onClick={() => onDo(title)} disabled={busy || title.trim() === ""}>
          {busy ? "Working…" : asking.what === "new" ? "Make it" : "Save the name"}
        </Button>
        <Button variant="ghost" onClick={onCancel} disabled={busy}>
          Cancel
        </Button>
      </div>
    </div>
  );
}

function Row({
  item,
  onOpen,
  onAsk,
  canChange,
}: {
  item: ContentItem;
  onOpen: () => void;
  onAsk: (asking: Asking) => void;
  canChange: boolean;
}) {
  const [menu, setMenu] = useState(false);

  return (
    <li className="flex flex-wrap items-center gap-2 border-b border-border py-2 last:border-b-0">
      <FileText aria-hidden="true" className="size-4 shrink-0 text-subtle-foreground" />
      <button
        type="button"
        onClick={onOpen}
        className="flex min-h-11 flex-1 items-center text-left text-[15px] font-medium"
      >
        {item.title}
      </button>
      {item.waiting && <Badge tone="changed">Not published</Badge>}
      {item.by && (
        <span className="text-sm text-subtle-foreground">
          {item.by}
          {onADay(item.at) ? ` · ${onADay(item.at)}` : ""}
        </span>
      )}
      {canChange && (
        <div className="relative">
          <button
            type="button"
            aria-label={`What to do with ${item.title}`}
            aria-expanded={menu}
            onClick={() => setMenu(!menu)}
            className="flex size-11 items-center justify-center rounded-md text-subtle-foreground hover:bg-muted"
          >
            <MoreHorizontal aria-hidden="true" className="size-4" />
          </button>
          {menu && (
            <div className="absolute right-0 z-10 mt-1 flex w-48 flex-col rounded-md border border-border bg-card p-1 shadow-md">
              <button
                type="button"
                onClick={() => {
                  setMenu(false);
                  onAsk({ what: "rename", page: item.key, title: item.title });
                }}
                className="flex min-h-11 items-center gap-2 rounded px-2 text-left text-[15px] hover:bg-muted"
              >
                <Pencil aria-hidden="true" className="size-4" />
                Rename
              </button>
              <button
                type="button"
                onClick={() => {
                  setMenu(false);
                  onAsk({ what: "copy", page: item.key, title: item.title });
                }}
                className="flex min-h-11 items-center gap-2 rounded px-2 text-left text-[15px] hover:bg-muted"
              >
                <Copy aria-hidden="true" className="size-4" />
                Make a copy
              </button>
              <button
                type="button"
                onClick={() => {
                  setMenu(false);
                  onAsk({ what: "delete", page: item.key, title: item.title });
                }}
                className="flex min-h-11 items-center gap-2 rounded px-2 text-left text-[15px] text-destructive-foreground hover:bg-muted"
              >
                <Trash2 aria-hidden="true" className="size-4" />
                Remove…
              </button>
            </div>
          )}
        </div>
      )}
    </li>
  );
}

export function ContentPage() {
  const { organisation } = useHere();
  const { project = "" } = useParams();
  const goTo = useNavigate();
  const [refresh, setRefresh] = useState(0);
  const [search, setSearch] = useState("");
  const [showing, setShowing] = useState<Record<string, number>>({});
  const [asking, setAsking] = useState<Asking | null>(null);
  const [busy, setBusy] = useState(false);
  const [trouble, setTrouble] = useState("");

  const [content] = useAsync<{ sections: ContentSection[] }>(
    () => readContent(organisation, project),
    [organisation, project, refresh],
  );

  async function act(title: string) {
    if (!asking) return;
    setBusy(true);
    setTrouble("");
    try {
      if (asking.what === "new") {
        const made = await createPage(organisation, project, asking.section, title);
        goTo(`/o/${organisation}/p/${project}/content/${made.key}`);
        return;
      }
      if (asking.what === "delete") {
        await deletePage(organisation, project, asking.page);
      } else {
        await renamePage(organisation, project, asking.page, title, asking.what === "copy");
      }
      setAsking(null);
      setRefresh(refresh + 1);
    } catch (error: unknown) {
      setTrouble(error instanceof Error ? error.message : "That did not work.");
    } finally {
      setBusy(false);
    }
  }

  if (content.state === "loading") {
    return <p className="text-[15px] text-subtle-foreground">Looking at what is here…</p>;
  }
  if (content.state === "failed") {
    return <p className="text-[15px]">{content.message}</p>;
  }

  const looking = search.trim().toLowerCase();

  return (
    <section className="flex flex-col gap-5">
      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      {asking && <Ask asking={asking} busy={busy} onDo={act} onCancel={() => setAsking(null)} />}

      <Field
        label="Search"
        type="search"
        value={search}
        placeholder="Find a page"
        onChange={(event) => setSearch(event.target.value)}
      />

      {content.value.sections.map((section) => {
        const matching = section.items.filter((item) =>
          looking === "" ? true : item.title.toLowerCase().includes(looking),
        );
        const limit = showing[section.key] ?? atATime;
        const shown = matching.slice(0, limit);

        return (
          <div key={section.key} className="flex flex-col gap-2">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h3 className="text-[15px] font-semibold">
                {section.title}
                <span className="ml-2 text-sm font-normal text-subtle-foreground">
                  {section.count}
                </span>
              </h3>
              {section.itemName && (
                <Button
                  variant="ghost"
                  onClick={() =>
                    setAsking({ what: "new", section: section.key, itemName: section.itemName! })
                  }
                >
                  <Plus aria-hidden="true" className="size-4" />
                  New {section.itemName}
                </Button>
              )}
            </div>

            {matching.length === 0 ? (
              <p className="text-[15px] text-subtle-foreground">
                {looking === "" ? "Nothing here yet." : "Nothing here matches that."}
              </p>
            ) : (
              <ul className="flex flex-col">
                {shown.map((item) => (
                  <Row
                    key={item.key}
                    item={item}
                    canChange={section.key !== "pages"}
                    onOpen={() => goTo(`/o/${organisation}/p/${project}/content/${item.key}`)}
                    onAsk={setAsking}
                  />
                ))}
              </ul>
            )}

            {matching.length > shown.length && (
              <button
                type="button"
                onClick={() => setShowing({ ...showing, [section.key]: limit + atATime })}
                className="flex min-h-11 w-fit items-center text-sm font-medium text-primary"
              >
                Show {Math.min(atATime, matching.length - shown.length)} more
              </button>
            )}
          </div>
        );
      })}
    </section>
  );
}
