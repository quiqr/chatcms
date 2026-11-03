import type { ReactNode } from "react";

import { PlatformAdminShell } from "../shells/PlatformAdminShell";
import { mayAdminister } from "../lib/administration";
import { say } from "../lib/say";
import { useAsync } from "../lib/useAsync";
import { useHere } from "../lib/whoami";

/**
 * Everything a platform admin does happens inside this shell. The counts
 * beside its nav come with the answer that says whether they may be here at
 * all, so that moving between the screens of the area is one request rather
 * than three.
 *
 * chatcms-audience: implementer
 */
export function PlatformAdminPage({ screen }: { screen: ReactNode }) {
  const { who } = useHere();
  const [me] = useAsync(mayAdminister, []);

  if (me.state === "loading") {
    return <p className="p-6 text-[15px] text-muted-foreground">{say("common.looking")}</p>;
  }
  if (me.state === "failed" || !me.value.administers) {
    return (
      <p role="alert" className="m-6 rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
        {me.state === "failed" ? me.message : say("manage.this-is-not-yours-to-manage")}
      </p>
    );
  }

  return (
    <PlatformAdminShell
      organisations={me.value.organisations ?? 0}
      users={me.value.users ?? 0}
      viewer={{ name: who.person.displayName || who.person.email }}
    >
      {screen}
    </PlatformAdminShell>
  );
}
