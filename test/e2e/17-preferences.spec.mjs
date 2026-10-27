// How somebody likes the application to behave is kept with the account, so
// that it follows them to another machine.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

async function theAppearance(page) {
  return page.evaluate(() => ({
    chosen: document.documentElement.getAttribute("data-appearance"),
    background: getComputedStyle(document.body).backgroundColor,
  }));
}

test("a dark appearance is chosen and follows to another browser", async ({
  signedIn: page,
  chatcms,
  browser,
}, testInfo) => {
  await page.goto(`${chatcms.url}/preferences`);

  const before = await theAppearance(page);

  await test.step("dark is chosen, and the screen changes at once", async () => {
    await page.getByRole("radio", { name: "Dark Always dark." }).click();

    await expect
      .poll(async () => (await theAppearance(page)).chosen)
      .toBe("dark");

    const after = await theAppearance(page);
    expect(after.background, "the screen did not change").not.toBe(before.background);
  });

  await test.step("and it is there on another machine", async () => {
    // A browser of its own: nothing is carried over but the account.
    const elsewhere = await browser.newContext();
    const other = await elsewhere.newPage();

    await other.goto(`${chatcms.url}/sign-in`);
    await other.getByLabel("Email").fill("owner@example.org");
    await other.getByLabel("Password").fill("a long enough passphrase");
    await other.getByRole("button", { name: /Sign in/i }).click();
    await other.waitForURL(/projects/);

    await expect.poll(async () => (await theAppearance(other)).chosen).toBe("dark");

    await elsewhere.close();
  });

  await worksOnAPhone(page, testInfo);
});

test("following the device is what somebody who chooses nothing gets", async ({
  signedIn: page,
  chatcms,
}) => {
  await page.goto(`${chatcms.url}/preferences`);

  await page.getByRole("radio", { name: "Dark Always dark." }).click();
  await expect.poll(async () => (await theAppearance(page)).chosen).toBe("dark");

  await page.getByRole("button", { name: "Follow my device again" }).click();

  // Nothing chosen means no attribute at all, so the stylesheet asks the
  // machine rather than being told.
  await expect.poll(async () => (await theAppearance(page)).chosen).toBeNull();
});

test("preferences are reachable from any screen", async ({ signedIn: page, chatcms }) => {
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content`);

  await page.getByTestId("person-menu").click();
  await page.getByText("Preferences").click();

  await page.waitForURL(/preferences/);
  await expect(page.getByRole("heading", { name: "How it looks" })).toBeVisible();
});
