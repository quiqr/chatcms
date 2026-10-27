import { afterEach, describe, expect, it } from "vitest";

import { speak } from "../say";
import { onADay, when, withTheClock } from "../when";

const theMoment = new Date("2026-10-27T15:00:00");

afterEach(() => {
  withTheClock(() => new Date());
});

function at(moment: Date) {
  withTheClock(() => moment);
}

describe("how a time is written", () => {
  it("says just now for something that has only just happened", () => {
    at(theMoment);
    expect(when(new Date(theMoment.getTime() - 20_000))).toBe("just now");
  });

  it("counts the minutes within the hour", () => {
    at(theMoment);
    expect(when(new Date(theMoment.getTime() - 12 * 60_000))).toBe("12 min ago");
  });

  it("gives the time for something else today", () => {
    at(theMoment);
    const earlier = new Date("2026-10-27T09:12:00");
    expect(when(earlier)).toMatch(/09:12|9:12/);
    expect(when(earlier)).not.toMatch(/October/);
  });

  it("says yesterday", () => {
    at(theMoment);
    expect(when(new Date("2026-10-26T09:12:00"))).toMatch(/^Yesterday at /);
  });

  it("names the day within the week", () => {
    at(theMoment);
    // 23 October 2026 is a Friday, four days before.
    expect(when(new Date("2026-10-23T09:12:00"))).toMatch(/^Friday at /);
  });

  it("gives a date once it is further away than a week", () => {
    at(theMoment);
    expect(when(new Date("2026-09-02T09:12:00"))).toMatch(/September/);
  });

  it("names the year only when it is not this one", () => {
    at(theMoment);
    expect(onADay(new Date("2026-03-01T09:00:00"))).not.toMatch(/2026/);
    expect(onADay(new Date("2025-03-01T09:00:00"))).toMatch(/2025/);
  });

  it("says nothing about a moment that is not one", () => {
    at(theMoment);
    expect(when("not a date at all")).toBe("");
    expect(onADay("not a date at all")).toBe("");
  });
});

describe("a date in the language that is in force", () => {
  it("writes the words and the day in Dutch", () => {
    at(theMoment);
    speak("nl");

    expect(when(new Date(theMoment.getTime() - 20_000))).toBe("zojuist");
    expect(when(new Date(theMoment.getTime() - 12 * 60_000))).toBe("12 min geleden");

    const yesterday = new Date("2026-10-26T09:12:00");
    expect(when(yesterday)).toMatch(/^Gisteren om /);

    const thisWeek = new Date("2026-10-23T09:12:00");
    expect(when(thisWeek)).toMatch(/^vrijdag om /);

    expect(onADay(new Date("2026-03-04T09:12:00"))).toBe("4 maart");

    speak("en");
  });
});
