import { cva, type VariantProps } from "class-variance-authority";
import type { HTMLAttributes } from "react";

import { cn } from "../../lib/cn";

const badgeVariants = cva(
  "inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium",
  {
    variants: {
      tone: {
        changed: "bg-warning-soft text-warning-foreground",
        new: "bg-info-soft text-info-foreground",
        published: "bg-transparent text-muted-foreground px-0",
        live: "bg-success-soft text-success-foreground",
        attention: "bg-destructive-soft text-destructive-foreground",
        neutral: "bg-muted text-muted-foreground",
      },
    },
    defaultVariants: { tone: "neutral" },
  },
);

const tonesWithADot = new Set(["live", "attention"]);

const dotColour: Record<string, string> = {
  live: "bg-success",
  attention: "bg-destructive",
};

export type BadgeProps = HTMLAttributes<HTMLSpanElement> &
  VariantProps<typeof badgeVariants>;

export function Badge({ className, tone, children, ...props }: BadgeProps) {
  const name = tone ?? "neutral";

  return (
    <span className={cn(badgeVariants({ tone }), className)} {...props}>
      {tonesWithADot.has(name) && (
        <span
          data-testid="badge-dot"
          aria-hidden="true"
          className={cn("size-1.5 rounded-full", dotColour[name])}
        />
      )}
      {children}
    </span>
  );
}

export { badgeVariants };
