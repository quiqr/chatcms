// Acceptance test 6: a change pushed straight to the remote shows up in
// ChatCMS by itself.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

test("work done outside ChatCMS arrives", async ({ signedIn: page, chatcms }, testInfo) => {
  await test.step("somebody edits the site outside ChatCMS", async () => {
    chatcms.pushFromSomewhereElse(
      chatcms.projectId(),
      "content/news/from-elsewhere.md",
      "---\ntitle: Written elsewhere\n---\n\nSomebody else wrote this.\n",
      "Add a page from outside",
    );
  });

  await test.step("ChatCMS brings it in and says whose it was", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/sync`);
    await page.getByRole("button", { name: "Check for updates" }).click();

    await expect(page.getByText(/Joris updated this site/)).toBeVisible({ timeout: 60_000 });
  });

  await test.step("the page is there to edit", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content`);
    await expect(page.getByRole("button", { name: "From elsewhere", exact: true })).toBeVisible();
  });

  await worksOnAPhone(page, testInfo);
});
