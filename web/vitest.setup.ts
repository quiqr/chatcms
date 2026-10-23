import "@testing-library/jest-dom/vitest";
import { cleanup, configure } from "@testing-library/react";
import { afterEach } from "vitest";

// A screen waits for an answer from the server before it renders anything, so
// every assertion about one is a wait. The default second is enough on a
// developer's machine and not on a loaded build machine, where the suite then
// fails for no reason anybody can act on. Five seconds is still short enough
// that a genuinely stuck test fails rather than hangs.
configure({ asyncUtilTimeout: 5000 });

afterEach(() => {
  cleanup();
});

// jsdom does not implement matchMedia. The interface asks it which frame to
// use, so a stand-in is installed that reports a wide window unless a test
// replaces it.
if (!window.matchMedia) {
  window.matchMedia = ((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addEventListener: () => {},
    removeEventListener: () => {},
    addListener: () => {},
    removeListener: () => {},
    dispatchEvent: () => false,
  })) as unknown as typeof window.matchMedia;
}
