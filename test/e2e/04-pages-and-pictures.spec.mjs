// Acceptance test 4: create, edit and delete a page, and upload an image.
import { test, expect, saves, worksOnAPhone } from "./fixtures.mjs";

test("a page is made, edited and removed", async ({ signedIn: page, chatcms }, testInfo) => {
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content`);

  await test.step("a page is made and opens", async () => {
    await page.getByRole("button", { name: /New / }).first().click();
    await page.getByLabel("Name").fill("Autumn prices");
    await page.getByRole("button", { name: "Make it" }).click();

    await page.waitForURL(/content\/news\/autumn-prices/);
    await expect(page.getByRole("textbox", { name: "Headline" })).toHaveValue("Autumn prices");
  });

  await test.step("it is edited", async () => {
    await saves(page, () => page.getByRole("textbox", { name: "Headline" }).fill("Summer prices"));
  });

  await test.step("it is listed, and marked as not published", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content`);
    await expect(page.getByRole("button", { name: "Autumn prices", exact: true })).toBeVisible();
    await expect(page.getByText("Not published").first()).toBeVisible();
  });

  await test.step("removing asks first, and says what will happen", async () => {
    await page.getByRole("button", { name: "What to do with Autumn prices" }).click();
    await page.getByRole("button", { name: "Remove…" }).click();

    const asking = page.getByRole("dialog");
    await expect(asking).toContainText("the next time you publish");
    await expect(asking).toContainText("bring it back");

    await asking.getByRole("button", { name: "Yes, remove it" }).click();
    await expect(page.getByRole("button", { name: "Autumn prices", exact: true })).toHaveCount(0);
  });

  await worksOnAPhone(page, testInfo);
});

test("a picture is added, described and used", async ({ signedIn: page, chatcms }, testInfo) => {
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/media`);

  await test.step("a picture is added", async () => {
    await page.getByLabel("Add pictures").setInputFiles({
      name: "Our New Loaf.png",
      mimeType: "image/png",
      buffer: onePixel(),
    });

    await expect(page.getByText("our-new-loaf.png")).toBeVisible({ timeout: 30_000 });
  });

  await test.step("it is described for somebody using a screen reader", async () => {
    await page.getByText("our-new-loaf.png").click();

    const details = page.getByRole("complementary", { name: /About our-new-loaf/ });
    await expect(details).toContainText("Not used anywhere yet");

    // Leaving the field is what writes the description, so the write is
    // waited for rather than assumed.
    const written = page.waitForResponse(
      (answer) => answer.url().includes("/images/") && answer.request().method() === "PUT",
      { timeout: 30_000 },
    );
    await details.getByLabel("Description for screen readers").fill("A loaf on a board");
    await details.getByLabel("Description for screen readers").blur();
    await written;

    await page.reload();
    await page.getByText("our-new-loaf.png").click();
    await expect(page.getByLabel("Description for screen readers")).toHaveValue("A loaf on a board");
  });

  await worksOnAPhone(page, testInfo);
});

// onePixel is the smallest real PNG, so that what is uploaded is a picture
// rather than bytes named like one.
function onePixel() {
  return Buffer.from(
    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
    "base64",
  );
}
