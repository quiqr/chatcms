// Acceptance test 7: publishing shows what changed, then pushes, which is
// what the customer's own pipeline is watching for.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

test("what changed is shown, and then it is published", async ({ signedIn: page, chatcms }, testInfo) => {
  const project = chatcms.projectId();
  const before = chatcms.onTheRemote(project, "rev-parse", "refs/heads/main").trim();

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/publish`);

  await test.step("what changed is said as pages and fields", async () => {
    await expect(page.getByRole("heading", { name: /Ready to publish \d+ changes?/ })).toBeVisible();
    await expect(page.getByText("Home page")).toBeVisible();
    await expect(page.getByText("Changed", { exact: true })).toBeVisible();
  });

  await test.step("the words that differ are shown", async () => {
    await page.getByRole("button", { name: "Show changes" }).first().click();

    // The seed leaves one word changed, so one word is struck through and
    // one is added. Anything else would mean the comparison is wrong.
    await expect(page.locator("del").first()).toBeVisible();
    await expect(page.locator("ins").first()).toBeVisible();
  });

  await test.step("a note can be written for the history", async () => {
    await page.getByLabel("What's in this update?").fill("Spring prices");
  });

  await test.step("publishing reaches the remote", async () => {
    await page.getByRole("button", { name: /^Publish \d+ changes?$/ }).click();
    await expect(page.getByText("Your site is up to date")).toBeVisible({ timeout: 60_000 });

    const after = chatcms.onTheRemote(project, "rev-parse", "refs/heads/main").trim();
    expect(after, "the remote did not move").not.toBe(before);

    const live = chatcms.onTheRemote(project, "cat-file", "blob", "refs/heads/main:content/_index.md");
    expect(live).toContain("plums");
  });

  await test.step("there is nothing left waiting", async () => {
    await page.reload();
    await expect(page.getByRole("heading", { name: "Nothing is waiting" })).toBeVisible();
  });

  await worksOnAPhone(page, testInfo);
});
