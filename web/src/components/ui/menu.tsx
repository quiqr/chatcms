import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import type { ComponentPropsWithoutRef, ReactNode } from "react";

import { cn } from "../../lib/cn";

export const Menu = DropdownMenu.Root;
export const MenuTrigger = DropdownMenu.Trigger;

export function MenuContent({
  className,
  children,
  ...props
}: ComponentPropsWithoutRef<typeof DropdownMenu.Content>) {
  return (
    <DropdownMenu.Portal>
      <DropdownMenu.Content
        sideOffset={6}
        className={cn(
          "min-w-48 rounded-lg border border-border bg-popover p-1 text-popover-foreground shadow-[0_12px_32px_rgb(16_24_40_/_0.12)]",
          className,
        )}
        {...props}
      >
        {children}
      </DropdownMenu.Content>
    </DropdownMenu.Portal>
  );
}

export function MenuItem({
  className,
  destructive,
  children,
  ...props
}: ComponentPropsWithoutRef<typeof DropdownMenu.Item> & { destructive?: boolean }) {
  return (
    <DropdownMenu.Item
      className={cn(
        "flex h-11 cursor-default select-none items-center gap-2 rounded-sm px-3 text-sm outline-hidden data-highlighted:bg-muted",
        destructive && "text-destructive-foreground",
        className,
      )}
      {...props}
    >
      {children}
    </DropdownMenu.Item>
  );
}

export function MenuSeparator({ className }: { className?: string }) {
  return <DropdownMenu.Separator className={cn("my-1 h-px bg-border-subtle", className)} />;
}

export function MenuLabel({ children }: { children: ReactNode }) {
  return (
    <DropdownMenu.Label className="px-3 py-2 text-xs font-medium text-subtle-foreground">
      {children}
    </DropdownMenu.Label>
  );
}
