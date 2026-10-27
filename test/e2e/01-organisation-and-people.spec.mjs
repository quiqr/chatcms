// Acceptance test 1: as admin, create an organisation and invite the
// customer as editor.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

test("an administrator starts an organisation and somebody is invited to it", async ({
  administering: page,
  chatcms,
}, testInfo) => {
  await test.step("only an administrator is offered it", async () => {
    await expect(page.getByRole("link", { name: "Administration" })).toBeVisible();
    await page.goto(`${chatcms.url}/administration`);
  });

  await test.step("an organisation is started, owned by somebody else", async () => {
    await page.getByLabel("Name").fill("Bakkerij Zuid");
    await expect(page.getByText("Its address will be /o/bakkerij-zuid")).toBeVisible();

    await page.getByLabel("Who owns it").fill("owner@example.org");
    await page.getByRole("button", { name: "Start it" }).click();

    // The card for it, not the line about it in the record below.
    await expect(page.getByText("/o/bakkerij-zuid")).toBeVisible();
  });

  await test.step("what was done is written down", async () => {
    await expect(page.getByText(/made an organisation/)).toBeVisible();
    await expect(page.getByText(/owned by owner@example.org/)).toBeVisible();
  });

  await test.step("the administrator is not in it", async () => {
    // They can see it, because administering is for managing organisations.
    // They are not one of its people, because administering is not
    // membership.
    await page.goto(`${chatcms.url}/o/bakkerij-zuid/settings`);

    await expect(page.getByRole("heading", { name: "Who is in it" })).toBeVisible();
    await expect(page.getByText("admin@example.org")).toHaveCount(0);
    await expect(page.getByText("owner@example.org")).toBeVisible();
  });

  await worksOnAPhone(page, testInfo);
});

test("somebody who does not administer is not offered any of it", async ({
  signedIn: page,
  chatcms,
}, testInfo) => {
  await expect(page.getByRole("link", { name: "Administration" })).toHaveCount(0);

  const refused = await page.request.post(`${chatcms.url}/api/administration/organisations`, {
    data: { slug: "sneaky", name: "Sneaky", ownerEmail: "owner@example.org" },
    failOnStatusCode: false,
  });
  expect(refused.status(), "an ordinary person made an organisation").toBe(403);

  await worksOnAPhone(page, testInfo);
});

test("a project is made and somebody invited to it as an editor", async ({
  signedIn: page,
  chatcms,
}, testInfo) => {
  await test.step("the organisation's projects are listed", async () => {
    await expect(page.getByRole("link", { name: /Korenaar/ })).toBeVisible();
  });

  await test.step("a new project is made and goes to its setup", async () => {
    await page.getByRole("button", { name: /New project/ }).click();

    const asking = page.getByRole("dialog", { name: "New project" });
    await asking.getByLabel("Name").fill("Studio Zuid");
    await expect(asking).toContainText("/o/technative/p/studio-zuid");

    await asking.getByRole("button", { name: /Make it and set it up/ }).click();
    await page.waitForURL(/\/p\/studio-zuid\/setup/);
  });

  await test.step("somebody is invited as an editor", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/people`);

    await page.getByLabel("Invite someone").fill("marieke@korenaar.example");
    await page.getByLabel("Role for the person you invite").selectOption("editor");
    await page.getByRole("button", { name: /Send invite/i }).click();

    await expect(page.getByText("marieke@korenaar.example").first()).toBeVisible();
  });

  await worksOnAPhone(page, testInfo);
});
