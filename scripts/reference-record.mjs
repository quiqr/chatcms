// Makes one edit in the old Quiqr, through its own form, and waits for the
// application to say it saved. Driving the form rather than its API is the
// point: the form components are the part of the old application whose
// behaviour is in question, and an API call would step over them.
//
// It clicks its way in rather than opening a form's address, because the old
// application's server answers "Not Found" to a browser asking for one: its
// routes exist only once its own code is running.
//
//   node scripts/reference-record.mjs <port> <site name> <single title> <field> <value>
import { chromium } from "../test/e2e/node_modules/playwright-core/index.mjs";

const [port, siteName, singleTitle, field, value] = process.argv.slice(2);
if (value === undefined) {
  console.error("usage: reference-record.mjs <port> <site name> <single title> <field> <value>");
  process.exit(1);
}

const browser = await chromium.launch();
const page = await browser.newPage();
const trouble = [];
page.on("console", (message) => {
  if (message.type() === "error") trouble.push(message.text().slice(0, 200));
});

async function giveUp(why) {
  console.error(`reference-record: ${why}`);
  if (trouble.length > 0) console.error(`the screen said: ${trouble.join(" | ")}`);
  await browser.close();
  process.exit(1);
}

try {
  await page.goto(`http://127.0.0.1:${port}/`, { waitUntil: "networkidle" });

  await page.getByText(siteName, { exact: true }).first().click();
  await page.waitForURL(/\/workspaces\//, { timeout: 20000 });

  await page.getByText(singleTitle, { exact: true }).first().click();
  await page.waitForURL(/\/singles\//, { timeout: 20000 });

  const control = page.getByLabel(field, { exact: true });
  await control.waitFor({ state: "visible", timeout: 20000 });

  const was = await control.inputValue();
  if (was === value) {
    await giveUp(`${field} already says ${JSON.stringify(value)}, so a save would prove nothing`);
  }

  await control.fill(value);
  await control.blur();

  // The old application's save is a floating button that stays disabled
  // until something changed, and goes back to disabled once the save is
  // accepted. That is the only signal it gives that the file is written.
  const save = page.getByRole("button", { name: "add" });
  await save.waitFor({ state: "visible", timeout: 10000 });
  if (await save.isDisabled()) {
    await giveUp("the form did not notice the change: its save button stayed disabled");
  }

  await save.click();
  for (let waited = 0; waited < 100; waited++) {
    if (await save.isDisabled()) {
      await browser.close();
      console.log(`    ${field}: ${JSON.stringify(was)} -> ${JSON.stringify(value)}`);
      process.exit(0);
    }
    await page.waitForTimeout(100);
  }
  await giveUp("the save did not finish within ten seconds");
} catch (why) {
  await giveUp(why.message);
}
