import { locale, say } from "./say";

/**
 * How a time is written, in one place. Four screens each decided for
 * themselves and produced three different formats; the design asks for
 * relative and human, which is one rule.
 */

const minute = 60_000;
const hour = 60 * minute;
const day = 24 * hour;

/** Now is the clock, replaced by the tests so that "12 min ago" can be one. */
export let now = () => new Date();

export function withTheClock(clock: () => Date) {
  now = clock;
}

/**
 * when writes a moment the way somebody says it: minutes ago for something
 * that just happened, the day and the time within the week, and a date once
 * it is further away than that.
 */
export function when(at?: Date | string | null): string {
  if (!at) return "";
  const moment = typeof at === "string" ? new Date(at) : at;
  if (Number.isNaN(moment.getTime())) return "";

  const since = now().getTime() - moment.getTime();

  if (since < minute) return say("when.just-now");
  if (since < hour) {
    return say("when.minutes-ago", { minutes: Math.floor(since / minute) });
  }

  const today = startOfDay(now());
  const thatDay = startOfDay(moment);
  const daysBack = Math.round((today.getTime() - thatDay.getTime()) / day);

  const clock = moment.toLocaleTimeString(locale(), { hour: "2-digit", minute: "2-digit" });

  if (daysBack === 0) return clock;
  if (daysBack === 1) return say("when.yesterday-at", { clock });
  if (daysBack < 7) {
    return say("when.day-at", {
      day: moment.toLocaleDateString(locale(), { weekday: "long" }),
      clock,
    });
  }

  return onADay(moment);
}

/** onADay is a date with no time, for a list of days rather than moments. */
export function onADay(at?: Date | string | null): string {
  if (!at) return "";
  const moment = typeof at === "string" ? new Date(at) : at;
  if (Number.isNaN(moment.getTime())) return "";

  const sameYear = moment.getFullYear() === now().getFullYear();
  return moment.toLocaleDateString(locale(), {
    day: "numeric",
    month: "long",
    year: sameYear ? undefined : "numeric",
  });
}

function startOfDay(at: Date): Date {
  return new Date(at.getFullYear(), at.getMonth(), at.getDate());
}
