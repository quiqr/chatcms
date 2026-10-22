import { useCallback, useEffect, useState } from "react";

export type Loading<T> =
  | { state: "loading" }
  | { state: "ready"; value: T }
  | { state: "failed"; message: string };

/**
 * useAsync loads something once and gives back a way to load it again. It
 * exists so that every screen reports loading and failure the same way.
 */
export function useAsync<T>(load: () => Promise<T>, keys: unknown[]): [Loading<T>, () => void] {
  const [result, setResult] = useState<Loading<T>>({ state: "loading" });
  const [attempt, setAttempt] = useState(0);

  const again = useCallback(() => setAttempt((count) => count + 1), []);

  useEffect(() => {
    let cancelled = false;
    setResult({ state: "loading" });

    load()
      .then((value) => {
        if (!cancelled) setResult({ state: "ready", value });
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        setResult({
          state: "failed",
          message:
            error instanceof Error
              ? error.message
              : "We could not reach the server. Nothing is lost; please try again.",
        });
      });

    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...keys, attempt]);

  return [result, again];
}
