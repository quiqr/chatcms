/**
 * How the chosen appearance reaches the page. The stylesheet does the rest:
 * an attribute on the root element, and the tokens under it take over.
 *
 * It is also written to the browser so that the next load can apply it
 * before anything is drawn. The account is where it is kept; this copy only
 * saves a flash of the wrong one while the server is asked.
 */
const remembered = "chatcms-appearance";

export function applyAppearance(appearance: string) {
  const root = document.documentElement;

  if (appearance === "light" || appearance === "dark") {
    root.setAttribute("data-appearance", appearance);
  } else {
    root.removeAttribute("data-appearance");
  }

  try {
    if (appearance) {
      window.localStorage.setItem(remembered, appearance);
    } else {
      window.localStorage.removeItem(remembered);
    }
  } catch {
    // A browser that refuses to remember still draws correctly; it just
    // flashes the machine's appearance for a moment on the next load.
  }
}

/** applyWhatWasRemembered runs before the first paint. */
export function applyWhatWasRemembered() {
  try {
    applyAppearance(window.localStorage.getItem(remembered) ?? "");
  } catch {
    applyAppearance("");
  }
}
