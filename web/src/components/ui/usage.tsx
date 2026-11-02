import { plural, say } from "../../lib/say";
import { cn } from "../../lib/cn";

/** Counted is one of the two things an organisation can run out of. */
export type Counted = {
  held: number;
  allowed?: number;
  full: boolean;
};

export type Usage = {
  members: Counted;
  projects: Counted;
};

/**
 * HowFull says what an organisation holds against what it is allowed. One
 * component, because the organisation's own screens and the platform admin's
 * must not disagree about how full it is, and two of these would.
 *
 * With no limit there is a number and no bar: a bar with no end to fill is a
 * shape that says nothing.
 */
export function HowFull({ counted, of }: { counted: Counted; of: "seats" | "sites" }) {
  if (counted.allowed === undefined) {
    return (
      <span className="text-[13px] text-subtle-foreground">
        {say("limits.no-limit", { held: counted.held })}
      </span>
    );
  }

  const share = Math.min(1, counted.held / counted.allowed);
  const nearly = share >= 0.8 && share < 1;

  return (
    <span className="flex flex-col gap-1">
      <span className="flex flex-wrap items-center gap-2 text-[13px] text-subtle-foreground">
        {say(of === "seats" ? "limits.seats-used" : "limits.sites-used", {
          held: counted.held,
          allowed: counted.allowed,
        })}
        {counted.full && (
          <span className="font-medium text-destructive-foreground">{say("limits.full")}</span>
        )}
        {nearly && (
          <span className="font-medium text-warning-foreground">{say("limits.almost-full")}</span>
        )}
      </span>
      <span className="h-1.5 w-full max-w-[160px] overflow-hidden rounded-full bg-muted">
        <span
          className={cn(
            "block h-full rounded-full",
            counted.full ? "bg-destructive" : nearly ? "bg-warning" : "bg-primary",
          )}
          style={{ width: `${Math.round(share * 100)}%` }}
        />
      </span>
    </span>
  );
}

/**
 * whyThereIsNoRoom is what a screen says beside a button it has disabled: the
 * limit, and what can be done about it. A refusal somebody reads after
 * pressing is a refusal they could have been spared.
 */
export function whyThereIsNoRoom(counted: Counted, of: "seats" | "sites"): string | null {
  if (!counted.full || counted.allowed === undefined) return null;

  return plural(
    counted.allowed,
    of === "seats" ? "limits.no-seats-left-one" : "limits.no-sites-left-one",
    of === "seats" ? "limits.no-seats-left-many" : "limits.no-sites-left-many",
    { allowed: counted.allowed },
  );
}
