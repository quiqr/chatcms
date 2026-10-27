import { useState } from "react";
import { useNavigate } from "react-router";

import { Avatar } from "../components/ui/avatar";
import { cn } from "../lib/cn";
import { Menu, MenuContent, MenuItem, MenuSeparator, MenuTrigger } from "../components/ui/menu";
import { signOut, signOutEverywhere } from "../lib/session";

/**
 * PersonMenu is who you are, your account, and the way out. It is one
 * component used by every shell rather than one each: this existed as a link
 * in the organisation shell and as a picture that did nothing in the other
 * three, so from inside a site there was no way to yourself and no way out
 * at all.
 */
export function PersonMenu({ name, className }: { name: string; className?: string }) {
  const goTo = useNavigate();
  const [busy, setBusy] = useState(false);

  async function leave(everywhere: boolean) {
    if (busy) return;
    setBusy(true);
    try {
      await (everywhere ? signOutEverywhere() : signOut());
    } catch {
      // The session is gone either way as far as this screen is concerned,
      // and sending somebody back to a screen they can no longer read would
      // be worse than sending them to sign in.
    }
    // Said out loud, because a session that merely ran out ends on the same
    // screen and the two should not look alike.
    goTo("/sign-in?signed-out=1", { replace: true });
  }

  return (
    <Menu>
      <MenuTrigger
        aria-label="You"
        // The avatar is 32px, which is a picture rather than a target. The
        // hit area around it is what a thumb has to find.
        className={cn("flex size-11 items-center justify-center rounded-full hover:bg-muted", className)}
        data-testid="person-menu"
      >
        <Avatar name={name} />
      </MenuTrigger>

      <MenuContent align="end">
        <p className="px-3 py-2 text-[13px] text-subtle-foreground">{name}</p>

        <MenuSeparator />

        <MenuItem onSelect={() => goTo("/account")}>Your account</MenuItem>

        <MenuSeparator />

        <MenuItem onSelect={() => void leave(false)}>Sign out</MenuItem>
        <MenuItem onSelect={() => void leave(true)}>Sign out everywhere</MenuItem>
      </MenuContent>
    </Menu>
  );
}
