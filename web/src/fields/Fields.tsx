import { ChevronDown, GripVertical, Plus, X } from "lucide-react";
import { useState } from "react";

import { Checkbox, Field } from "../components/ui/field";
import { Button } from "../components/ui/button";

import type { Control, Option } from "../lib/tenancy";

export type Change = (path: string, value: unknown) => void;

function Label({ control, children }: { control: Control; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1.5">
      <span className="text-sm font-medium">{control.title}</span>
      {children}
      {control.tip && <p className="text-[13px] text-subtle-foreground">{control.tip}</p>}
    </div>
  );
}

function asText(value: unknown) {
  if (value === null || value === undefined) return "";
  if (typeof value === "string") return value;
  return String(value);
}

function Text({ control, onChange }: { control: Control; onChange: Change }) {
  const long = control.config?.multiLine === true;
  const value = asText(control.value);

  if (!long) {
    return (
      <Field
        label={control.title}
        hint={control.tip}
        type="text"
        value={value}
        onChange={(event) => onChange(control.path, event.target.value)}
      />
    );
  }

  return (
    <Label control={control}>
      <textarea
        aria-label={control.title}
        value={value}
        rows={4}
        onChange={(event) => onChange(control.path, event.target.value)}
        className="min-h-24 rounded-md border border-input bg-card p-3 text-base sm:text-[15px]"
      />
      <p className="text-[13px] text-subtle-foreground">{value.length} characters</p>
    </Label>
  );
}

function Number_({ control, onChange }: { control: Control; onChange: Change }) {
  const config = control.config ?? {};
  return (
    <Field
      label={control.title}
      hint={control.tip}
      type="number"
      value={asText(control.value)}
      min={config.min as number | undefined}
      max={config.max as number | undefined}
      step={config.step as number | undefined}
      onChange={(event) =>
        onChange(control.path, event.target.value === "" ? null : Number(event.target.value))
      }
    />
  );
}

function Slider({ control, onChange }: { control: Control; onChange: Change }) {
  const config = control.config ?? {};
  return (
    <Label control={control}>
      <div className="flex items-center gap-3">
        <input
          type="range"
          aria-label={control.title}
          value={asText(control.value) || String(config.min ?? 0)}
          min={config.min as number | undefined}
          max={config.max as number | undefined}
          step={config.step as number | undefined}
          onChange={(event) => onChange(control.path, Number(event.target.value))}
          className="h-11 flex-1 accent-primary"
        />
        <span className="w-12 text-right text-[15px] tabular-nums">{asText(control.value)}</span>
      </div>
    </Label>
  );
}

function DateField({ control, onChange }: { control: Control; onChange: Change }) {
  // A date in a site can carry a time and a zone. Only the day is edited
  // here, and the rest of what was there is kept.
  const held = asText(control.value);
  const day = held.slice(0, 10);
  const rest = held.slice(10);

  return (
    <Field
      label={control.title}
      hint={control.tip}
      type="date"
      value={day}
      onChange={(event) =>
        onChange(control.path, event.target.value === "" ? null : event.target.value + rest)
      }
    />
  );
}

function Select({ control, onChange }: { control: Control; onChange: Change }) {
  const options: Option[] = control.options ?? [];
  return (
    <Label control={control}>
      <select
        aria-label={control.title}
        value={asText(control.value)}
        onChange={(event) => onChange(control.path, event.target.value || null)}
        className="h-11 rounded-md border border-input bg-card px-3 text-base sm:text-[15px]"
      >
        <option value="">Not set</option>
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </Label>
  );
}

function Boolean_({ control, onChange }: { control: Control; onChange: Change }) {
  return (
    <Checkbox
      label={control.title}
      checked={control.value === true}
      onChange={(event) => onChange(control.path, event.target.checked)}
    />
  );
}

function Colour({ control, onChange }: { control: Control; onChange: Change }) {
  // chatcms-content-colour: the site's own colour, which the person picks
  const value = asText(control.value) || "#000000";
  return (
    <Label control={control}>
      <div className="flex items-center gap-3">
        <input
          type="color"
          aria-label={control.title}
          value={value}
          onChange={(event) => onChange(control.path, event.target.value)}
          className="size-11 rounded-md border border-input bg-card"
        />
        <span className="text-[15px] tabular-nums">{value}</span>
      </div>
    </Label>
  );
}

function asList(value: unknown): unknown[] {
  return Array.isArray(value) ? value : [];
}

function ListOfValues({ control, onChange }: { control: Control; onChange: Change }) {
  const items = asList(control.value);

  function write(next: unknown[]) {
    onChange(control.path, next);
  }

  return (
    <Label control={control}>
      <ul className="flex flex-col gap-2">
        {items.map((item, index) => (
          <li key={index} className="flex items-center gap-2">
            <GripVertical aria-hidden="true" className="size-4 shrink-0 text-subtle-foreground" />
            <input
              type="text"
              aria-label={`${control.title} ${index + 1}`}
              value={asText(item)}
              onChange={(event) => {
                const next = [...items];
                next[index] = event.target.value;
                write(next);
              }}
              className="h-11 flex-1 rounded-md border border-input bg-card px-3 text-base sm:text-[15px]"
            />
            <button
              type="button"
              aria-label={`Remove ${control.title} ${index + 1}`}
              onClick={() => write(items.filter((_, at) => at !== index))}
              className="flex size-11 items-center justify-center rounded-md text-subtle-foreground hover:bg-muted"
            >
              <X aria-hidden="true" className="size-4" />
            </button>
          </li>
        ))}
      </ul>
      <Button variant="ghost" onClick={() => write([...items, ""])} className="w-fit">
        <Plus aria-hidden="true" className="size-4" />
        Add {control.title.toLowerCase()}
      </Button>
    </Label>
  );
}

function ListOfItems({ control, onChange }: { control: Control; onChange: Change }) {
  const items = asList(control.value) as Record<string, unknown>[];
  const shape = control.fields ?? [];
  const [open, setOpen] = useState<number | null>(0);

  function write(next: unknown[]) {
    onChange(control.path, next);
  }

  function titleOf(item: Record<string, unknown>, index: number) {
    const first = shape[0];
    const said = first ? asText(item?.[first.path]) : "";
    return said || `${control.title} ${index + 1}`;
  }

  return (
    <Label control={control}>
      <ul className="flex flex-col gap-2">
        {items.map((item, index) => (
          <li key={index} className="rounded-md border border-border">
            <div className="flex items-center gap-2 px-2">
              <GripVertical aria-hidden="true" className="size-4 shrink-0 text-subtle-foreground" />
              <button
                type="button"
                aria-expanded={open === index}
                onClick={() => setOpen(open === index ? null : index)}
                className="flex min-h-11 flex-1 items-center gap-1.5 text-left text-[15px] font-medium"
              >
                <ChevronDown
                  aria-hidden="true"
                  className={"size-4 transition-transform " + (open === index ? "rotate-180" : "")}
                />
                {titleOf(item, index)}
              </button>
              <button
                type="button"
                aria-label={`Remove ${titleOf(item, index)}`}
                onClick={() => write(items.filter((_, at) => at !== index))}
                className="flex size-11 items-center justify-center rounded-md text-subtle-foreground hover:bg-muted"
              >
                <X aria-hidden="true" className="size-4" />
              </button>
            </div>
            {open === index && (
              <div className="flex flex-col gap-4 border-t border-border p-3">
                {shape.map((inner) => (
                  <Draw
                    key={inner.path}
                    control={{ ...inner, value: item?.[inner.path] }}
                    onChange={(_, value) => {
                      const next = [...items];
                      next[index] = { ...item, [inner.path]: value };
                      write(next);
                    }}
                  />
                ))}
              </div>
            )}
          </li>
        ))}
      </ul>
      <Button variant="ghost" onClick={() => write([...items, {}])} className="w-fit">
        <Plus aria-hidden="true" className="size-4" />
        Add {control.title.toLowerCase()}
      </Button>
    </Label>
  );
}

function Group({ control, onChange }: { control: Control; onChange: Change }) {
  const [open, setOpen] = useState(true);
  const inside = control.fields ?? [];

  return (
    <section className="flex flex-col gap-3 rounded-lg border border-border p-3">
      <button
        type="button"
        aria-expanded={open}
        onClick={() => setOpen(!open)}
        className="flex min-h-11 items-center gap-1.5 text-left text-[15px] font-semibold"
      >
        <ChevronDown
          aria-hidden="true"
          className={"size-4 transition-transform " + (open ? "rotate-180" : "")}
        />
        {control.title}
      </button>
      {open && (
        <div className="flex flex-col gap-4">
          {inside.map((inner) => (
            <Draw key={inner.path} control={inner} onChange={onChange} />
          ))}
        </div>
      )}
    </section>
  );
}

function ReadOnly({ control }: { control: Control }) {
  return (
    <Label control={control}>
      <p className="rounded-md bg-muted px-3 py-2.5 text-[15px] text-subtle-foreground">
        {asText(control.value) || "Not set"}
      </p>
    </Label>
  );
}

function Note({ control }: { control: Control }) {
  return <p className="text-[15px] text-subtle-foreground">{control.title}</p>;
}

function Unknown({ control, onChange }: { control: Control; onChange: Change }) {
  return (
    <div className="flex flex-col gap-1.5">
      <Field
        label={control.title}
        type="text"
        value={asText(control.value)}
        onChange={(event) => onChange(control.path, event.target.value)}
      />
      <p className="text-[13px] text-subtle-foreground">
        This kind of field is shown as plain text for now. What you type is saved exactly as you
        write it.
      </p>
    </div>
  );
}

export function Draw({ control, onChange }: { control: Control; onChange: Change }) {
  if (control.unknown) {
    return <Unknown control={control} onChange={onChange} />;
  }

  switch (control.type) {
    case "section":
    case "nest":
    case "pull":
      return <Group control={control} onChange={onChange} />;
    case "accordion":
      return <ListOfItems control={control} onChange={onChange} />;
    case "leaf-array":
    case "chips":
      return <ListOfValues control={control} onChange={onChange} />;
    case "select":
      return <Select control={control} onChange={onChange} />;
    case "boolean":
      return <Boolean_ control={control} onChange={onChange} />;
    case "date":
      return <DateField control={control} onChange={onChange} />;
    case "number":
      return <Number_ control={control} onChange={onChange} />;
    case "slider":
      return <Slider control={control} onChange={onChange} />;
    case "color":
      return <Colour control={control} onChange={onChange} />;
    case "readonly":
    case "uniq":
      return <ReadOnly control={control} />;
    case "info":
    case "empty-line":
      return <Note control={control} />;
    default:
      return <Text control={control} onChange={onChange} />;
  }
}
