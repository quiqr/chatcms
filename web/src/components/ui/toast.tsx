import { Toaster as Sonner, toast } from "sonner";

export function Toaster() {
  return (
    <Sonner
      theme="dark"
      position="bottom-center"
      toastOptions={{
        classNames: {
          toast: "rounded-lg border border-border bg-secondary text-secondary-foreground",
        },
      }}
    />
  );
}

export { toast };
