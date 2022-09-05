// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("../utils/kubeapps-login");
const utils = require("../utils/util-functions");

test("Deploys package with default values in main cluster", async ({ page }) => {
  // Log in
  const k = new KubeappsLogin(page);
  await k.doLogin("kubeapps-operator@example.com", "password", process.env.ADMIN_TOKEN);

  // Switch to user's namespace using UI
  await page.click(".kubeapps-dropdown .kubeapps-nav-link");
  await page.selectOption('select[name="namespaces"]', "kubeapps-user-namespace");
  await page.click('cds-button:has-text("Change Context")');

  // Select package to deploy
  await page.click('a.nav-link:has-text("Catalog")');
  await page.locator("input#search").fill("carvel");
  await page.waitForTimeout(3000);
  await page.click('a:has-text("Carvel package for testing installation")');
  await page.click('cds-button:has-text("Deploy") >> nth=0');

  // Deploy package
  await page.waitForSelector('select[name="package-versions"]');
  const versionSelector = await page.locator('select[name="package-versions"]');
  await versionSelector?.selectOption("2.0.0");
  await expect(versionSelector).toHaveValue("2.0.0", { timeout: 5000 });
  const releaseNameLocator = page.locator("#releaseName");
  await releaseNameLocator.waitFor();
  await expect(releaseNameLocator).toHaveText("");
  const releaseName = utils.getRandomName("test-04-release");
  await releaseNameLocator.fill(releaseName);
  await page.selectOption("#serviceaccount-selector select", "carvel-reconciler");
  await page.locator('cds-button:has-text("Deploy")').click();
  console.log(`Creating release "${releaseName}"`);

  // Assertions
  await page.waitForSelector("css=.application-status-pie-chart-number >> text=3", {
    timeout: utils.getDeploymentTimeout(),
  });
  await page.waitForSelector("css=.application-status-pie-chart-title >> text=Ready", {
    timeout: utils.getDeploymentTimeout(),
  });

  // Clean up
  await page.locator('cds-button:has-text("Delete")').click();
  await page.locator('cds-modal-actions button:has-text("Delete")').click();
});
