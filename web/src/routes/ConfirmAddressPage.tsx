import { useEffect, useState } from "react";
import { NavLink, useSearchParams } from "react-router";

import { proveTheNewAddress } from "../lib/session";

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
          error instanceof Error
            ? error.message
            : "That link cannot be used any more. Ask for a new one from your account.",
        );
      });

    return () => {
      cancelled = true;
    };
  }, [token]);

  return (
    <main className="mx-auto flex min-h-dvh max-w-[420px] flex-col justify-center gap-4 p-6">
      <span className="flex items-center gap-2">
        <img src="/quiqr-logo.svg" alt="Quiqr" className="h-6 w-auto" />
        <span className="text-[15px] text-muted-foreground">ChatCMS</span>
      </span>

      {state === "working" && (
        <p className="text-[15px] text-muted-foreground">Confirming...</p>
      )}

      {state === "done" && (
        <>
          <h1 className="text-[28px] font-semibold tracking-tight">That is your address now</h1>
          <p role="status" className="text-[15px] text-muted-foreground">
            You sign in with it from now on.
          </p>
          <NavLink to="/projects" className="inline-flex min-h-11 items-center text-[15px] text-primary">
            Go to your projects
          </NavLink>
        </>
      )}

      {state === "failed" && (
        <>
          <h1 className="text-[28px] font-semibold tracking-tight">That did not work</h1>
          <p role="alert" className="text-[15px] text-muted-foreground">{trouble}</p>
          <NavLink to="/account" className="inline-flex min-h-11 items-center text-[15px] text-primary">
            Back to your account
          </NavLink>
        </>
      )}
    </main>
  );
}
