import { ArrowLeft, FileText, History, Image, Upload } from "lucide-react";
import type { ReactNode } from "react";
import { NavLink } from "react-router";

import { Button } from "../components/ui/button";
import { Segmented } from "../components/ui/segmented";
import { AssistantSlot } from "./ProjectShell";
import { SyncIndicator } from "./SyncIndicator";
import type { DemoProject } from "../demo/content";

export type EditorLayout = "form" | "split" | "preview";

export type EditorShellProps = {
  project: DemoProject;
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
  breadcrumb,
  layout,
  onLayoutChange,
  form,
  preview,
}: EditorShellProps) {
  return (
    <div data-testid="editor-shell" className="flex min-h-dvh flex-wrap">
      <nav
        aria-label="Project sections"
        data-testid="editor-rail"
        className="hidden w-16 shrink-0 flex-col items-center gap-2 border-r border-border bg-card py-3 sm:flex"
      >
        {railEntries.map((entry) => (
          <NavLink
            key={entry.key}
            to={`/p/${project.slug}/${entry.key}`}
            aria-label={entry.label}
            className="flex size-11 items-center justify-center rounded-md text-muted-foreground hover:bg-muted"
          >
            <entry.icon aria-hidden="true" className="size-5" />
          </NavLink>
        ))}
      </nav>

      <div className="flex min-w-0 shrink grow-[999] basis-[560px] flex-col">
        <header className="flex min-h-14 flex-wrap items-center gap-3 border-b border-border bg-card px-6 py-3">
          <Button variant="ghost" size="iconCompact" asChild aria-label="Back to content">
            <NavLink to={`/p/${project.slug}/content`}>
              <ArrowLeft aria-hidden="true" className="size-4" />
            </NavLink>
          </Button>
          <h1 className="flex-1 truncate text-lg font-semibold">{breadcrumb}</h1>
          <SyncIndicator />
          <Segmented
            label="How to show this page"
            value={layout}
            onValueChange={(next) => onLayoutChange(next as EditorLayout)}
            options={[
              { value: "form", label: "Form" },
              { value: "split", label: "Side by side" },
              { value: "preview", label: "Preview" },
            ]}
            className="hidden md:inline-flex"
          />
          <AssistantSlot />
          <Button count={project.unpublished}>Publish</Button>
        </header>

        <div data-testid="editor-slots" className="flex min-w-0 flex-1 flex-wrap items-stretch">
          {layout !== "preview" && (
            <section
              data-testid="editor-slot-form"
              aria-label="Page fields"
              className="min-w-0 shrink grow basis-[420px] p-6"
            >
              <div className="mx-auto w-full max-w-[760px]">{form}</div>
            </section>
          )}

          {layout !== "form" && (
            <section
              data-testid="editor-slot-preview"
              aria-label="Preview"
              className="min-w-0 shrink grow basis-[420px] border-l border-border bg-background p-6"
            >
              {preview}
            </section>
          )}

          <aside
            data-testid="editor-slot-assistant"
            aria-label="Assistant"
            hidden
            className="w-[340px] shrink-0 border-l border-border bg-card"
          />
        </div>
      </div>
    </div>
  );
}
