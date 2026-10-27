/**
 * Every string the interface says comes from here. Nothing is written where
 * it is shown, which is what makes a second language a file rather than a
 * sweep through sixty screens. A check holds the rule; see
 * web/scripts/check-strings.mjs.
 *
 * say is a plain function rather than a hook, because text is also written
 * where there is no component: a date, a helper, the reason something was
 * refused.
 */
import { en } from "../strings/en";
import { nl } from "../strings/nl";

export type Strings = typeof en;
export type Key = keyof Strings;

export const languages = ["en", "nl"] as const;
export type Language = (typeof languages)[number];

const catalogues: Record<Language, Strings> = { en, nl };

/** What each language is called, in itself. */
export const languageNames: Record<Language, string> = {
  en: "English",
  nl: "Nederlands",
};

const remembered = "chatcms-language";

let spoken: Language = "en";

/** spokenNow is the language in force. */
export function spokenNow(): Language {
  return spoken;
}

/**
 * known says whether this is a language the interface has, which is how
 * what the browser asks for is turned into what can be answered.
 */
export function known(asked: string | null | undefined): Language | null {
  if (!asked) return null;
  const head = asked.toLowerCase().split("-")[0] ?? "";
  return (languages as readonly string[]).includes(head) ? (head as Language) : null;
}

/**
 * speak puts a language in force. An empty choice means the one the browser
 * asks for, because somebody who has chosen nothing should still be read to
 * in their own language.
 */
export function speak(chosen: string | null | undefined): Language {
  spoken = known(chosen) ?? fromTheBrowser();

  if (typeof document !== "undefined") {
    document.documentElement.setAttribute("lang", spoken);
  }

  try {
    if (chosen) {
      window.localStorage.setItem(remembered, chosen);
    } else {
      window.localStorage.removeItem(remembered);
    }
  } catch {
    // A browser that refuses to remember still reads correctly; it just
    // starts in the browser's language for a moment on the next load.
  }
  return spoken;
}

/**
 * speakWhatWasRemembered runs before the first paint, so that nobody reads
 * a screenful of English on the way to the answer that says Dutch.
 */
export function speakWhatWasRemembered(): Language {
  try {
    return speak(window.localStorage.getItem(remembered) ?? "");
  } catch {
    return speak("");
  }
}

function fromTheBrowser(): Language {
  if (typeof navigator === "undefined") return "en";
  for (const asked of navigator.languages ?? [navigator.language]) {
    const found = known(asked);
    if (found) return found;
  }
  return "en";
}

/**
 * say is the text for a key, with anything in braces filled in. A key that
 * is not in the catalogue is a failed type check rather than a blank screen.
 */
export function say(key: Key, fill?: Record<string, string | number>): string {
  const text: string = catalogues[spoken][key] ?? en[key];
  if (!fill) return text;

  return text.replace(/\{(\w+)\}/g, (whole, name: string) =>
    name in fill ? String(fill[name]) : whole,
  );
}

/**
 * plural picks between the one and the many, and fills in the count. Every
 * language this interface has works that way; one that does not would need
 * its own rule here rather than a third form in every sentence.
 */
export function plural(count: number, one: Key, many: Key, fill?: Record<string, string | number>): string {
  return say(count === 1 ? one : many, { count, ...fill });
}

/**
 * maybe is say for a key that may not exist, which is what a refusal from
 * the server is: it carries a key, and a refusal nobody has translated yet
 * is better said in the server's own words than not at all.
 */
export function maybe(key: string, fill?: Record<string, string | number>): string | null {
  if (!(key in en)) return null;
  return say(key as Key, fill);
}

/**
 * locale is what a date is written in. English keeps whatever the machine
 * asks for, which is how a British and an American reader each get their
 * own order; a chosen language takes over from it.
 */
export function locale(): string | undefined {
  return spoken === "en" ? undefined : spoken;
}
