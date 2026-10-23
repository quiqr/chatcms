// Acceptance test 5: live preview shows changes within a few seconds.
import { test, expect, onAPhone, saves, worksOnAPhone } from "./fixtures.mjs";

test("a change shows in the preview within seconds", async ({ signedIn: page, chatcms }, testInfo) => {
  await page.goto(`${chatcms.url}/p/korenaar/content/home`);

  // On a phone the form and the preview take turns, so each has to be asked
  // for. On a desk they are side by side and nothing needs pressing.
  const phone = onAPhone(testInfo);
  const showPreview = async () => {
    if (phone) await page.getByRole("radio", { name: "Preview" }).click();
  };
  const showForm = async () => {
    if (phone) await page.getByRole("radio", { name: "Edit" }).click();
  };

  await showPreview();
  const preview = page.frameLocator('iframe[title="Your site"]');

  await test.step("the preview shows the site", async () => {
    // The first ask starts a Hugo server, which takes a moment on a cold
    // machine, so this one wait is longer than the rest.
    await expect(preview.locator("body")).toContainText("Korenaar Bakery", { timeout: 90_000 });
  });

  await test.step("a change shows in the preview within seconds", async () => {
    await showForm();
    await saves(page, () =>
      page.getByRole("textbox", { name: "Page title" }).fill("Korenaar Bakery and Sons"),
    );
    await showPreview();

    // Hugo rebuilds by itself when the file changes, which takes it a moment.
    // Asking again until it shows, with a deadline, is how long that moment
    // is measured; sleeping for a guess would be either slow or flaky.
    await expect(async () => {
      await page.getByRole("button", { name: "Refresh" }).click();
      await expect(preview.locator("body")).toContainText("Korenaar Bakery and Sons", {
        timeout: 2_000,
      });
    }).toPass({ timeout: 20_000, intervals: [250, 500, 1_000] });
  });

  await worksOnAPhone(page, testInfo);
});

test("a preview cannot be read by somebody not signed in", async ({ page, chatcms }) => {
  const answer = await page.request.get(`${chatcms.url}/preview/technative/korenaar/`, {
    maxRedirects: 0,
  });

  expect(answer.status(), "an unpublished site was served to nobody").toBe(401);
});
