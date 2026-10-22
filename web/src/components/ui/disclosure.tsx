import * as Collapsible from "@radix-ui/react-collapsible";
import { ChevronRight } from "lucide-react";
import type { ReactNode } from "react";
import { useState } from "react";

import { cn } from "../../lib/cn";

export type DisclosureProps = {
  summary: string;
  children: ReactNode;
  defaultOpen?: boolean;
  className?: string;
};

export function Disclosure({ summary, children, defaultOpen = false, className }: DisclosureProps) {
  const [open, setOpen] = useState(defaultOpen);

  return (
    <Collapsible.Root open={open} onOpenChange={setOpen} className={cn("w-full", className)}>
      <Collapsible.Trigger className="flex h-11 w-full items-center gap-2 text-left text-sm font-medium text-muted-foreground">
        <ChevronRight
          aria-hidden="true"
          className={cn("size-4 transition-transform", open && "rotate-90")}
        />
        {summary}
      </Collapsible.Trigger>
      <Collapsible.Content className="pb-2 pt-1">{children}</Collapsible.Content>
    </Collapsible.Root>
  );
}
