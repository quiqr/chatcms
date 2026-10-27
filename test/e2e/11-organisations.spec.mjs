// Organisations: the address says which one, switching shows each one's own
// work, and nothing from one appears in the other.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

test("an administrator starts a second organisation and the two stay apart", async ({
  administering: page,
  chatcms,
}, testInfo) => {
  await test.step("a second organisation, owned by the same person", async () => {
    await page.goto(`${chatcms.url}/administration`);
    await page.getByLabel("Name").fill("Bakkerij Zuid");
    await page.getByLabel("Who owns it").fill("owner@example.org");
    await page.getByRole("button", { name: "Start it" }).click();

    await expect(page.getByText("/o/bakkerij-zuid")).toBeVisible();
  });

  await worksOnAPhone(page, testInfo);
});

test("somebody in two organisations sees each one's own projects", async ({
  signedIn: page,
  chatcms,
  administering,
}, testInfo) => {
  // The administrator makes the second one; the owner is put in it.
  await administering.goto(`${chatcms.url}/administration`);
  await administering.getByLabel("Name").fill("Bakkerij Zuid");
  await administering.getByLabel("Who owns it").fill("owner@example.org");
  await administering.getByRole("button", { name: "Start it" }).click();
  await expect(administering.getByText("/o/bakkerij-zuid")).toBeVisible();

  await test.step("the first one holds the seeded sites", async () => {
    await page.goto(`${chatcms.url}/o/technative/projects`);
    await expect(page.getByText("Korenaar Bakery")).toBeVisible();
  });

  await test.step("the second one holds none of them", async () => {
    await page.goto(`${chatcms.url}/o/bakkerij-zuid/projects`);
    await expect(page.getByText("Korenaar Bakery")).toHaveCount(0);
    await expect(page.getByText("Import a site or start from a template")).toBeVisible();
  });

  await test.step("the switcher lists both", async () => {
    await page.getByRole("button", { name: "Bakkerij Zuid" }).click();

    const switcher = page.getByRole("menu");
    await expect(switcher.getByText("TechNative")).toBeVisible();
    await expect(switcher.getByText("Bakkerij Zuid")).toBeVisible();
  });

  await worksOnAPhone(page, testInfo);
});

test("an organisation somebody is not in is nothing there", async ({ signedIn: page, chatcms }) => {
  const refused = await page.request.get(
    `${chatcms.url}/api/organisations/nobody-has-this/projects`,
    { failOnStatusCode: false },
  );

  expect(refused.status(), "an organisation nobody is in was listed").toBe(404);
});

test("an organisation can be renamed, and its address stays", async ({
  administering: page,
  chatcms,
}) => {
  await page.goto(`${chatcms.url}/o/technative/settings`);

  await page.getByLabel("Name").fill("TechNative BV");
  await page.getByRole("button", { name: "Save the name" }).click();

  await expect(page.getByText("Saved.")).toBeVisible();

  // The address is what links already sent are written against.
  await page.goto(`${chatcms.url}/o/technative/projects`);
  await expect(page.getByRole("button", { name: /TechNative BV/ }).first()).toBeVisible();
});
