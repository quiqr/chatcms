// Acceptance test 2: import the customer's Hugo repository with its Quiqr
// model, using a deploy key.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

test("a site is brought in with a key of its own", async ({ signedIn: page, chatcms }, testInfo) => {
  // studio-noord is seeded with no site, which is what an implementer starts
  // from.
  await page.goto(`${chatcms.url}/p/studio-noord/setup`);

  await test.step("the project has a key, and only the half that may be shown", async () => {
    await page.getByRole("button", { name: "Make a key" }).click();
    const key = page.getByText(/ssh-ed25519/);
    await expect(key).toBeVisible();
    await expect(page.locator("body")).not.toContainText("PRIVATE KEY");
  });

  await test.step("a repository is brought in and what arrived is reported", async () => {
    const remote = chatcms.remoteOf(chatcms.projectId());

    await page.getByLabel("Address of the site in git").fill(remote);
    await page.getByRole("button", { name: "Bring it in" }).click();

    await expect(page.getByText(/Hugo site/i)).toBeVisible({ timeout: 60_000 });
    await expect(page.getByText(/model/i).first()).toBeVisible();
  });

  await worksOnAPhone(page, testInfo);
});
