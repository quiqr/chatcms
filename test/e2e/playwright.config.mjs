import { defineConfig, devices } from "@playwright/test";

// Every acceptance test runs twice: once as somebody at a desk and once as
// somebody on a phone, because "works on a phone" is itself one of the
// acceptance tests and a separate pass would let the two drift.
export default defineConfig({
  testDir: ".",
  fullyParallel: false,
  workers: 1,
  forbidOnly: true,
  retries: 0,
  reporter: process.env.CI ? [["list"]] : [["list"]],
  timeout: 120_000,
  expect: { timeout: 20_000 },
  use: {
    actionTimeout: 20_000,
    navigationTimeout: 30_000,
  },
  projects: [
    {
      name: "desk",
      testMatch: /.*\.spec\.mjs$/,
      use: { ...devices["Desktop Chrome"], viewport: { width: 1440, height: 900 } },
    },
    {
      name: "phone",
      testMatch: /.*\.spec\.mjs$/,
      use: { ...devices["Desktop Chrome"], viewport: { width: 390, height: 844 }, isMobile: false },
    },
    {
      name: "shots-desk",
      testMatch: /.*\.shots\.mjs$/,
      use: { ...devices["Desktop Chrome"], viewport: { width: 1440, height: 900 } },
    },
    {
      name: "shots-phone",
      testMatch: /.*\.shots\.mjs$/,
      use: { ...devices["Desktop Chrome"], viewport: { width: 390, height: 844 }, isMobile: false },
    },
  ],
});
