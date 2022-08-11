// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// playwright.config.js
// @ts-check
const { devices } = require("@playwright/test");

/** @type {import('@playwright/test').PlaywrightTestConfig} */
const config = {
  globalSetup: require.resolve("./global-setup"),
  globalTimeout:
    (process.env.CI_TIMEOUT_MINUTES ? parseInt(process.env.CI_TIMEOUT_MINUTES) : 10) * 60 * 1000, // Global timeout for the whole execution
  timeout:
    (process.env.TEST_TIMEOUT_MINUTES ? parseInt(process.env.TEST_TIMEOUT_MINUTES) : 4) * 60 * 1000, // Default timeout for each test
  retries: 2,
  use: {
    headless: true,
    viewport: { width: 1280, height: 720 },
    ignoreHTTPSErrors: true,
    video: "retain-on-failure",
    screenshot: "only-on-failure",
    trace: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  reporter: [["list"], ["html", { open: "never", outputFolder: "reports/html" }]],
  outputDir: "reports/tests",
};

module.exports = config;
