// Acting as somebody else: what they see, what they may do, and a trail
// that says both names in ChatCMS and in git.
import { test, expect, saves, worksOnAPhone } from "./fixtures.mjs";

async function actAs(page, chatcms, email) {
  await page.goto(`${chatcms.url}/administration/people/${encodeURIComponent(email)}`);
  await page.getByRole("button", { name: /^Act as / }).click();
  await page.waitForURL(/projects/);
}

test("an administrator acts as a customer, and the trail says both", async ({
  administering: page,
  chatcms,
}, testInfo) => {
  await actAs(page, chatcms, "marieke@korenaar.example");

  await test.step("a band says so, and the administration screen is gone", async () => {
    await expect(page.getByRole("status")).toContainText("acting as");
    await expect(page.getByRole("status")).toContainText("marieke@korenaar.example");
    await expect(page.getByRole("link", { name: "Administration" })).toHaveCount(0);

    // The band is a screen like any other: it has to fit, and the way out
    // has to be hittable with a thumb.
    await worksOnAPhone(page, testInfo);
  });

  await test.step("what she may do is what they may do", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
    await saves(page, async () => {
      await page.getByRole("textbox", { name: "Page title" }).fill("Korenaar Bakkerij");
    });
  });

  await test.step("the site's own history names both", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/history`);
    await expect(page.getByText(/Pim Snel as /).first()).toBeVisible();
  });

  await test.step("and so does git, so the trail outlives ChatCMS", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/publish`);
    await page.getByRole("button", { name: /^Publish \d+ changes?$/ }).click();
    await expect(page.getByText("Your site is up to date")).toBeVisible({ timeout: 60_000 });

    const authors = chatcms.onTheRemote(
      chatcms.projectId(),
      "log",
      "-20",
      "--format=%an <%ae>",
      "refs/heads/main",
    );
    expect(authors).toContain("Pim Snel as ");
    expect(authors).toContain("admin@example.org");
  });

  await test.step("being yourself again brings administration back", async () => {
    await page.getByRole("button", { name: "Be myself again" }).click();
    await page.waitForURL(/administration/);

    await expect(page.getByRole("status")).toHaveCount(0);
    await expect(page.getByRole("heading", { name: "People" })).toBeVisible();
  });

  await worksOnAPhone(page, testInfo);
});

test("beginning and ending are both written down", async ({ administering: page, chatcms }) => {
  await actAs(page, chatcms, "sam@korenaar.example");
  await page.getByRole("button", { name: "Be myself again" }).click();
  await page.waitForURL(/administration/);

  // The history is fetched after the screen is drawn, so this waits for it
  // rather than reading the page while it still says it is looking.
  await expect(page.getByText(/began acting as somebody/).first()).toBeVisible();
  await expect(page.getByText(/stopped acting as somebody/).first()).toBeVisible();
});

test("an administrator cannot be acted as", async ({ administering: page, chatcms }) => {
  await page.goto(`${chatcms.url}/administration/people/admin%40example.org`);

  // The screen does not offer it, because acting as an administrator is a
  // way of becoming one.
  await expect(page.getByRole("button", { name: /^Act as / })).toHaveCount(0);

  const refused = await page.request.post(
    `${chatcms.url}/api/administration/people/admin%40example.org/act-as`,
    { failOnStatusCode: false },
  );
  expect(refused.status(), "an administrator was acted as").not.toBe(204);
});

test("somebody who does not administer cannot act as anybody", async ({
  signedIn: page,
  chatcms,
}) => {
  const refused = await page.request.post(
    `${chatcms.url}/api/administration/people/marieke%40korenaar.example/act-as`,
    { failOnStatusCode: false },
  );
  expect(refused.status(), "an owner acted as one of their people").toBe(403);
});
