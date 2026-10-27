import { ArrowLeft, FileText, History, Image, Upload } from "lucide-react";
import type { ReactNode } from "react";
import { NavLink } from "react-router";

import { Button } from "../components/ui/button";
import { Segmented } from "../components/ui/segmented";
import { PersonMenu } from "./PersonMenu";
import { AssistantSlot } from "./ProjectShell";
import { useHere } from "../lib/whoami";
import { SyncIndicator } from "./SyncIndicator";
import type { Project } from "../lib/tenancy";
import { say } from "../lib/say";

export type EditorLayout = "form" | "split" | "preview";

export type EditorShellProps = {
  project: Project;
  viewer: { name: string };
  breadcrumb: string;
  layout: EditorLayout;
  onLayoutChange: (layout: EditorLayout) => void;
  form: ReactNode;
  preview: ReactNode;
};

const railEntries = [
  { key: "content", label: "Content", icon: FileText },
  { key: "media", label: "Media", icon: Image },
  { key: "publish", label: "Publish", icon: Upload },
  { key: "history", label: "History", icon: History },
];

export function EditorShell({
  project,
  viewer,
  breadcrumb,
  layout,
  onLayoutChange,
  form,
  preview,
}: EditorShellProps) {
  const { organisation } = useHere();
  return (
    <div data-testid="editor-shell" className="flex min-h-dvh flex-wrap">
      <nav
        aria-label={say("editor-shell.project-sections")}
        data-testid="editor-rail"
        className="hidden w-16 shrink-0 flex-col items-center gap-2 border-r border-border bg-card py-3 sm:flex"
      >
        {railEntries.map((entry) => (
          <NavLink
            key={entry.key}
            to={`/o/${organisation}/p/${project.slug}/${entry.key}`}
            aria-label={entry.label}
            className="flex size-11 items-center justify-center rounded-md text-muted-foreground hover:bg-muted"
          >
            <entry.icon aria-hidden="true" className="size-5" />
          </NavLink>
        ))}
      </nav>

      <div className="flex min-w-0 shrink grow-[999] basis-[560px] flex-col">
        <header className="flex min-h-14 flex-wrap items-center gap-3 border-b border-border bg-card px-6 py-3">
          <Button variant="ghost" size="iconCompact" asChild aria-label={say("editor-shell.back-to-content")}>
            <NavLink to={`/o/${organisation}/p/${project.slug}/content`}>
              <ArrowLeft aria-hidden="true" className="size-4" />
            </NavLink>
          </Button>
          <h1 className="flex-1 truncate text-lg font-semibold">{breadcrumb}</h1>
          <SyncIndicator />
          <Segmented
            label={say("editor-shell.how-to-show-this-page")}
            value={layout}
            onValueChange={(next) => onLayoutChange(next as EditorLayout)}
            options={[
              { value: "form", label: "Form" },
              { value: "split", label: say("editor-shell.side-by-side") },
              { value: "preview", label: "Preview" },
            ]}
            className="hidden md:inline-flex"
          />
          <AssistantSlot />
          <PersonMenu name={viewer.name} />
          <Button asChild count={project.unpublished}>
            <NavLink to={`/o/${organisation}/p/${project.slug}/publish`}>{say("common.publish")}</NavLink>
          </Button>
        </header>

        <main data-testid="editor-slots" className="flex min-w-0 flex-1 flex-wrap items-stretch">
          {layout !== "preview" && (
            <section
              data-testid="editor-slot-form"
              aria-label={say("editor-shell.page-fields")}
              className="min-w-0 shrink grow basis-[420px] p-6"
            >
              <div className="mx-auto w-full max-w-[760px]">{form}</div>
            </section>
          )}

          {layout !== "form" && (
            <section
              data-testid="editor-slot-preview"
              aria-label={say("editor-shell.preview")}
              className="min-w-0 shrink grow basis-[420px] border-l border-border bg-background p-6"
            >
              {preview}
            </section>
          )}

          <aside
            data-testid="editor-slot-assistant"
            aria-label={say("editor-shell.assistant")}
            hidden
            className="w-[340px] shrink-0 border-l border-border bg-card"
          />
        </main>
      </div>
    </div>
  );
}
