import { useEffect, useState } from "react";

import { mayAdminister } from "../lib/administration";

/**
 * useAdministers says whether the person looking administers the
 * installation, so that the shell can leave out what they may not do rather
 * than offering it and refusing afterwards. It answers no until the server
 * has said otherwise, which is the safe way round.
 */
export function useAdministers() {
  const [administers, setAdministers] = useState(false);

  useEffect(() => {
    let cancelled = false;
    mayAdminister()
      .then((answer) => {
        if (!cancelled) setAdministers(answer.administers);
      })
      .catch(() => {
        // Not being able to ask is not being allowed.
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return administers;
}
