// The interface in Dutch: chosen once, in force at once, and still in force
// on another machine. One pass through the screens a content creator uses,
// which is how a string nobody moved fails for us rather than for a
// customer.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

async function choose(page, chatcms, name) {
  await page.goto(`${chatcms.url}/preferences`);
  await page.getByRole("radio", { name }).click();
}

test("the interface is read in Dutch, and follows to another machine", async ({
  signedIn: page,
  chatcms,
  browser,
}, testInfo) => {
  await test.step("Dutch is chosen, and the screen changes at once", async () => {
    await choose(page, chatcms, /Nederlands/);

    await expect(page.getByRole("heading", { name: "Voorkeuren" })).toBeVisible();
    await expect.poll(() => page.locator("html").getAttribute("lang")).toBe("nl");
  });

  await test.step("every screen a content creator uses is in Dutch", async () => {
    await page.goto(`${chatcms.url}/projects`);
    await expect(page.getByRole("heading", { name: "Projecten" })).toBeVisible();
    await expect(page.getByText(/nog niet gepubliceerd/).first()).toBeVisible();

    // The phone shell has no heading of its own, so this is the screen
    // itself rather than the frame around it.
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content`);
    await expect(page.getByPlaceholder("Een pagina zoeken")).toBeVisible();
    await expect(page.getByRole("button", { name: /^Nieuw / }).first()).toBeVisible();

    await page.goto(`${chatcms.url}/o/technative/p/korenaar/media`);
    await expect(page.getByRole("button", { name: "Afbeeldingen toevoegen" }).first()).toBeVisible();

    await page.goto(`${chatcms.url}/o/technative/p/korenaar/publish`);
    await expect(page.getByLabel("Wat zit er in deze update?")).toBeVisible();

    await page.goto(`${chatcms.url}/o/technative/p/korenaar/history`);
    await expect(page.getByRole("button", { name: /Bekijken wat erin zit/ }).first()).toBeVisible();
  });

  await test.step("no English is left on those screens", async () => {
    const everything = (await page.locator("main").textContent()) ?? "";
    for (const word of ["Looking", "Publish", "Changes", "Nothing has been"]) {
      expect(everything, `"${word}" is still English`).not.toContain(word);
    }
  });

  await test.step("and it is Dutch on another machine", async () => {
    // A browser of its own: nothing is carried over but the account.
    const elsewhere = await browser.newContext();
    const other = await elsewhere.newPage();

    await other.goto(`${chatcms.url}/sign-in`);
    await other.getByLabel("Email").fill("owner@example.org");
    await other.getByLabel("Password").fill("a long enough passphrase");
    await other.getByRole("button", { name: /Sign in/i }).click();
    await other.waitForURL(/projects/);

    await expect(other.getByRole("heading", { name: "Projecten" })).toBeVisible();

    await elsewhere.close();
  });

  await worksOnAPhone(page, testInfo);

  // Back to English, so that the rest of the suite reads what it expects.
  await choose(page, chatcms, /Read everything in English|Alles in het Engels/);
});

test("what the server refused is said in Dutch too", async ({ signedIn: page, chatcms }) => {
  await choose(page, chatcms, /Nederlands/);

  await page.goto(`${chatcms.url}/account`);
  await page.getByLabel("Je wachtwoord nu").fill("niet het juiste");
  await page.getByLabel("Nieuw wachtwoord").fill("een lang genoeg wachtwoord");
  await page.getByRole("button", { name: "Mijn wachtwoord wijzigen" }).click();

  // The server says this in English and names it; the screen says the name.
  await expect(page.getByRole("alert")).toContainText("Dat is niet je huidige wachtwoord.");

  await choose(page, chatcms, /Alles in het Engels/);
});

test("a date is written in the language that is in force", async ({
  signedIn: page,
  chatcms,
}) => {
  await choose(page, chatcms, /Nederlands/);

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/history`);

  // The seeded history is from today, so this is the clock rather than a
  // day; what matters is that no English weekday or month is in it.
  const everything = (await page.locator("main").textContent()) ?? "";
  for (const word of ["Yesterday", "Monday", "January", "min ago"]) {
    expect(everything, `"${word}" is still English`).not.toContain(word);
  }

  await choose(page, chatcms, /Alles in het Engels/);
});
