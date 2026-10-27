// An organisation's own owner manages who belongs to it, rather than asking
// whoever runs the service every time somebody joins.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

test("an owner invites somebody to the organisation and then withdraws it", async ({
  signedIn: page,
  chatcms,
}, testInfo) => {
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/people`);

  await test.step("the screen shows both levels", async () => {
    await expect(page.getByRole("radio", { name: "This project" })).toBeVisible();
    await expect(page.getByRole("radio", { name: "Organisation" })).toBeVisible();
  });

  await test.step("somebody is invited to the organisation", async () => {
    await page.getByRole("radio", { name: "Organisation" }).click();

    await page.getByLabel("Invite someone").fill("nieuw@korenaar.example");
    await page.getByLabel("Role for the person you invite").selectOption("member");
    await page.getByRole("button", { name: /Send invite/i }).click();

    await expect(page.getByText("nieuw@korenaar.example").first()).toBeVisible();
    await expect(page.getByText("Still to accept").first()).toBeVisible();
  });

  await test.step("they were sent a way in", async () => {
    const message = await chatcms.waitForAMessage("nieuw@korenaar.example");
    expect(message.Text).toContain("set-password?token=");
  });

  await test.step("the invitation is withdrawn, and they are gone", async () => {
    await page.getByRole("button", { name: /Actions for nieuw@korenaar.example/ }).click();
    await page.getByText("Withdraw the invitation").click();

    await expect(page.getByText("nieuw@korenaar.example")).toHaveCount(0);
  });

  await worksOnAPhone(page, testInfo);
});

test("the project level still shows who may edit this site", async ({ signedIn: page, chatcms }) => {
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/people`);

  // The two levels are different sets of people, and switching must not
  // show one where the other belongs.
  await page.getByRole("radio", { name: "Organisation" }).click();
  await expect(page.getByLabel("Role for the person you invite")).toHaveValue("member");

  await page.getByRole("radio", { name: "This project" }).click();
  await expect(page.getByLabel("Role for the person you invite")).toHaveValue("editor");
});

test("somebody who merely belongs cannot manage the organisation", async ({
  signedIn: page,
  chatcms,
}) => {
  const refused = await page.request.post(
    `${chatcms.url}/api/organisations/studio-noord/people`,
    { data: { email: "sneaky@example.org", role: "owner" }, failOnStatusCode: false },
  );

  // The owner of technative has no part in studio-noord beyond the seed, so
  // the answer is either a refusal or nothing there; never a new owner.
  expect([403, 404]).toContain(refused.status());
});
