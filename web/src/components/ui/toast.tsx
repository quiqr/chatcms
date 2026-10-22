import type React from "react";
import { Toaster as Sonner, toast } from "sonner";

export function Toaster() {
  return (
    <Sonner
      theme="dark"
      position="bottom-center"
      // Without this the toast container keeps its own fixed width and makes
      // the page scroll sideways on a phone, which is the one thing the
      // design says must never happen.
      style={{ "--width": "min(356px, calc(100vw - 2rem))" } as React.CSSProperties}
      toastOptions={{
        classNames: {
          toast: "rounded-lg border border-border bg-secondary text-secondary-foreground",
        },
      }}
    />
  );
}

export { toast };
