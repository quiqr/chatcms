import * as TooltipPrimitive from "@radix-ui/react-tooltip";
import type { ComponentPropsWithoutRef, ReactNode } from "react";

import { cn } from "../../lib/cn";

export const TooltipProvider = TooltipPrimitive.Provider;

export function Tooltip({
  content,
  children,
  className,
  ...props
}: ComponentPropsWithoutRef<typeof TooltipPrimitive.Content> & {
  content: ReactNode;
  children: ReactNode;
}) {
  return (
    <TooltipPrimitive.Root>
      <TooltipPrimitive.Trigger asChild>{children}</TooltipPrimitive.Trigger>
      <TooltipPrimitive.Portal>
        <TooltipPrimitive.Content
          sideOffset={6}
          className={cn(
            "max-w-72 rounded-md border border-border bg-popover px-3 py-2 text-xs text-popover-foreground shadow-[0_12px_32px_rgb(16_24_40_/_0.12)]",
            className,
          )}
          {...props}
        >
          {content}
        </TooltipPrimitive.Content>
      </TooltipPrimitive.Portal>
    </TooltipPrimitive.Root>
  );
}
