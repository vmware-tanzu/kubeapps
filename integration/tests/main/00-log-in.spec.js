// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("../utils/kubeapps-login");
const utils = require("../utils/util-functions");

test.describe("Log in", () => {
  test("Logs in successfully as view user", async ({ page }) => {
    const k = new KubeappsLogin(page);
    // TODO (castelblanque) Set and use proper env variables
    // like `process.env.VIEW_USER` and `process.env.VIEW_PASSWORD` or similar
    await k.doLogin("kubeapps-user@example.com", "password", process.env.VIEW_TOKEN);

    await page.waitForSelector('css=h1 >> text="Applications"');
  });
});
