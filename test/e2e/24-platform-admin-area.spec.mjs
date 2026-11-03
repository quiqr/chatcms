// The platform admin area: its own shell, every organisation with what it
// holds, and one organisation opened on a screen that holds everything.
import { test, expect, asksFor, worksOnAPhone } from "./fixtures.mjs";

test("the area lists every organisation, and filters are the server's", async ({
  administering: page,
  chatcms,
}, testInfo) => {
  await test.step("the shell says where you are", async () => {
    await page.goto(`${chatcms.url}/administration`);

    // /administration is the area, and the area opens on its organisations.
    await expect(page.getByRole("heading", { name: "Organisations" })).toBeVisible();
    await expect(page.getByText("Platform admin").first()).toBeVisible();
    await expect(page.getByRole("link", { name: /Back to my projects/ })).toBeVisible();
  });

  await test.step("every organisation is there, with what it holds", async () => {
    await expect(page.getByRole("link", { name: /TechNative/ }).first()).toBeVisible();
    await expect(page.getByText(/no limit/).first()).toBeVisible();
  });

  await test.step("a second one, and the filters that tell them apart", async () => {
    await page.getByRole("button", { name: /New organisation/ }).click();
    await page.getByLabel("Name", { exact: true }).fill("Bakkerij Zuid");
    await page.getByLabel("Who administers it").fill("owner@example.org");
    await page.getByRole("button", { name: "Start it" }).click();

    await expect(page.getByRole("link", { name: /Bakkerij Zuid/ }).first()).toBeVisible();

    // Nobody administers an organisation whose owner was given it, so this
    // is about the one at a limit instead.
    await asksFor(page, "PUT", `${chatcms.url}/api/administration/organisations/technative/limits`, {
      members: 1,
      projects: null,
    });

    await page.getByRole("button", { name: "At a limit" }).click();
    await expect(page.getByRole("link", { name: /TechNative/ }).first()).toBeVisible();
    await expect(page.getByRole("link", { name: /Bakkerij Zuid/ })).toHaveCount(0);

    await page.getByRole("button", { name: "All" }).click();
    await expect(page.getByRole("link", { name: /Bakkerij Zuid/ }).first()).toBeVisible();
  });

  await worksOnAPhone(page, testInfo);
});

test("one organisation is opened, limited and deleted", async ({
  administering: page,
  chatcms,
}, testInfo) => {
  await page.goto(`${chatcms.url}/administration/organisations/technative`);

  await test.step("its limits are set with a stepper", async () => {
    await page.getByRole("checkbox", { name: "Limit sites" }).check();
    await page.getByRole("button", { name: "One more" }).last().click();
    await page.getByRole("button", { name: "Save changes" }).click();

    await expect(page.getByText("Saved.")).toBeVisible();
    await expect(page.getByText(/of 3 sites used/)).toBeVisible();
  });

  await test.step("deleting says what goes and asks for the name", async () => {
    await page.getByRole("button", { name: /Remove this organisation/ }).click();

    await expect(page.getByText(/are removed from ChatCMS/)).toBeVisible();
    await expect(page.getByText(/lose access; their accounts stay/)).toBeVisible();
    await expect(page.getByText(/keep working/)).toBeVisible();

    const yes = page.getByRole("button", { name: "Yes, delete it" });
    await expect(yes).toBeDisabled();

    await page.getByLabel(/Type .* to confirm/).fill("not the name");
    await expect(yes).toBeDisabled();

    await page.getByLabel(/Type .* to confirm/).fill("TechNative");
    await expect(yes).toBeEnabled();
  });

  await test.step("and it goes, with its sites", async () => {
    await page.getByRole("button", { name: "Yes, delete it" }).click();
    await page.waitForURL(/\/administration\/organisations$/);

    await expect(page.getByRole("link", { name: /TechNative/ })).toHaveCount(0);
  });

  await worksOnAPhone(page, testInfo);
});

test("somebody who does not administer the installation cannot reach the area", async ({
  signedIn: page,
  chatcms,
}) => {
  const refused = await page.request.get(`${chatcms.url}/api/administration/organisations`, {
    failOnStatusCode: false,
  });
  expect(refused.status(), "an organisation admin listed every organisation").toBe(403);
});
