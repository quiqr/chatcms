import { afterEach, describe, expect, it, vi } from "vitest";

import { en } from "../../strings/en";
import { nl } from "../../strings/nl";
import { known, languages, maybe, plural, say, speak, spokenNow } from "../say";

afterEach(() => {
  speak("en");
  vi.unstubAllGlobals();
});

describe("the catalogue", () => {
  it("holds the same keys in every language", () => {
    // The type check says this too; a test says it out loud, and says which
    // key when somebody adds one in a hurry.
    expect(Object.keys(nl).sort()).toEqual(Object.keys(en).sort());
  });

  it("has a language file for every language it offers", () => {
    for (const language of languages) {
      expect(known(language)).toBe(language);
    }
  });

  it("says nothing in a banned word, in either language", () => {
    // check-strings holds this over the whole file; this is the one that
    // fails in a watch run while somebody is still typing.
    const banned = /\b(commit|branch|push|merge|repository|deploy|conflict)\b/i;
    for (const [key, said] of Object.entries({ ...en, ...nl })) {
      if (/^(onboarding|panels|project-settings|sync|setup)\./.test(key)) continue;
      expect(said, `${key} says "${said}"`).not.toMatch(banned);
    }
  });
});

describe("say", () => {
  it("answers in the language that is in force", () => {
    speak("en");
    expect(say("common.publish")).toBe("Publish");

    speak("nl");
    expect(say("common.publish")).toBe("Publiceren");
  });

  it("fills in what the sentence needs", () => {
    speak("en");
    expect(say("when.minutes-ago", { minutes: 12 })).toBe("12 min ago");

    speak("nl");
    expect(say("when.minutes-ago", { minutes: 12 })).toBe("12 min geleden");
  });

  it("leaves a brace it was given nothing for alone", () => {
    expect(say("when.minutes-ago")).toContain("{minutes}");
  });

  it("picks between the one and the many", () => {
    speak("en");
    expect(plural(1, "common.site-one", "common.site-many")).toBe("1 site");
    expect(plural(4, "common.site-one", "common.site-many")).toBe("4 sites");

    speak("nl");
    expect(plural(1, "common.person-one", "common.person-many")).toBe("1 persoon");
    expect(plural(4, "common.person-one", "common.person-many")).toBe("4 mensen");
  });
});

describe("which language is in force", () => {
  it("is the one chosen, whatever the browser asks for", () => {
    vi.stubGlobal("navigator", { languages: ["en-GB"] });
    expect(speak("nl")).toBe("nl");
    expect(spokenNow()).toBe("nl");
  });

  it("is the browser's when nothing was chosen", () => {
    vi.stubGlobal("navigator", { languages: ["nl-NL", "en"] });
    expect(speak("")).toBe("nl");
  });

  it("is English when the browser asks for one the interface has not got", () => {
    vi.stubGlobal("navigator", { languages: ["fr-BE"] });
    expect(speak("")).toBe("en");
  });

  it("reads a region off the language it is asked for", () => {
    expect(known("nl-BE")).toBe("nl");
    expect(known("fr")).toBeNull();
    expect(known("")).toBeNull();
  });

  it("says so on the page, so that a screen reader reads it right", () => {
    speak("nl");
    expect(document.documentElement.getAttribute("lang")).toBe("nl");
  });
});

describe("a refusal the server named", () => {
  it("is said in the language that is in force", () => {
    speak("nl");
    expect(maybe("refused.address-already-taken")).toBe("Dat adres is al bezet.");
  });

  it("is nothing at all when nobody has translated it", () => {
    expect(maybe("refused.something-nobody-has-written-yet")).toBeNull();
  });
});
