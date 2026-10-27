// Not an acceptance test: a capture of every screen, at both widths, next to
// the design it was drawn from. The bean asks for differences to be listed
// rather than silently accepted, and a pixel comparison against a drawing
// would only ever say "different" about a thousand pixels of antialiasing.
// So this writes the screens out and docs/design/COMPARISON.md says what a
// reader of both sees.
import { mkdirSync } from "node:fs";
import { dirname, join } from "node:path";

import { test, expect, onAPhone } from "./fixtures.mjs";

const into = process.env.CHATCMS_SHOTS ?? "shots";

const screens = [
  ["projects", "/projects"],
  ["onboarding", "/o/technative/p/studio-noord/setup"],
  ["content", "/o/technative/p/korenaar/content"],
  ["editor", "/o/technative/p/korenaar/content/home"],
  ["media", "/o/technative/p/korenaar/media"],
  ["people", "/o/technative/p/korenaar/people"],
  ["history", "/o/technative/p/korenaar/history"],
  ["sync", "/o/technative/p/korenaar/sync"],
  ["publish", "/o/technative/p/korenaar/publish"],
];

async function capture(page, name, testInfo) {
  // Waiting for the main region rather than for the network to go quiet: the
  // editor holds a preview that talks to Hugo for as long as it is open, so
  // "quiet" never arrives there. Then for the screen to have stopped saying
  // it is still looking, so that the capture is of a screen, not of a wait.
  await expect(page.locator("main")).toBeVisible();
  await expect(page.getByText(/^(Looking|Checking|Opening)\b.*…$/)).toHaveCount(0);

  const file = join(into, onAPhone(testInfo) ? "phone" : "desktop", `${name}.png`);
  mkdirSync(dirname(file), { recursive: true });
  await page.screenshot({ path: file, fullPage: true });
}

test("login is captured", async ({ page, chatcms }, testInfo) => {
  await page.goto(`${chatcms.url}/sign-in`);
  await capture(page, "login", testInfo);
});

for (const [name, where] of screens) {
  test(`${name} is captured`, async ({ signedIn: page, chatcms }, testInfo) => {
    await page.goto(`${chatcms.url}${where}`);
    await capture(page, name, testInfo);
  });
}
