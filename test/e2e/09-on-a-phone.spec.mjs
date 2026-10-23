// Acceptance test 9: everything works on a phone.
//
// Every other spec ends by checking the two phone rules on whatever screen it
// finished on, so this one walks the screens none of them lands on and
// checks the same two rules there. Between them, every screen is covered on
// a phone by the same assertions rather than by a tour nobody keeps in step.
import { test, expect, onAPhone, worksOnAPhone } from "./fixtures.mjs";

const screens = [
  ["the projects", "/projects"],
  ["the content browser", "/p/korenaar/content"],
  ["the pictures", "/p/korenaar/media"],
  ["the people", "/p/korenaar/people"],
  ["setting a project up", "/p/studio-noord/setup"],
  ["how the site stands", "/p/korenaar/sync"],
];

for (const [what, where] of screens) {
  test(`${what} works on a phone`, async ({ signedIn: page, chatcms }, testInfo) => {
    test.skip(!onAPhone(testInfo), "this is about the phone layout");

    await page.goto(`${chatcms.url}${where}`);
    // Waited for, so that the rules are checked against a screen that has
    // drawn rather than against something still loading.
    await page.waitForLoadState("networkidle");
    await expect(page.locator("body")).not.toBeEmpty();

    await worksOnAPhone(page, testInfo);
  });
}

test("the editor works on a phone", async ({ signedIn: page, chatcms }, testInfo) => {
  test.skip(!onAPhone(testInfo), "this is about the phone layout");

  await page.goto(`${chatcms.url}/p/korenaar/content/home`);
  await expect(page.getByRole("textbox", { name: "Page title" })).toBeVisible();

  await worksOnAPhone(page, testInfo);
});
