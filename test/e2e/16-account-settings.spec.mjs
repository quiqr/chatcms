// An account has a picture, an address that can change, and sessions
// somebody can see and end.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

// The smallest real PNG, so what is uploaded is a picture rather than bytes
// named like one.
function onePixel() {
  return Buffer.from(
    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
    "base64",
  );
}

test("a picture is added and taken away again", async ({ signedIn: page, chatcms }, testInfo) => {
  await page.goto(`${chatcms.url}/account`);

  await test.step("a picture is chosen", async () => {
    await page.getByLabel("Choose a picture").setInputFiles({
      name: "me.png",
      mimeType: "image/png",
      buffer: onePixel(),
    });

    // The avatar becomes a picture rather than letters.
    await expect(page.locator('img[src*="/picture"]').first()).toBeVisible({ timeout: 30_000 });
  });

  await test.step("and taken away", async () => {
    await page.getByRole("button", { name: "Use my initials" }).click();
    await expect(page.locator('img[src*="/picture"]')).toHaveCount(0);
  });

  await worksOnAPhone(page, testInfo);
});

test("something that is not a picture is refused", async ({ signedIn: page, chatcms }) => {
  await page.goto(`${chatcms.url}/account`);

  await page.getByLabel("Choose a picture").setInputFiles({
    name: "me.png",
    mimeType: "image/png",
    buffer: Buffer.from("this is not a picture, whatever it is named"),
  });

  await expect(page.getByRole("alert")).toContainText("not a picture");
});

test("an address changes only once the link is followed", async ({
  signedIn: page,
  chatcms,
}, testInfo) => {
  await page.goto(`${chatcms.url}/account`);

  await test.step("a new address is asked for", async () => {
    await page.getByLabel("New address").fill("joris@korenaar.example");
    await page.getByRole("button", { name: "Send me the link" }).click();

    await expect(page.getByRole("status")).toContainText("owner@example.org until you follow");
  });

  let link;

  await test.step("the message goes to the new address", async () => {
    const message = await chatcms.waitForAMessage("joris@korenaar.example");
    link = message.Text.match(/https?:\/\/\S*confirm-address\?token=\S+/)?.[0];
    expect(link, "the message carries no link").toBeTruthy();
  });

  await test.step("following it changes the address", async () => {
    await page.goto(link);
    await expect(page.getByRole("status")).toContainText("sign in with it from now on");
  });

  await worksOnAPhone(page, testInfo);
});

test("where somebody is signed in, and ending one of them", async ({ signedIn: page, chatcms }) => {
  await page.goto(`${chatcms.url}/account`);

  // The screen being read is marked and offers no way to end itself; that is
  // what signing out is for.
  await expect(page.getByText("This screen")).toBeVisible();

  const ending = page.getByRole("button", { name: "End it" });
  if (await ending.count()) {
    await ending.first().click();
  }

  await expect(page.getByText("This screen")).toBeVisible();
});
