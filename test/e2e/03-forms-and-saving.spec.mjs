// Acceptance test 3: all forms from the model render, and saving gives the
// same front matter as the old Quiqr.
import { test, expect, saves, worksOnAPhone } from "./fixtures.mjs";

test("every kind of field is drawn, and saving changes only what was typed", async ({ signedIn: page, chatcms }, testInfo) => {
  // The seeded site has work waiting, so the published file is not what the
  // editor is editing. Publishing first makes the two the same, which is
  // what lets the comparison afterwards mean anything.
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/publish`);
  await page.getByRole("button", { name: /^Publish \d+ changes?$/ }).click();
  await expect(page.getByText("Your site is up to date")).toBeVisible({ timeout: 60_000 });

  const before = chatcms.onTheRemote(chatcms.projectId(), "cat-file", "blob", "refs/heads/main:content/_index.md");

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);

  await test.step("every kind of field the model declares is drawn", async () => {
    await expect(page.getByRole("textbox", { name: "Page title" })).toHaveValue("Korenaar Bakery");
    await expect(page.getByRole("textbox", { name: "Introduction" })).toHaveValue(/pears/);
    await expect(page.getByRole("checkbox", { name: "Open today" })).toBeChecked();
    await expect(page.getByRole("combobox", { name: "Shop style" })).toHaveValue("traditional");
    await expect(page.getByRole("textbox", { name: "Specialities 1" })).toHaveValue("Sourdough");
    await expect(page.getByRole("textbox", { name: "Day" }).first()).toHaveValue("Monday");
    await expect(page.getByRole("textbox", { name: "Description" })).toHaveValue(/Oude Gracht/);
    await expect(page.getByRole("textbox", { name: "Page text" })).toBeVisible();
  });

  await test.step("there is no Save button, because saving happens by itself", async () => {
    await expect(page.getByRole("button", { name: /^Save/ })).toHaveCount(0);
  });

  await test.step("what is typed is saved", async () => {
    await saves(page, () =>
      page.getByRole("textbox", { name: "Page title" }).fill("Korenaar Bakery & Co"),
    );

    await page.reload();
    await expect(page.getByRole("textbox", { name: "Page title" })).toHaveValue("Korenaar Bakery & Co");
  });

  await test.step("the file differs only in what was typed", async () => {
    // Published so that the remote carries the save, then compared line for
    // line: anything else changing would be a change nobody asked for.
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/publish`);
    await page.getByRole("button", { name: /^Publish \d+ changes?$/ }).click();
    await expect(page.getByText("Your site is up to date")).toBeVisible({ timeout: 60_000 });

    const after = chatcms.onTheRemote(chatcms.projectId(), "cat-file", "blob", "refs/heads/main:content/_index.md");
    const wanted = before.replace("title: Korenaar Bakery\n", "title: Korenaar Bakery & Co\n");
    expect(after).toBe(wanted);
  });

  await worksOnAPhone(page, testInfo);
});
