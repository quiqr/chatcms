// A site as a whole: copied, moved to another organisation, downloaded and
// put out of the way.
import { test, expect, startsAnOrganisation, worksOnAPhone } from "./fixtures.mjs";
import { owner } from "./harness.mjs";

test("a site is duplicated, moved, archived and downloaded", async ({
  signedIn: page,
  administering: admin,
  chatcms,
}, testInfo) => {
  await test.step("there is a second organisation to move into", async () => {
    await startsAnOrganisation(admin, chatcms, "Stichting Broodnodig", owner.email);
    await expect(admin.getByText("Stichting Broodnodig")).toBeVisible();
  });

  await page.goto(`${chatcms.url}/o/technative/manage/projects`);

  await test.step("a copy needs a name of its own", async () => {
    await page.getByRole("button", { name: /What can be done with Korenaar Bakery/ }).click();
    await page.getByRole("menuitem", { name: /Duplicate/ }).click();

    const name = page.getByLabel("Name of the copy");
    await expect(name).toHaveValue("Korenaar Bakery");
    await expect(page.getByRole("button", { name: "Duplicate site" })).toBeDisabled();
    await expect(page.getByText(/already taken/)).toBeVisible();

    await name.fill("Korenaar Winkel");
    await page.getByRole("button", { name: "Duplicate site" }).click();

    await expect(page.getByText("Korenaar Winkel")).toBeVisible();
    await expect(page.getByText("Korenaar Bakery")).toBeVisible();
  });

  await test.step("the copy moves to the other organisation", async () => {
    await page.getByRole("button", { name: /What can be done with Korenaar Winkel/ }).click();
    await page.getByRole("menuitem", { name: /Move to another organisation/ }).click();

    await page.getByRole("radio", { name: /Stichting Broodnodig/ }).check();
    await page.getByRole("button", { name: "Move site" }).click();

    await expect(page.getByText("Korenaar Winkel")).toHaveCount(0);
  });

  await test.step("and is there afterwards", async () => {
    await page.goto(`${chatcms.url}/o/stichting-broodnodig/manage/projects`);
    await expect(page.getByText("Korenaar Winkel")).toBeVisible();
  });

  await page.goto(`${chatcms.url}/o/technative/manage/projects`);

  await test.step("archiving puts a site out of the way and back again", async () => {
    await page.getByRole("button", { name: /What can be done with Korenaar Bakery/ }).click();
    await page.getByRole("menuitem", { name: "Archive" }).click();

    await expect(page.getByRole("heading", { name: /Archived · 1/ })).toBeVisible();
    await expect(page.getByText(/count toward the limit/)).toBeVisible();

    await page.getByRole("button", { name: "Bring it back" }).click();

    await expect(page.getByRole("heading", { name: /Archived/ })).toHaveCount(0);
    await expect(page.getByRole("link", { name: /Open/ }).first()).toBeVisible();
  });

  await test.step("the whole site comes down as a zip", async () => {
    const link = page.getByRole("menuitem", { name: "Download as zip" });
    await page.getByRole("button", { name: /What can be done with Korenaar Bakery/ }).click();
    await expect(link).toHaveAttribute("href", /\/projects\/korenaar\/zip$/);
    await page.keyboard.press("Escape");

    const zip = await page.request.get(
      `${chatcms.url}/api/organisations/technative/projects/korenaar/zip`,
      { failOnStatusCode: false },
    );

    expect(zip.status(), `the zip was refused:\n${chatcms.log}`).toBe(200);
    expect(zip.headers()["content-disposition"]).toContain("korenaar.zip");
    const bytes = await zip.body();
    expect(bytes.subarray(0, 2).toString()).toBe("PK");
  });

  await worksOnAPhone(page, testInfo);
});
