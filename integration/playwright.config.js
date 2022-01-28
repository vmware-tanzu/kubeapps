// playwright.config.js
// @ts-check
const { devices } = require("@playwright/test");

/** @type {import('@playwright/test').PlaywrightTestConfig} */
const config = {
  globalTimeout: (process.env.CI_TIMEOUT ? parseInt(process.env.CI_TIMEOUT) : 10) * 60 * 1000,
  retries: 1,
  use: {
    headless: true,
    viewport: { width: 1280, height: 720 },
    ignoreHTTPSErrors: true,
    video: 'retain-on-failure',
    screenshot: 'on',
    trace: 'on',
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  reporter: [ ['html', { open: 'never', outputFolder: 'reports/' }] ],
  outputDir: 'reports/',
};

module.exports = config;
