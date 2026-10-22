import * as DialogPrimitive from "@radix-ui/react-dialog";
import type { ComponentPropsWithoutRef, ReactNode } from "react";

import { cn } from "../../lib/cn";

export const Dialog = DialogPrimitive.Root;
export const DialogTrigger = DialogPrimitive.Trigger;
export const DialogClose = DialogPrimitive.Close;

export function DialogContent({
  className,
  title,
  description,
  children,
  ...props
}: ComponentPropsWithoutRef<typeof DialogPrimitive.Content> & {
  title: string;
  description?: ReactNode;
}) {
  return (
    <DialogPrimitive.Portal>
      <DialogPrimitive.Overlay className="fixed inset-0 bg-foreground/30" />
      <DialogPrimitive.Content
        className={cn(
          "fixed left-1/2 top-1/2 w-[min(32rem,calc(100vw-2rem))] -translate-x-1/2 -translate-y-1/2 rounded-2xl border border-border bg-card p-6 shadow-[0_24px_64px_rgb(16_24_40_/_0.24)]",
          className,
        )}
        {...props}
      >
        <DialogPrimitive.Title className="text-[22px] font-semibold tracking-tight">
          {title}
        </DialogPrimitive.Title>
        {description && (
          <DialogPrimitive.Description className="mt-2 text-[15px] text-muted-foreground">
            {description}
          </DialogPrimitive.Description>
        )}
        <div className="mt-5">{children}</div>
      </DialogPrimitive.Content>
    </DialogPrimitive.Portal>
  );
}
