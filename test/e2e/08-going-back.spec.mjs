// Acceptance test 8: roll back to an earlier version, and have that itself
// be something that can be undone.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

test("a site goes back to an earlier version, and back again", async ({ signedIn: page, chatcms }, testInfo) => {
  const project = chatcms.projectId();

  await test.step("the site is published, so there is a version to go back from", async () => {
    await page.goto(`${chatcms.url}/p/korenaar/publish`);
    await page.getByLabel("What's in this update?").fill("Spring prices");
    await page.getByRole("button", { name: /^Publish \d+ changes?$/ }).click();
    await expect(page.getByText("Your site is up to date")).toBeVisible({ timeout: 60_000 });

    const live = chatcms.onTheRemote(project, "cat-file", "blob", "refs/heads/main:content/_index.md");
    expect(live).toContain("plums");
  });

  await test.step("the history reads in plain language, with one version live", async () => {
    await page.goto(`${chatcms.url}/p/korenaar/history`);

    await expect(page.getByText("Live now")).toHaveCount(1);
    await expect(page.getByRole("heading", { name: "Today" })).toBeVisible();

    const everything = (await page.locator("body").textContent()) ?? "";
    for (const word of ["commit", "SHA", "branch", "HEAD"]) {
      expect(everything.toLowerCase()).not.toContain(word.toLowerCase());
    }
  });

  await test.step("going back says what will happen, and nothing is lost", async () => {
    await page.getByRole("button", { name: "Go back to this version…" }).last().click();

    const asking = page.getByRole("dialog");
    await expect(asking).toContainText("Nothing is lost");
    await expect(asking.getByLabel("Publish right away")).toBeChecked();

    await asking.getByRole("button", { name: "Yes, go back" }).click();
    await expect(page.getByText(/Your site is back to/)).toBeVisible({ timeout: 60_000 });

    const live = chatcms.onTheRemote(project, "cat-file", "blob", "refs/heads/main:content/_index.md");
    expect(live).toContain("apples");
  });

  await test.step("going back is itself something to undo", async () => {
    // The version it moved away from is still in the history, so going back
    // to that one brings the site forward again.
    await page.getByRole("button", { name: "Go back to this version…" }).first().click();

    const asking = page.getByRole("dialog");
    await asking.getByRole("button", { name: "Yes, go back" }).click();
    await expect(page.getByText(/Your site is back to/)).toBeVisible({ timeout: 60_000 });

    const live = chatcms.onTheRemote(project, "cat-file", "blob", "refs/heads/main:content/_index.md");
    expect(live).toContain("plums");
  });

  await worksOnAPhone(page, testInfo);
});
