import { CloudCheck, CloudOff, RefreshCw } from "lucide-react";

import { cn } from "../lib/cn";

export type SyncState = "saved" | "saving" | "offline";

const states: Record<SyncState, { label: string; icon: typeof CloudCheck; tone: string }> = {
  saved: { label: "All changes saved", icon: CloudCheck, tone: "text-success" },
  saving: { label: "Saving...", icon: RefreshCw, tone: "text-info" },
  offline: { label: "Offline, saved on this device", icon: CloudOff, tone: "text-warning" },
};

export function SyncIndicator({ state = "saved" }: { state?: SyncState }) {
  const { label, icon: Icon, tone } = states[state];

  return (
    <span className="inline-flex items-center gap-2 text-[13px] text-muted-foreground" aria-live="polite">
      <Icon aria-hidden="true" className={cn("size-[18px]", tone, state === "saving" && "animate-spin")} />
      {label}
    </span>
  );
}
