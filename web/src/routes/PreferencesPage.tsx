import { useState } from "react";

import { Button } from "../components/ui/button";
import { RadioCards } from "../components/ui/radio-cards";
import { chooseAppearance, chooseLanguage } from "../lib/session";
import { applyAppearance } from "../lib/appearance";
import { useHere } from "../lib/whoami";
import { languageNames, languages, say, speak, type Key } from "../lib/say";
import { whatWentWrong } from "../lib/api";

export function PreferencesPage() {
  const { who, again } = useHere();

  const [chosen, setChosen] = useState(who.appearance ?? "");
  const [language, setLanguage] = useState(who.language ?? "");
  const [trouble, setTrouble] = useState("");
  const [busy, setBusy] = useState(false);

  async function choose(next: string) {
    setChosen(next);
    setBusy(true);
    setTrouble("");

    // Applied before the server answers, because somebody who just chose an
    // appearance should see it, not wait for it.
    applyAppearance(next);

    try {
      await chooseAppearance(next);
      again();
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error, "preferences.that-could-not-be-saved"));
    } finally {
      setBusy(false);
    }
  }

  async function chooseTheLanguage(next: string) {
    setLanguage(next);
    setBusy(true);
    setTrouble("");

    // Put in force before the server answers, because somebody who just
    // chose a language should read it, not wait for it.
    speak(next);

    try {
      await chooseLanguage(next);
      again();
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error, "preferences.that-could-not-be-saved"));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex max-w-[520px] flex-col gap-6">
      <section className="flex flex-col gap-4">
        <h2 className="text-[15px] font-semibold">{say("preferences.how-it-looks")}</h2>

        <RadioCards
          label={say("preferences.how-it-looks")}
          value={chosen}
          onValueChange={(next) => void choose(next)}
          options={[
            {
              value: "",
              title: say("preferences.follow-my-device"),
              description: say("preferences.light-or-dark-whichever-your-device-is-set-to"),
            },
            { value: "light", title: "Light", description: say("preferences.always-light") },
            { value: "dark", title: "Dark", description: say("preferences.always-dark") },
          ]}
        />

        {trouble && (
          <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
            {trouble}
          </p>
        )}

        <p className="text-[13px] text-subtle-foreground">
          {say("preferences.this-follows-you-it-is-the")}
        </p>

        <div>
          <Button variant="outline" disabled={busy} onClick={() => void choose("")}>
            {say("preferences.follow-my-device-again")}
          </Button>
        </div>
      </section>

      <section className="flex flex-col gap-4">
        <h2 className="text-[15px] font-semibold">{say("preferences.language")}</h2>

        <RadioCards
          label={say("preferences.language")}
          value={language}
          onValueChange={(next) => void chooseTheLanguage(next)}
          options={[
            {
              value: "",
              title: say("preferences.follow-my-browser"),
              description: say("preferences.whichever-your-browser-asks-for"),
            },
            // The key is put together from the language: "preferences.in-en"
            // and "preferences.in-nl", which is what each language says
            // about itself.
            ...languages.map((each) => ({
              value: each,
              title: languageNames[each],
              description: say(`preferences.in-${each}` as Key),
            })),
          ]}
        />

        <p className="text-[13px] text-subtle-foreground">
          {say("preferences.this-follows-you-it-is-the")}
        </p>
      </section>
    </div>
  );
}
