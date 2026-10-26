import { test as base, expect } from "@playwright/test";

import { Harness, administrator, owner } from "./harness.mjs";

// Each spec gets a ChatCMS of its own: its own database, its own remote and
// its own seeded site. Nothing a test does can reach another, so the order
// they run in cannot matter and a failure is never somebody else's leftovers.
export const test = base.extend({
  chatcms: async ({}, use) => {
    const harness = new Harness();
    // Starting can fail with a PostgreSQL already running, so the teardown
    // covers the start as well. Otherwise one bad start leaves a database
    // behind for the rest of the afternoon.
    try {
      await harness.start();
      await use(harness);
    } finally {
      await harness.stop();
    }
  },

  signedIn: async ({ page, chatcms }, use) => {
    await signIn(page, chatcms, owner);
    await use(page);
  },

  // Somebody who administers the installation, which is who acceptance test 1
  // begins with.
  administering: async ({ page, chatcms }, use) => {
    await signIn(page, chatcms, administrator);
    await use(page);
  },
});

async function signIn(page, chatcms, who) {
  await page.goto(`${chatcms.url}/sign-in`);
  await page.getByLabel("Email").fill(who.email);
  await page.getByLabel("Password").fill(who.password);
  await page.getByRole("button", { name: /Sign in/i }).click();
  try {
    await page.waitForURL(/projects/);
  } catch (why) {
    // Saying what the screen and the server said, rather than only that a
    // navigation did not happen. A suite that cannot be read from its own
    // output is a suite somebody has to reproduce by hand.
    const said = await page.getByRole("alert").allInnerTexts().catch(() => []);
    throw new Error(
      `signing in as ${who.email} did not reach the projects: ${why.message}\n` +
        `the screen said: ${said.join(" | ") || "nothing"}\n` +
        `the server said:\n${chatcms.log}`,
    );
  }
}

export { expect };

/**
 * noSidewaysScrolling is half of what "works on a phone" means. The other
 * half is that everything can be hit with a thumb, which is checked by
 * everyThingIsBigEnoughToHit.
 */
export async function noSidewaysScrolling(page) {
  const scrolls = await page.evaluate(
    () => document.documentElement.scrollWidth > document.documentElement.clientWidth + 1,
  );
  expect(scrolls, "the page scrolls sideways").toBe(false);
}

const smallestTarget = 44;

export async function everyThingIsBigEnoughToHit(page) {
  const tooSmall = await page.evaluate((least) => {
    const found = [];
    for (const element of document.querySelectorAll("button, a[href], select, input:not([type=hidden])")) {
      const box = element.getBoundingClientRect();
      if (box.width === 0 && box.height === 0) continue;
      const style = getComputedStyle(element);
      if (style.visibility === "hidden" || style.display === "none") continue;
      // A control inside a label is operated by the label, which is what
      // carries the hit area.
      if (element.closest("label") && element.tagName === "INPUT") continue;
      // A control hidden from sight is reached some other way, so it is not
      // a target a thumb has to find. The thing that opens it is.
      if (box.width <= 1 || box.height <= 1) continue;
      if (box.height < least - 0.5) {
        found.push(`${element.tagName.toLowerCase()} "${(element.textContent ?? "").trim().slice(0, 30)}" is ${Math.round(box.height)}px`);
      }
    }
    return found;
  }, smallestTarget);

  expect(tooSmall, "controls too small to hit with a thumb").toEqual([]);
}

/**
 * onAPhone says whether this pass is the phone one, so that a spec can check
 * the two phone rules without being written twice.
 */
export function onAPhone(testInfo) {
  return testInfo.project.name.endsWith("phone");
}

/**
 * worksOnAPhone runs the phone checks when this is the phone pass, and does
 * nothing otherwise. Every spec ends with it, which is how acceptance test 9
 * is covered by all the others rather than by a screen tour of its own.
 */
export async function worksOnAPhone(page, testInfo) {
  if (!onAPhone(testInfo)) return;
  await noSidewaysScrolling(page);
  await everyThingIsBigEnoughToHit(page);
}

/**
 * saves runs something that changes a page and waits for the save to reach
 * the server. Waiting for the words "All changes saved" would not do: they
 * are also what the screen says before anything has been typed, so the
 * assertion would pass without a save having happened at all.
 */
export async function saves(page, change) {
  const written = page.waitForResponse(
    (answer) =>
      answer.url().includes("/pages/") &&
      ["PUT", "POST"].includes(answer.request().method()) &&
      answer.ok(),
    { timeout: 30_000 },
  );
  await change();
  await written;
  await expect(page.getByLabel("Whether this page is saved")).toHaveText("All changes saved");
}
