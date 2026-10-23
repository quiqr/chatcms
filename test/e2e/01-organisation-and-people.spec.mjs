// Acceptance test 1: as admin, create an organisation and invite the
// customer as editor.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

test("a project can be made and somebody invited to it as an editor", async ({ signedIn: page, chatcms }, testInfo) => {
  await test.step("the organisation's projects are listed", async () => {
    await expect(page.getByRole("link", { name: /Korenaar/ })).toBeVisible();
  });

  await test.step("a new project is made and goes to its setup", async () => {
    await page.getByRole("button", { name: /New project/ }).click();

    const asking = page.getByRole("dialog", { name: "New project" });
    await asking.getByLabel("Name").fill("Studio Zuid");
    await expect(asking).toContainText("/p/studio-zuid");

    await asking.getByRole("button", { name: /Make it and set it up/ }).click();
    await page.waitForURL(/\/p\/studio-zuid\/setup/);
  });

  await test.step("somebody is invited as an editor", async () => {
    await page.goto(`${chatcms.url}/p/korenaar/people`);

    await page.getByLabel("Invite someone").fill("marieke@korenaar.example");
    await page.getByLabel("Role for the person you invite").selectOption("editor");
    await page.getByRole("button", { name: /Send invite/i }).click();

    await expect(page.getByText("marieke@korenaar.example").first()).toBeVisible();
  });

  await worksOnAPhone(page, testInfo);
});
