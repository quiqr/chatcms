import { useState } from "react";

import { Button } from "../components/ui/button";
import { RadioCards } from "../components/ui/radio-cards";
import { chooseAppearance } from "../lib/session";
import { applyAppearance } from "../lib/appearance";
import { useHere } from "../lib/whoami";

export function PreferencesPage() {
  const { who, again } = useHere();

  const [chosen, setChosen] = useState(who.appearance ?? "");
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
      setTrouble(error instanceof Error ? error.message : "That could not be saved.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex max-w-[520px] flex-col gap-6">
      <section className="flex flex-col gap-4">
        <h2 className="text-[15px] font-semibold">How it looks</h2>

        <RadioCards
          label="How it looks"
          value={chosen}
          onValueChange={(next) => void choose(next)}
          options={[
            {
              value: "",
              title: "Follow my device",
              description: "Light or dark, whichever your device is set to.",
            },
            { value: "light", title: "Light", description: "Always light." },
            { value: "dark", title: "Dark", description: "Always dark." },
          ]}
        />

        {trouble && (
          <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
            {trouble}
          </p>
        )}

        <p className="text-[13px] text-subtle-foreground">
          This follows you: it is the same wherever you sign in.
        </p>

        <div>
          <Button variant="outline" disabled={busy} onClick={() => void choose("")}>
            Follow my device again
          </Button>
        </div>
      </section>

      <section className="flex flex-col gap-2">
        <h2 className="text-[15px] font-semibold">Language</h2>
        <p className="text-[15px] text-muted-foreground">
          English, for now. Dutch is being worked on.
        </p>
      </section>
    </div>
  );
}
