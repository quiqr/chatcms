// An organisation may be allowed so many members and so many sites. Both are
// optional and set by the platform admin, and both are refused before
// anything runs.
import { test, expect, asksFor, worksOnAPhone } from "./fixtures.mjs";

async function setLimits(page, chatcms, organisation, body) {
  const answer = await asksFor(
    page,
    "PUT",
    `${chatcms.url}/api/administration/organisations/${organisation}/limits`,
    body,
  );
  expect(answer.status(), "setting the limits").toBe(204);
}

test("a site is refused when the organisation has no room for one", async ({
  administering: boss,
  signedIn: page,
  chatcms,
}, testInfo) => {
  // The seed's organisation holds two sites; one is fewer than it has.
  await setLimits(boss, chatcms, "technative", { members: null, projects: 2 });

  await test.step("the screen says how full it is, and offers nothing", async () => {
    await page.goto(`${chatcms.url}/o/technative/projects`);

    await expect(page.getByText("2 of 2 sites used")).toBeVisible();
    await expect(page.getByText("Full")).toBeVisible();
    await expect(page.getByText(/All 2 sites are in use/)).toBeVisible();
    await expect(page.getByRole("button", { name: /New project/ })).toBeDisabled();
  });

  await test.step("and the server refuses it anyway, in its own words", async () => {
    // The screen is a courtesy; the limit is the server's.
    const refused = await asksFor(page, "POST",
      `${chatcms.url}/api/organisations/technative/projects`,
      { slug: "third", name: "A third site", domain: "" });

    expect(refused.status()).toBe(409);
    const body = await refused.json();
    expect(body.error.key).toBe("refused.no-room-for-a-site");
    expect(body.error.message).toContain("that makes 3 sites, and the limit is 2".replace("that", "That"));
  });

  await test.step("raising the limit is all it takes", async () => {
    await setLimits(boss, chatcms, "technative", { members: null, projects: 5 });

    await page.reload();
    await expect(page.getByRole("button", { name: /New project/ })).toBeEnabled();
    await expect(page.getByText("2 of 5 sites used")).toBeVisible();
  });

  await worksOnAPhone(page, testInfo);
});

test("a seat is refused when they are all held", async ({ administering: page, chatcms }) => {
  await page.goto(`${chatcms.url}/administration/organisations/technative`);

  // The seeded organisation holds one member: whoever administers it. The
  // people on its sites hold a role there rather than a seat here.
  await expect(page.getByRole("button", { name: /Take .* out/ }).first()).toBeVisible();
  await setLimits(page, chatcms, "technative", { members: 1, projects: null });

  await page.reload();

  await expect(page.getByText("1 of 1 seats used")).toBeVisible();
  await expect(page.getByText(/The one seat is in use/)).toBeVisible();
  await expect(page.getByRole("button", { name: "Put them in" })).toBeDisabled();

  // The platform admin is not above the limit they set: raising it is what
  // they may do that nobody else can.
  const refused = await asksFor(page, "POST",
    `${chatcms.url}/api/administration/organisations/technative/people`,
    { email: "sam@korenaar.example", role: "member" });
  expect(refused.status()).toBe(409);
  expect((await refused.json()).error.key).toBe("refused.no-room-for-a-member");
});

test("an organisation with no limits says nothing about room", async ({
  administering: page,
  chatcms,
}) => {
  // Nothing has been set on the seeded organisation, which is the state
  // every organisation starts in.
  await page.goto(`${chatcms.url}/administration/organisations/technative`);

  await expect(page.getByText(/no limit/).first()).toBeVisible();
  await expect(page.getByText(/seats are in use/)).toHaveCount(0);
  await expect(page.getByText(/sites are in use/)).toHaveCount(0);
});

test("somebody who is not a platform admin cannot set a limit", async ({
  signedIn: page,
  chatcms,
}) => {
  const refused = await asksFor(page, "PUT",
    `${chatcms.url}/api/administration/organisations/technative/limits`,
    { members: 99, projects: 99 });

  expect(refused.status(), "an organisation admin set their own limit").toBe(403);
});
