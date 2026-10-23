import { Bold, Image, Italic, Link2, List, Type } from "lucide-react";
import { useRef, useState } from "react";

import { Segmented } from "../components/ui/segmented";

type Wrap = { before: string; after: string; sample: string };

const actions: Array<{ key: string; label: string; icon: typeof Bold; wrap: Wrap }> = [
  { key: "heading", label: "Heading", icon: Type, wrap: { before: "## ", after: "", sample: "Heading" } },
  { key: "bold", label: "Bold", icon: Bold, wrap: { before: "**", after: "**", sample: "bold" } },
  { key: "italic", label: "Italic", icon: Italic, wrap: { before: "_", after: "_", sample: "italic" } },
  { key: "link", label: "Link", icon: Link2, wrap: { before: "[", after: "](https://)", sample: "text" } },
  { key: "list", label: "List", icon: List, wrap: { before: "- ", after: "", sample: "item" } },
  { key: "image", label: "Image", icon: Image, wrap: { before: "![", after: "](/images/)", sample: "description" } },
];

/**
 * BodyEditor edits a page's own text. It writes Markdown either way: the
 * toolbar puts the marks in for somebody who would rather not remember them,
 * and the source view shows exactly what is saved.
 */
export function BodyEditor({
  value,
  onChange,
  label = "Page text",
}: {
  value: string;
  onChange: (value: string) => void;
  label?: string;
}) {
  const [showing, setShowing] = useState<"formatted" | "markdown">("formatted");
  const box = useRef<HTMLTextAreaElement>(null);

  function apply(wrap: Wrap) {
    const area = box.current;
    if (!area) return;

    const from = area.selectionStart;
    const to = area.selectionEnd;
    const chosen = value.slice(from, to) || wrap.sample;
    const next = value.slice(0, from) + wrap.before + chosen + wrap.after + value.slice(to);

    onChange(next);
    requestAnimationFrame(() => {
      area.focus();
      area.setSelectionRange(from + wrap.before.length, from + wrap.before.length + chosen.length);
    });
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="text-sm font-medium">{label}</span>
        <Segmented
          label="How to edit the text"
          value={showing}
          options={[
            { value: "formatted", label: "Formatted" },
            { value: "markdown", label: "Markdown" },
          ]}
          onValueChange={(next) => setShowing(next as "formatted" | "markdown")}
        />
      </div>

      {showing === "formatted" && (
        <div className="flex flex-wrap gap-1" role="toolbar" aria-label="Text formatting">
          {actions.map((action) => (
            <button
              key={action.key}
              type="button"
              aria-label={action.label}
              onClick={() => apply(action.wrap)}
              className="flex size-11 items-center justify-center rounded-md text-subtle-foreground hover:bg-muted"
            >
              <action.icon aria-hidden="true" className="size-4" />
            </button>
          ))}
        </div>
      )}

      <textarea
        ref={box}
        aria-label={label}
        value={value}
        rows={14}
        spellCheck={showing === "formatted"}
        onChange={(event) => onChange(event.target.value)}
        className={
          "min-h-64 rounded-md border border-input bg-card p-3 text-base sm:text-[15px] " +
          (showing === "markdown" ? "font-mono text-[13px]" : "")
        }
      />
    </div>
  );
}
