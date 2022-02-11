// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("../utils/kubeapps-login");
const utils = require("../utils/util-functions");

test("Deploys package with default values in the second cluster", async ({ page }) => {
  test.setTimeout(120000);

  // Log in
  const k = new KubeappsLogin(page);
  await k.doLogin("kubeapps-operator@example.com", "password", process.env.ADMIN_TOKEN);

  // Change cluster using ui
  await page.click(".kubeapps-dropdown .kubeapps-nav-link");
  await page.selectOption('select[name="clusters"]', "second-cluster");
  await page.click('cds-button:has-text("Change Context")');

  // Select package to deploy
  await page.click('a.nav-link:has-text("Catalog")');
  await page.locator("input#search").fill("apache");
  await page.waitForTimeout(3000);
  await page.click('a:has-text("foo apache chart for CI")');
  await page.click('cds-button:has-text("Deploy") >> nth=0');

  // Deploy package
  const releaseNameLocator = page.locator("#releaseName");
  await releaseNameLocator.waitFor();
  await expect(releaseNameLocator).toHaveText("");
  const releaseName = utils.getRandomName("test-05-release");
  console.log(`Creating release "${releaseName}"`);
  await releaseNameLocator.fill(releaseName);
  await page.locator('cds-button:has-text("Deploy")').click();

  // Assertions
  await page.waitForSelector("css=.application-status-pie-chart-number >> text=1", {
    timeout: utils.getDeploymentTimeout(),
  });
  await page.waitForSelector("css=.application-status-pie-chart-title >> text=Ready", {
    timeout: utils.getDeploymentTimeout(),
  });

  // Clean up
  await page.locator('cds-button:has-text("Delete")').click();
  await page.locator('cds-modal-actions button:has-text("Delete")').click();
});
