import * as ToggleGroup from "@radix-ui/react-toggle-group";

import { cn } from "../../lib/cn";

export type SegmentedOption = {
  value: string;
  label: string;
};

export type SegmentedProps = {
  label: string;
  options: SegmentedOption[];
  value: string;
  onValueChange: (value: string) => void;
  className?: string;
};

export function Segmented({
  label,
  options,
  value,
  onValueChange,
  className,
}: SegmentedProps) {
  return (
    <ToggleGroup.Root
      type="single"
      aria-label={label}
      value={value}
      onValueChange={(next) => {
        if (next) onValueChange(next);
      }}
      className={cn("inline-flex items-center gap-0.5 rounded-lg bg-muted p-[3px]", className)}
    >
      {options.map((option) => (
        <ToggleGroup.Item
          key={option.value}
          value={option.value}
          className="h-11 rounded-sm px-2.5 sm:h-9 text-[13px] font-medium text-muted-foreground transition-colors data-[state=on]:bg-card data-[state=on]:text-foreground data-[state=on]:shadow-sm"
        >
          {option.label}
        </ToggleGroup.Item>
      ))}
    </ToggleGroup.Root>
  );
}
