import * as RadioGroup from "@radix-ui/react-radio-group";
import type { ReactNode } from "react";

import { cn } from "../../lib/cn";

export type RadioCardOption = {
  value: string;
  title: string;
  description?: ReactNode;
};

export type RadioCardsProps = {
  label: string;
  options: RadioCardOption[];
  value: string;
  onValueChange: (value: string) => void;
  className?: string;
};

export function RadioCards({
  label,
  options,
  value,
  onValueChange,
  className,
}: RadioCardsProps) {
  return (
    <RadioGroup.Root
      aria-label={label}
      value={value}
      onValueChange={onValueChange}
      className={cn("grid gap-3", className)}
    >
      {options.map((option) => (
        <RadioGroup.Item
          key={option.value}
          value={option.value}
          className="min-h-11 rounded-lg border border-border bg-card p-4 text-left transition-colors data-[state=checked]:border-[1.5px] data-[state=checked]:border-primary data-[state=checked]:bg-primary-tint"
        >
          <span className="block text-[15px] font-semibold text-foreground">{option.title}</span>
          {option.description && (
            <span className="mt-1 block text-sm text-subtle-foreground">{option.description}</span>
          )}
        </RadioGroup.Item>
      ))}
    </RadioGroup.Root>
  );
}
