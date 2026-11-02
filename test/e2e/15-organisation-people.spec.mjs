// An organisation's own admin manages who belongs to it, rather than asking
// whoever runs the platform every time somebody joins.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

test("an organisation admin invites somebody and then takes them out", async ({
  signedIn: page,
  chatcms,
}, testInfo) => {
  // Who belongs to an organisation is managed from the organisation, which
  // is its own area rather than a tab on one of its projects.
  await page.goto(`${chatcms.url}/o/technative/manage/members`);

  await test.step("somebody is invited to the organisation", async () => {
    await page.getByLabel("Email").fill("nieuw@korenaar.example");
    await page.getByRole("button", { name: "Send invite" }).click();

    await expect(page.getByText("nieuw@korenaar.example")).toBeVisible();
    await expect(page.getByText("Still to accept").first()).toBeVisible();
  });

  await test.step("they were sent a way in", async () => {
    const message = await chatcms.waitForAMessage("nieuw@korenaar.example");
    expect(message.Text).toContain("set-password?token=");
  });

  await test.step("and taken out again", async () => {
    await page.getByRole("button", { name: /Take nieuw@korenaar.example out/ }).click();
    await expect(page.getByText("nieuw@korenaar.example")).toHaveCount(0);
  });

  await worksOnAPhone(page, testInfo);
});

test("the people of a site are the site's own", async ({ signedIn: page, chatcms }) => {
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/people`);

  // The organisation half of this screen moved to the organisation's area,
  // so what is offered here is a role on this site.
  await expect(page.getByLabel("Role for the person you invite")).toHaveValue("editor");
  await expect(page.getByRole("radio", { name: "Organisation" })).toHaveCount(0);
});

test("somebody who merely belongs cannot manage the organisation", async ({
  signedIn: page,
  chatcms,
}) => {
  const refused = await page.request.post(
    `${chatcms.url}/api/organisations/studio-noord/people`,
    { data: { email: "sneaky@example.org", role: "admin" }, failOnStatusCode: false },
  );

  // The admin of technative has no part in studio-noord beyond the seed, so
  // the answer is either a refusal or nothing there; never a new admin.
  expect([403, 404]).toContain(refused.status());
});
