import type { InputHTMLAttributes, ReactNode } from "react";
import { useId } from "react";

import { cn } from "../../lib/cn";

export type FieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string;
  hint?: ReactNode;
  action?: ReactNode;
  /** trouble is what is wrong with what is in the field, said beside it. */
  trouble?: string;
};

export function Field({ label, hint, action, trouble, className, id, ...props }: FieldProps) {
  const generated = useId();
  const fieldId = id ?? generated;
  const hintId = hint ? `${fieldId}-hint` : undefined;
  const troubleId = trouble ? `${fieldId}-trouble` : undefined;

  return (
    <div className="flex flex-col gap-1.5">
      <div className="flex items-baseline justify-between gap-2">
        <label htmlFor={fieldId} className="text-sm font-medium">
          {label}
        </label>
        {action}
      </div>
      <input
        id={fieldId}
        aria-describedby={cn(hintId, troubleId) || undefined}
        aria-invalid={trouble ? true : undefined}
        className={cn(
          "h-11 rounded-md border bg-card px-3 text-base sm:text-[15px]",
          trouble ? "border-destructive" : "border-input",
          "focus-visible:border-primary focus-visible:ring-[3px] focus-visible:ring-primary-soft",
          className,
        )}
        {...props}
      />
      {trouble && (
        <p id={troubleId} className="text-[13px] text-destructive-foreground">
          {trouble}
        </p>
      )}
      {hint && (
        <p id={hintId} className="text-[13px] text-subtle-foreground">
          {hint}
        </p>
      )}
    </div>
  );
}

export function Checkbox({ label, id, className, ...props }: FieldProps) {
  const generated = useId();
  const fieldId = id ?? generated;

  return (
    <label htmlFor={fieldId} className="flex min-h-11 items-center gap-2.5 text-[15px]">
      <input
        id={fieldId}
        type="checkbox"
        // chatcms-target: the label around this one carries the hit area
        className={cn("size-4 rounded-sm border-input accent-primary", className)}
        {...props}
      />
      {label}
    </label>
  );
}
