// Signing out: there is a way out of every screen, and taking it really
// ends the session.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

// The screens somebody who edits a site actually uses, which is where the
// way out was missing: it existed on the projects screen's shell and on no
// other.
const everyScreen = [
  ["the projects", "/projects"],
  ["a site", "/o/technative/p/korenaar/content"],
  ["the editor", "/o/technative/p/korenaar/content/home"],
];

for (const [what, where] of everyScreen) {
  test(`${what} offers the way out`, async ({ signedIn: page, chatcms }, testInfo) => {
    await page.goto(`${chatcms.url}${where}`);

    await page.getByTestId("person-menu").click();
    await expect(page.getByText("Your account")).toBeVisible();
    await expect(page.getByText("Sign out", { exact: true })).toBeVisible();

    await worksOnAPhone(page, testInfo);
  });
}

test("the administration offers the way out", async ({ administering: page, chatcms }, testInfo) => {
  await page.goto(`${chatcms.url}/administration`);

  await page.getByTestId("person-menu").click();
  await expect(page.getByText("Your account")).toBeVisible();
  await expect(page.getByText("Sign out", { exact: true })).toBeVisible();

  await worksOnAPhone(page, testInfo);
});

test("signing out ends the session, and says so", async ({ signedIn: page, chatcms }, testInfo) => {
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content`);

  await test.step("the way out is taken", async () => {
    await page.getByTestId("person-menu").click();
    await page.getByText("Sign out", { exact: true }).click();

    await page.waitForURL(/sign-in/);
    await expect(page.getByRole("status")).toContainText("You have signed out.");
  });

  await test.step("the session is really gone", async () => {
    // Going back in the browser must not show a signed-in screen.
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content`);
    await expect(page).toHaveURL(/sign-in/);

    const refused = await page.request.get(
      `${chatcms.url}/api/organisations/technative/projects`,
      { failOnStatusCode: false },
    );
    expect(refused.status(), "the old session still works").toBe(401);
  });

  await worksOnAPhone(page, testInfo);
});

test("signing out everywhere ends the one asking too", async ({ signedIn: page, chatcms }) => {
  await page.goto(`${chatcms.url}/projects`);

  await page.getByTestId("person-menu").click();
  await page.getByText("Sign out everywhere").click();

  await page.waitForURL(/sign-in/);

  const refused = await page.request.get(
    `${chatcms.url}/api/organisations/technative/projects`,
    { failOnStatusCode: false },
  );
  expect(refused.status(), "the session that asked still works").toBe(401);
});
