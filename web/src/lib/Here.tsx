import { useEffect, useRef, type ReactNode } from "react";
import { useParams } from "react-router";

import { useAsync } from "./useAsync";
import { applyAppearance } from "./appearance";
import { HereContext, whoAmI, type WhoAmI } from "./whoami";

/**
 * Knowing is what asks the server who is looking and where, once, and gives
 * it to every screen below. Until it has an answer it draws nothing, because
 * a shell that guesses its own organisation is how the wrong one gets shown.
 */
export function Knowing({ children }: { children: ReactNode }) {
  const { organisation } = useParams();
  const [who, again] = useAsync(whoAmI, []);

  // The last answer is kept while a newer one is being fetched. Without it,
  // asking again after a rename blanks every screen below and loses whatever
  // the person was in the middle of.
  const last = useRef<WhoAmI | null>(null);
  useEffect(() => {
    if (who.state === "ready") last.current = who.value;
  }, [who]);

  const known = who.state === "ready" ? who.value : last.current;

  // What the account says wins over what this browser remembered, because
  // the account is where the choice is kept and this machine may be a new
  // one.
  useEffect(() => {
    if (who.state === "ready") applyAppearance(who.value.appearance ?? "");
  }, [who]);

  if (!known && who.state === "loading") {
    return <p className="p-6 text-[15px] text-muted-foreground">Looking...</p>;
  }

  if (!known && who.state === "failed") {
    return (
      <p role="alert" className="m-6 rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
        {who.message}
      </p>
    );
  }

  // With no organisation in the address, the first one somebody belongs to is
  // where they are. An administrator who belongs to none still has the
  // organisations they may manage listed, so this picks one of those.
  const named = organisation ?? known!.organisations[0]?.slug ?? "";

  return (
    <HereContext.Provider value={{ who: known!, organisation: named, again }}>
      {children}
    </HereContext.Provider>
  );
}
