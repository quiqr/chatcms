// Every account on the installation, and what may be done to one.
import { test, expect, asksFor, signingIn, worksOnAPhone } from "./fixtures.mjs";

test("an account is put right, reset, blocked and deleted", async ({
  administering: page,
  chatcms,
}, testInfo) => {
  await page.goto(`${chatcms.url}/administration/users`);

  await test.step("the three states an account is in", async () => {
    // The seed has somebody who never accepted their invitation.
    await expect(page.getByText("Active").first()).toBeVisible();
    await expect(page.getByText("Invited").first()).toBeVisible();
  });

  await test.step("a name typed wrong is put right", async () => {
    await page.getByRole("button", { name: /What can be done with marieke@korenaar.example/ }).click();
    await page.getByRole("menuitem", { name: "Edit name and email" }).click();

    await page.getByLabel("Their name").fill("Marieke de Vries");
    await page.getByRole("button", { name: "Save" }).click();

    await expect(page.getByText("Marieke de Vries")).toBeVisible();
  });

  await test.step("a password reset is sent, never set", async () => {
    await page.getByRole("button", { name: /What can be done with Marieke de Vries/ }).click();
    await page.getByRole("menuitem", { name: "Send password reset" }).click();

    await expect(page.getByRole("status")).toContainText("on its way");

    const message = await chatcms.waitForAMessage("marieke@korenaar.example");
    expect(message.Text).toContain("set-password?token=");
  });

  await test.step("blocking stops them signing in", async () => {
    await page.getByRole("button", { name: /What can be done with Joris Bakker/ }).click();
    await page.getByRole("menuitem", { name: "Block sign-in" }).click();

    await expect(page.getByText("Blocked").first()).toBeVisible();

    const refused = await signingIn(page, chatcms, "owner@example.org", "a long enough passphrase");
    expect(refused.status(), "a blocked account signed in").not.toBe(200);
  });

  await test.step("and letting them in again", async () => {
    await page.getByRole("button", { name: /What can be done with Joris Bakker/ }).click();
    await page.getByRole("menuitem", { name: "Let them sign in again" }).click();

    const back = await signingIn(page, chatcms, "owner@example.org", "a long enough passphrase");
    expect(back.status(), "unblocking did not let them in").toBe(200);
  });

  await test.step("deleting asks first, and says what stays", async () => {
    await page.getByRole("button", { name: /What can be done with sam@korenaar.example/ }).click();
    await page.getByRole("menuitem", { name: /Delete user/ }).click();

    await expect(page.getByText(/What that person wrote stays/)).toBeVisible();

    await page.getByRole("button", { name: "Yes, delete the account" }).click();
    await expect(page.getByText("sam@korenaar.example")).toHaveCount(0);
  });

  await worksOnAPhone(page, testInfo);
});

test("the list is filtered and paged by the server", async ({
  administering: page,
  chatcms,
}) => {
  await page.goto(`${chatcms.url}/administration/users`);

  await test.step("the platform admins are a filter", async () => {
    await page.getByRole("button", { name: "Platform admins" }).click();

    await expect(page.getByText("admin@example.org")).toBeVisible();
    await expect(page.getByText("owner@example.org")).toHaveCount(0);
  });

  await test.step("and so is having no organisation", async () => {
    await page.getByRole("button", { name: "Not in any organisation" }).click();

    // The seeded people are all in one; whoever administers the platform is
    // in none, which is the point of administering not being membership.
    await expect(page.getByText("admin@example.org")).toBeVisible();
  });

  await test.step("the server answers the filter, not the browser", async () => {
    const answer = await asksFor(
      page,
      "GET",
      `${chatcms.url}/api/administration/people?only=blocked`,
    );
    const body = await answer.json();
    expect(body.people).toEqual([]);
  });
});

test("nobody blocks or deletes their own account", async ({ administering: page, chatcms }) => {
  await page.goto(`${chatcms.url}/administration/users`);

  await page.getByRole("button", { name: /What can be done with Pim Snel/ }).click();

  await expect(page.getByRole("menuitem", { name: "Block sign-in" })).toBeDisabled();
  await expect(page.getByRole("menuitem", { name: /Delete user/ })).toBeDisabled();

  const refused = await asksFor(
    page,
    "DELETE",
    `${chatcms.url}/api/administration/people/admin%40example.org`,
  );
  expect(refused.status(), "a platform admin deleted their own account").toBe(409);
});

test("somebody who does not administer cannot do any of it", async ({
  signedIn: page,
  chatcms,
}) => {
  for (const [method, path] of [
    ["PUT", "/api/administration/people/owner%40example.org"],
    ["DELETE", "/api/administration/people/owner%40example.org"],
    ["POST", "/api/administration/people/owner%40example.org/password-reset"],
  ]) {
    const refused = await asksFor(page, method, `${chatcms.url}${path}`, { email: "a@b.c" });
    expect(refused.status(), `${method} ${path}`).toBe(403);
  }
});
