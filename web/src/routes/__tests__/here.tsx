import type { ReactNode } from "react";

import { HereContext } from "../../lib/whoami";

/**
 * Inside puts a screen where it expects to be: knowing who is looking and
 * which organisation the address names. Every screen test needs it, because
 * no screen guesses its own organisation any more.
 */
export function Inside({
  children,
  organisation = "technative",
  administers = false,
}: {
  children: ReactNode;
  organisation?: string;
  administers?: boolean;
}) {
  return (
    <HereContext.Provider
      value={{
        who: {
          person: { id: "person-1", email: "joris@example.org", displayName: "Joris Bakker" },
          administers,
          organisations: [{ slug: organisation, name: "TechNative" }],
        },
        organisation,
        again: () => {},
      }}
    >
      {children}
    </HereContext.Provider>
  );
}
