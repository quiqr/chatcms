// Mail: an invitation is sent, arrives, and the link in it works. This is
// acceptance test 1's other half, which until now ended at "the link was
// written to the log".
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

test("an invitation arrives as a message, and the link in it sets a password", async ({
  signedIn: page,
  chatcms,
}, testInfo) => {
  const invited = "nieuw@korenaar.example";

  await test.step("somebody is invited", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/people`);
    await page.getByLabel("Invite someone").fill(invited);
    await page.getByLabel("Role for the person you invite").selectOption("editor");
    await page.getByRole("button", { name: /Send invite/i }).click();

    await expect(page.getByText(invited).first()).toBeVisible();
  });

  let link;

  await test.step("the message arrives, and reads for somebody who is not technical", async () => {
    const message = await chatcms.waitForAMessage(invited);

    expect(message.Subject).toMatch(/invited/i);
    expect(message.Text).toContain("set-password?token=");

    // The copy rules: an editor is never shown a git word.
    for (const word of ["commit", "branch", "repository", "deploy", "push", "merge"]) {
      expect(
        message.Text.toLowerCase(),
        `the invitation says "${word}"`,
      ).not.toContain(word);
    }

    link = message.Text.match(/https?:\/\/\S*set-password\?token=\S+/)?.[0];
    expect(link, "the message carries no link").toBeTruthy();
  });

  await test.step("the link sets a password and signs them in", async () => {
    await page.context().clearCookies();
    await page.goto(link);

    await page.getByLabel("New password").fill("another long enough passphrase");
    await page.getByRole("button", { name: /Set .*password|Save/i }).click();

    await page.waitForURL(/projects|p\//);
    // Waited for the destination to be drawn, so that the phone rules below
    // are checked against a screen rather than against a navigation.
    await expect(page.locator("main")).toBeVisible();
  });

  await worksOnAPhone(page, testInfo);
});

test("asking for a new password sends one message and says nothing either way", async ({
  page,
  chatcms,
}) => {
  await page.goto(`${chatcms.url}/forgot-password`);

  await page.getByLabel("Email").fill("owner@example.org");
  await page.getByRole("button", { name: /Send|Reset/i }).click();

  const message = await chatcms.waitForAMessage("owner@example.org");
  expect(message.Subject).toMatch(/password/i);
  expect(message.Text).toContain("set-password?token=");

  // An address nobody holds is answered the same way, so that the screen
  // cannot be used to find out who has an account.
  const said = await page.locator("body").innerText();

  await page.goto(`${chatcms.url}/forgot-password`);
  await page.getByLabel("Email").fill("nobody@korenaar.example");
  await page.getByRole("button", { name: /Send|Reset/i }).click();

  await expect(page.locator("body")).toContainText(said.split("\n").filter(Boolean).at(-1) ?? "");
});
