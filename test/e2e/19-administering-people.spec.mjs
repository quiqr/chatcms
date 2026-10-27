// Every person on the installation, where each belongs, and making one.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

test("everybody is listed with where they belong", async ({
  administering: page,
  chatcms,
}, testInfo) => {
  await page.goto(`${chatcms.url}/administration`);

  await test.step("the seeded people are all there", async () => {
    await expect(page.getByRole("link", { name: "Joris Bakker" })).toBeVisible();
    await expect(page.getByRole("link", { name: "marieke@korenaar.example" })).toBeVisible();
  });

  await test.step("and it says where each of them belongs", async () => {
    const row = page.getByRole("row").filter({ hasText: "marieke@korenaar.example" });
    await expect(row).toContainText("Korenaar Bakery (Editor)");
  });

  await test.step("somebody is found by part of their address", async () => {
    await page.getByLabel("Find somebody").fill("marieke");
    await page.getByRole("button", { name: "Find" }).click();

    await expect(page.getByRole("link", { name: "marieke@korenaar.example" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Joris Bakker" })).toHaveCount(0);
  });

  await worksOnAPhone(page, testInfo);
});

test("one person opens with everything they can reach", async ({
  administering: page,
  chatcms,
}, testInfo) => {
  await page.goto(`${chatcms.url}/administration/people/marieke%40korenaar.example`);

  await expect(page.getByRole("heading", { name: /marieke@korenaar.example/ })).toBeVisible();
  await expect(page.getByText("Still to accept")).toBeVisible();
  await expect(page.getByText("Has never been here")).toBeVisible();
  await expect(page.getByText("/o/technative/p/korenaar · editor")).toBeVisible();

  await worksOnAPhone(page, testInfo);
});

test("a person is made and sent a way in, with no password set for them", async ({
  administering: page,
  chatcms,
}) => {
  await page.goto(`${chatcms.url}/administration`);

  await page.getByLabel("Email").fill("nieuw@example.org");
  await page.getByLabel("Their name").fill("Nieuwe Collega");
  await page.getByRole("button", { name: "Make them" }).click();

  await expect(page.getByText(/nieuw@example.org has been sent a way in/)).toBeVisible();

  await test.step("they are on the list, with nothing yet to reach", async () => {
    await page.goto(`${chatcms.url}/administration/people/nieuw%40example.org`);
    await expect(page.getByText("Still to accept")).toBeVisible();
    await expect(page.getByText("In none.")).toBeVisible();
  });

  await test.step("the same address again is refused", async () => {
    await page.goto(`${chatcms.url}/administration`);
    await page.getByLabel("Email").fill("nieuw@example.org");
    await page.getByRole("button", { name: "Make them" }).click();

    await expect(page.getByRole("alert")).toContainText("already has an account");
  });
});

test("somebody who does not administer is refused the people", async ({
  signedIn: page,
  chatcms,
}) => {
  const list = await page.request.get(`${chatcms.url}/api/administration/people`, {
    failOnStatusCode: false,
  });
  expect(list.status(), "an owner listed every person on the installation").toBe(403);

  const one = await page.request.get(
    `${chatcms.url}/api/administration/people/owner%40example.org`,
    { failOnStatusCode: false },
  );
  expect(one.status(), "an owner read a person through administration").toBe(403);
});
