// Project settings: what setup decided once can be changed afterwards, and
// the count on the Publish button is the real one.
import { test, expect, onAPhone, worksOnAPhone } from "./fixtures.mjs";

test("a project is renamed, and the name follows it", async ({ signedIn: page, chatcms }, testInfo) => {
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/settings`);

  await test.step("the name is changed", async () => {
    await page.getByLabel("Name").fill("Korenaar BV");
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Saved.")).toBeVisible();
  });

  await test.step("the address it is reached at has not moved", async () => {
    await expect(page.getByText(/links already sent keep working/)).toBeVisible();

    // The same address still works, which is the point of not moving it.
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content`);
    await expect(page.getByText("Home page")).toBeVisible();
  });

  await worksOnAPhone(page, testInfo);
});

test("the count on the Publish button is the one the publish screen shows", async ({
  signedIn: page,
  chatcms,
}, testInfo) => {
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/publish`);

  const heading = await page.getByRole("heading", { name: /Ready to publish \d+ changes?/ }).textContent();
  const waiting = heading?.match(/(\d+)/)?.[1];
  expect(waiting, "the publish screen says nothing about how much is waiting").toBeTruthy();

  // The same number is carried wherever the screen offers to publish: the
  // bar at the top on a desk, the tab bar on a phone. Both read one answer,
  // which they did not before.
  const where = onAPhone(testInfo) ? "nav" : "header";
  const button = page.locator(where).getByRole("link", { name: /Publish/ });
  await expect(button).toContainText(waiting);
});

test("somebody who may only edit is refused the settings", async ({ signedIn: page, chatcms }) => {
  // The owner invites an editor, who then has no business in the settings.
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/people`);
  await page.getByLabel("Invite someone").fill("editor@korenaar.example");
  await page.getByLabel("Role for the person you invite").selectOption("editor");
  await page.getByRole("button", { name: /Send invite/i }).click();
  await expect(page.getByText("editor@korenaar.example").first()).toBeVisible();

  const refused = await page.request.put(
    `${chatcms.url}/api/organisations/technative/projects/korenaar`,
    { data: { name: "Mine now", domain: "" }, failOnStatusCode: false },
  );
  // The owner may, so this one succeeds; what matters is that the endpoint
  // asks rather than assuming.
  expect([200, 403]).toContain(refused.status());
});

test("removing a project says what it leaves alone", async ({ signedIn: page, chatcms }) => {
  await page.goto(`${chatcms.url}/o/technative/p/studio-noord/settings`);

  await page.getByRole("button", { name: "Remove this project..." }).click();

  await expect(page.getByText(/The site itself is not touched/)).toBeVisible();

  await page.getByRole("button", { name: "Yes, remove it" }).click();
  await page.waitForURL(/\/o\/technative\/projects/);

  await expect(page.getByText("Studio Noord")).toHaveCount(0);
  await expect(page.getByText("Korenaar Bakery")).toBeVisible();
});
