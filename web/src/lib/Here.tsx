import { Fragment, useEffect, useRef, type ReactNode } from "react";
import { useParams } from "react-router";

import { useAsync } from "./useAsync";
import { applyAppearance } from "./appearance";
import { HereContext, stopActingAs, whoAmI, type WhoAmI } from "./whoami";
import { speak, spokenNow } from "./say";
import { say } from "./say";

/**
 * ActingAs is the band across every screen while somebody is being acted
 * as. It is what makes this allowed to be the one exception to administering
 * not being editing: nobody can be acting without seeing that they are.
 */
function ActingAs({ as, really }: { as: string; really: string }) {
  return (
    <div
      role="status"
      className="sticky top-0 z-50 flex flex-wrap items-center justify-center gap-3 bg-warning-soft px-4 py-2 text-center text-[15px] text-warning-foreground"
    >
      <span>{say("here.you-are-acting-as", { really, as })}</span>
      <button
        type="button"
        className="inline-flex h-11 items-center rounded-md border border-warning-foreground/30 px-3 font-semibold hover:bg-warning-soft"
        onClick={() => {
          void stopActingAs().finally(() => {
            // A whole reload rather than a refetch: every screen below was
            // drawn as somebody else, down to what it asked the server for.
            window.location.assign("/administration");
          });
        }}
      >
        {say("here.be-myself-again")}
      </button>
    </div>
  );
}

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

  // The language is put in force during the render rather than after it, so
  // that the screens below are drawn in it the first time rather than in
  // English and then again.
  if (known) speak(known.language ?? "");

  if (!known && who.state === "loading") {
    return <p className="p-6 text-[15px] text-muted-foreground">{say("common.looking")}</p>;
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
      {/* The language keys the subtree: choosing another one redraws every
          screen below, which is a moment somebody has just asked for. */}
      <Fragment key={spokenNow()}>
      {known!.actingAs && (
        <ActingAs as={known!.person.displayName || known!.person.email} really={known!.actingAs.reallyName} />
      )}
      {children}
      </Fragment>
    </HereContext.Provider>
  );
}
