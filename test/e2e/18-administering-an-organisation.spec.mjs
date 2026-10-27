// An organisation opened from the administration screen: everything in it in
// one place, renamed, people put in and taken out, and removed when it is
// empty.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

async function start(page, chatcms, name, owner) {
  await page.goto(`${chatcms.url}/administration`);
  await page.getByLabel("Name").fill(name);
  await page.getByLabel("Who owns it").fill(owner);
  await page.getByRole("button", { name: "Start it" }).click();
}

test("one is opened, renamed and removed", async ({ administering: page, chatcms }, testInfo) => {
  await start(page, chatcms, "Bakkerij Zuid", "owner@example.org");
  await expect(page.getByText("/o/bakkerij-zuid")).toBeVisible();

  await test.step("it opens with who is in it and what it holds", async () => {
    await page.getByRole("link", { name: /Bakkerij Zuid/ }).click();
    await page.waitForURL(/administration\/o\/bakkerij-zuid/);

    await expect(
      page.getByRole("button", { name: "Take owner@example.org out" }),
    ).toBeVisible();
    await expect(page.getByRole("heading", { name: "Its sites" })).toBeVisible();
    await expect(page.getByText("None.")).toBeVisible();
  });

  await test.step("the name changes and the address stays", async () => {
    await page.getByLabel("Name").fill("Bakkerij Zuid BV");
    await page.getByRole("button", { name: "Save the name" }).click();

    await expect(page.getByText("/o/bakkerij-zuid")).toBeVisible();
    await expect(page.getByLabel("Name")).toHaveValue("Bakkerij Zuid BV");
  });

  await test.step("somebody is put in and taken out again", async () => {
    await page.getByLabel("Put somebody in").fill("marieke@korenaar.example");
    await page.getByRole("button", { name: "Put them in" }).click();

    const hers = page.getByRole("button", { name: "Take marieke@korenaar.example out" });
    await expect(hers).toBeVisible();

    await hers.click();
    await expect(hers).toHaveCount(0);
  });

  await test.step("removing it says what stays, and then it is gone", async () => {
    await page.getByRole("button", { name: "Remove this organisation..." }).click();
    await expect(page.getByText(/Nothing in git is touched/)).toBeVisible();

    await page.getByRole("button", { name: "Yes, remove it" }).click();
    await page.waitForURL(/administration$/);

    await expect(page.getByText("/o/bakkerij-zuid")).toHaveCount(0);
  });

  await worksOnAPhone(page, testInfo);
});

test("one that still holds a site says so instead of removing it", async ({
  administering: page,
  chatcms,
}) => {
  await page.goto(`${chatcms.url}/administration/o/technative`);

  await expect(page.getByText("Korenaar Bakery")).toBeVisible();
  await expect(page.getByText(/Remove them first/)).toBeVisible();
  await expect(page.getByRole("button", { name: "Remove this organisation..." })).toHaveCount(0);
});

test("somebody who does not administer cannot open one", async ({ signedIn: page, chatcms }) => {
  const refused = await page.request.get(
    `${chatcms.url}/api/administration/organisations/technative`,
    { failOnStatusCode: false },
  );

  expect(refused.status(), "an owner read the administration view of an organisation").toBe(403);
});
