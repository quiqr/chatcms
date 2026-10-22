import { useSyncExternalStore } from "react";

// The width below which the phone frame is used. It is the same number the
// stylesheet uses for its small breakpoint, named here so the two cannot
// drift apart silently.
export const phoneWidth = 640;

const query = `(max-width: ${phoneWidth - 1}px)`;

function subscribe(onChange: () => void) {
  const media = window.matchMedia(query);
  media.addEventListener("change", onChange);
  return () => media.removeEventListener("change", onChange);
}

/**
 * useIsPhone says whether the phone frame should be used.
 *
 * The design foundation rendered both frames and let the stylesheet hide one,
 * which is right for chrome and wrong for anything that loads: a screen
 * placed in both frames mounts twice and asks the server twice. One frame is
 * chosen here instead, read synchronously so there is nothing to flash.
 */
export function useIsPhone() {
  return useSyncExternalStore(
    subscribe,
    () => window.matchMedia(query).matches,
    () => false,
  );
}
