// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("../utils/kubeapps-login");
const utils = require("../utils/util-functions");

test.describe("Limited user simple deployments", () => {
  test("Regular user fails to deploy package due to missing permissions", async ({ page }) => {
    // Log in
    const k = new KubeappsLogin(page);
    await k.doLogin("kubeapps-user@example.com", "password", process.env.VIEW_TOKEN);

    // Change to user's namespace using UI
    await page.click(".kubeapps-dropdown .kubeapps-nav-link");
    await page.selectOption('select[name="namespaces"]', "default");
    await page.click('cds-button:has-text("Change Context")');

    // Select package to deploy
    await page.click('a.nav-link:has-text("Catalog")');
    await page.click('a .card-title:has-text("apache")');
    await page.click('cds-button:has-text("Deploy") >> nth=0');

    // Deploy package
    const releaseNameLocator = page.locator("#releaseName");
    await releaseNameLocator.waitFor();
    await expect(releaseNameLocator).toHaveText("");
    await releaseNameLocator.fill(utils.getRandomName("test-01-release"));
    await page.locator('cds-button:has-text("Deploy")').click();

    // Assertions
    await page.waitForSelector(".alert-items .alert-text");
    const errorLocator = page.locator(".alert-items .alert-text");
    await expect(errorLocator).toHaveCount(1);
    await page.waitForTimeout(5000);

    // For some reason, UI is showing different error messages randomly
    // Custom assertion logic
    const errorMsg = await errorLocator.textContent();
    console.log(`Error message on UI = "${errorMsg}"`);

    await page.waitForFunction(msg => {
      return msg.indexOf("secrets is forbidden") > -1 || msg.indexOf("unable to read secret") > -1;
    }, errorMsg);
  });

  test("Regular user fails to deploy package in its own namespace from repos with secret", async ({
    page,
  }) => {
    // Explanation: User has permissions to deploy in its namespace, but can't actually deploy
    // if the package is from a repo that has a secret to which the user doesn't have access

    // Log in
    const k = new KubeappsLogin(page);
    await k.doLogin("kubeapps-user@example.com", "password", process.env.VIEW_TOKEN);

    // Change to user's namespace using UI
    await page.click(".kubeapps-dropdown .kubeapps-nav-link");
    await page.selectOption('select[name="namespaces"]', "kubeapps-user-namespace");
    await page.click('cds-button:has-text("Change Context")');

    // Select package to deploy
    await page.click('a.nav-link:has-text("Catalog")');
    await page.click('a .card-title:has-text("apache")');
    await page.click('cds-button:has-text("Deploy") >> nth=0');

    // Deploy package
    const releaseNameLocator = page.locator("#releaseName");
    await releaseNameLocator.waitFor();
    await expect(releaseNameLocator).toHaveText("");
    await releaseNameLocator.fill(utils.getRandomName("test-01-release"));
    await page.locator('cds-button:has-text("Deploy")').click();

    // Assertions
    await page.waitForSelector(".alert-items .alert-text");
    const errorLocator = page.locator(".alert-items .alert-text");
    await expect(errorLocator).toHaveCount(1);
    await page.waitForTimeout(5000);

    // For some reason, UI is showing different error messages randomly
    // Custom assertion logic
    const errorMsg = await errorLocator.textContent();
    console.log(`Error message on UI = "${errorMsg}"`);
    await expect(errorLocator).toContainText("unable to read secret");
  });
});
