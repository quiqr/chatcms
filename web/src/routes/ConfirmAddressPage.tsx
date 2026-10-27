import { useEffect, useState } from "react";
import { NavLink, useSearchParams } from "react-router";

import { proveTheNewAddress } from "../lib/session";
import { say } from "../lib/say";
import { whatWentWrong } from "../lib/api";

/**
 * ConfirmAddressPage is where the link in the message sent to a new address
 * lands. Following it is what proves somebody reads that address, which is
 * the whole point of sending it there.
 */
export function ConfirmAddressPage() {
  const [address] = useSearchParams();
  const token = address.get("token") ?? "";

  const [state, setState] = useState<"working" | "done" | "failed">("working");
  const [trouble, setTrouble] = useState("");

  useEffect(() => {
    if (!token) {
      setState("failed");
      setTrouble("This link is not complete. Open the one from your email again.");
      return;
    }

    let cancelled = false;
    proveTheNewAddress(token)
      .then(() => {
        if (!cancelled) setState("done");
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        setState("failed");
        setTrouble(
          whatWentWrong(error, "confirm-address.that-link-cannot-be-used-any-more-ask-for-a-new"),
        );
      });

    return () => {
      cancelled = true;
    };
  }, [token]);

  return (
    <main className="mx-auto flex min-h-dvh max-w-[420px] flex-col justify-center gap-4 p-6">
      <span className="flex items-center gap-2">
        <img src="/quiqr-logo.svg" alt={say("common.quiqr")} className="h-6 w-auto" />
        <span className="text-[15px] text-muted-foreground">{say("common.chatcms")}</span>
      </span>

      {state === "working" && (
        <p className="text-[15px] text-muted-foreground">{say("confirm-address.confirming")}</p>
      )}

      {state === "done" && (
        <>
          <h1 className="text-[28px] font-semibold tracking-tight">{say("confirm-address.that-is-your-address-now")}</h1>
          <p role="status" className="text-[15px] text-muted-foreground">
            {say("confirm-address.you-sign-in-with-it-from")}
          </p>
          <NavLink to="/projects" className="inline-flex min-h-11 items-center text-[15px] text-primary">
            {say("confirm-address.go-to-your-projects")}
          </NavLink>
        </>
      )}

      {state === "failed" && (
        <>
          <h1 className="text-[28px] font-semibold tracking-tight">{say("confirm-address.that-did-not-work")}</h1>
          <p role="alert" className="text-[15px] text-muted-foreground">{trouble}</p>
          <NavLink to="/account" className="inline-flex min-h-11 items-center text-[15px] text-primary">
            {say("confirm-address.back-to-your-account")}
          </NavLink>
        </>
      )}
    </main>
  );
}
