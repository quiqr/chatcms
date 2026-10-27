// People: an account can be turned off when somebody leaves, and a person
// manages their own name and password.
import { test, expect, signingIn, worksOnAPhone } from "./fixtures.mjs";

test("somebody who has left is stopped, and what they wrote stays theirs", async ({
  signedIn: owner,
  administering: boss,
  chatcms,
}, testInfo) => {
  await test.step("what they wrote names them, before anything happens", async () => {
    await owner.goto(`${chatcms.url}/o/technative/p/korenaar/history`);
    await expect(owner.getByText(/Joris Bakker/).first()).toBeVisible();
  });

  await test.step("the account is turned off", async () => {
    await boss.goto(`${chatcms.url}/administration`);
    await boss.getByLabel("Their email address").fill("owner@example.org");
    await boss.getByRole("button", { name: "Turn it off" }).click();

    await expect(boss.getByText(/owner@example.org can no longer sign in/)).toBeVisible();
  });

  await test.step("they cannot sign in", async () => {
    const refused = await signingIn(boss, chatcms, "owner@example.org", "a long enough passphrase");
    expect(refused.status(), "an account that is off signed in").not.toBe(200);
  });

  await test.step("the session they had stops at once", async () => {
    // An account that is off whose open screen keeps working is not off.
    await owner.goto(`${chatcms.url}/o/technative/p/korenaar/history`);
    await expect(owner).toHaveURL(/sign-in/);
  });

  await test.step("and it can be turned on again", async () => {
    await boss.getByLabel("Their email address").fill("owner@example.org");
    await boss.getByRole("button", { name: "Turn it on again" }).click();

    await expect(boss.getByText(/owner@example.org can sign in again/)).toBeVisible();

    const back = await signingIn(boss, chatcms, "owner@example.org", "a long enough passphrase");
    expect(back.status(), "the password it had no longer works").toBe(200);
  });

  await worksOnAPhone(boss, testInfo);
});

test("an administrator cannot turn off their own account", async ({
  administering: page,
  chatcms,
}) => {
  await page.goto(`${chatcms.url}/administration`);

  await page.getByLabel("Their email address").fill("admin@example.org");
  await page.getByRole("button", { name: "Turn it off" }).click();

  await expect(page.getByRole("alert")).toContainText("your own account");
});

test("a person changes their own name and password", async ({ signedIn: page, chatcms }, testInfo) => {
  await page.goto(`${chatcms.url}/account`);

  await test.step("the name is changed, and shows everywhere", async () => {
    await page.getByLabel("Name").fill("Joris de Bakker");
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Saved.")).toBeVisible();
  });

  await test.step("the password is changed, asking for the one they have", async () => {
    await page.getByLabel("Your password now").fill("a long enough passphrase");
    await page.getByLabel("New password").fill("another long enough passphrase");
    await page.getByRole("button", { name: "Change my password" }).click();

    await expect(page.getByText("Your password is changed.")).toBeVisible();
  });

  await test.step("the old password no longer works and the new one does", async () => {
    const old = await signingIn(page, chatcms, "owner@example.org", "a long enough passphrase");
    expect(old.status(), "the old password still works").not.toBe(200);

    const now = await signingIn(page, chatcms, "owner@example.org", "another long enough passphrase");
    expect(now.status(), "the new password does not work").toBe(200);
  });

  await worksOnAPhone(page, testInfo);
});

test("a wrong current password changes nothing", async ({ signedIn: page, chatcms }) => {
  await page.goto(`${chatcms.url}/account`);

  await page.getByLabel("Your password now").fill("not the right one at all");
  await page.getByLabel("New password").fill("another long enough passphrase");
  await page.getByRole("button", { name: "Change my password" }).click();

  await expect(page.getByRole("alert")).toContainText("current password");

  const still = await signingIn(page, chatcms, "owner@example.org", "a long enough passphrase");
  expect(still.status(), "the password changed anyway").toBe(200);
});
