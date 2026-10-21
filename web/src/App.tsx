import { useEffect, useState } from "react";

import { fetchHealth, type Health } from "./api";

type State =
  | { kind: "loading" }
  | { kind: "ready"; health: Health }
  | { kind: "failed"; message: string };

export function App() {
  const [state, setState] = useState<State>({ kind: "loading" });

  useEffect(() => {
    let cancelled = false;

    fetchHealth()
      .then((health) => {
        if (!cancelled) setState({ kind: "ready", health });
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        const message =
          error instanceof Error ? error.message : "The server could not be reached.";
        setState({ kind: "failed", message });
      });

    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <main className="mx-auto flex min-h-dvh max-w-xl flex-col justify-center gap-4 p-6">
      <h1 className="text-2xl font-semibold tracking-tight">Quiqr ChatCMS</h1>
      {state.kind === "loading" && <p>Checking the server...</p>}
      {state.kind === "failed" && <p role="alert">{state.message}</p>}
      {state.kind === "ready" && (
        <dl className="grid grid-cols-2 gap-2 text-sm">
          <dt className="text-neutral-500">Server</dt>
          <dd>{state.health.serverVersion}</dd>
          <dt className="text-neutral-500">Schema</dt>
          <dd>{state.health.schemaVersion}</dd>
          <dt className="text-neutral-500">Database</dt>
          <dd>{state.health.database}</dd>
        </dl>
      )}
    </main>
  );
}
