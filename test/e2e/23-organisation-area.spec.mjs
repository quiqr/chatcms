// An organisation is a thing in itself: it has screens of its own, reached
// from the switcher by whoever administers it.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

test("an organisation admin manages the organisation from its own area", async ({
  signedIn: page,
  chatcms,
}, testInfo) => {
  await test.step("the switcher offers the way in", async () => {
    await page.goto(`${chatcms.url}/o/technative/projects`);
    await page.getByRole("button", { name: /TechNative/ }).first().click();

    await page.getByRole("menuitem", { name: /Manage TechNative/ }).click();
    await page.waitForURL(/\/o\/technative\/manage\/projects/);
  });

  await test.step("the area says what the organisation holds", async () => {
    await expect(page.getByRole("heading", { name: "Projects" })).toBeVisible();
    await expect(page.getByText("Korenaar Bakery")).toBeVisible();
    await expect(page.getByText(/Limits are set by the platform admin/)).toBeVisible();
  });

  await test.step("what is not built yet is drawn and disabled", async () => {
    await page.getByRole("button", { name: /What can be done with Korenaar Bakery/ }).click();

    const menu = page.getByRole("menu");
    await expect(menu.getByRole("menuitem", { name: /People and roles/ })).toBeEnabled();
    await expect(menu.getByRole("menuitem", { name: /Duplicate/ })).toBeDisabled();
    await expect(menu.getByText("Not built yet").first()).toBeVisible();

    await page.keyboard.press("Escape");
  });

  await test.step("somebody is invited to the organisation, and then taken out", async () => {
    await page.getByRole("link", { name: /^Members/ }).click();
    await page.waitForURL(/\/manage\/members/);

    await page.getByLabel("Email").fill("nieuw@korenaar.example");
    await page.getByRole("button", { name: "Send invite" }).click();

    await expect(page.getByText("nieuw@korenaar.example")).toBeVisible();
    await expect(page.getByText("Still to accept").first()).toBeVisible();

    await page.getByRole("button", { name: /Take nieuw@korenaar.example out/ }).click();
    await expect(page.getByText("nieuw@korenaar.example")).toHaveCount(0);
  });

  await worksOnAPhone(page, testInfo);
});

test("the project screen keeps the project", async ({ signedIn: page, chatcms }) => {
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/people`);

  // The organisation half of this screen moved to the organisation's own
  // area; what is left is the people on this site.
  await expect(page.getByRole("heading", { name: /People/ }).first()).toBeVisible();
  await expect(page.getByRole("radio", { name: "Organisation" })).toHaveCount(0);
  await expect(page.getByLabel("Role for the person you invite")).toBeVisible();
});

test("somebody who merely belongs cannot reach the area", async ({
  signedIn: page,
  chatcms,
}) => {
  // The seeded editor belongs to a site, not to the organisation's
  // management.
  const refused = await page.request.get(
    `${chatcms.url}/api/organisations/studio-noord/manage`,
    { failOnStatusCode: false },
  );

  expect(refused.status(), "an organisation nobody administers was managed").not.toBe(200);
});
