// Copyright 2022-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("../utils/kubeapps-login");
const utils = require("../utils/util-functions");

test("Upgrades an application", async ({ page }) => {
  test.setTimeout(360000);
  const deployTimeout = utils.getDeploymentTimeout();

  // Log in
  const k = new KubeappsLogin(page);
  await k.doLogin("kubeapps-operator@example.com", "password", process.env.EDIT_TOKEN);

  // Go to catalog
  await page.click('a.nav-link:has-text("Catalog")');
  await page.click('.filters-menu label:has-text("bitnami")');
  await page.waitForSelector("input#search");
  await page.locator("input#search").fill("apache");
  await page.waitForTimeout(3000);

  // Select package
  await page.click('.card-title:has-text("kubeapps-apache")');

  // Select an older version to be installed
  await page.waitForSelector('select[name="package-versions"]');
  const defaultVersion = await page.inputValue('select[name="package-versions"]');
  await page.selectOption('select[name="package-versions"]', "8.6.2");
  const olderVersion = await page.inputValue('select[name="package-versions"]');
  expect(defaultVersion).not.toBe(olderVersion);

  // Deploy package
  await page.click('cds-button:has-text("Deploy") >> nth=0');

  // Increase replicas
  await page.locator('input[id^="replicaCount_text"]').fill("2");

  // Wait until changes are applied (due to the debounce in the input)
  await page.waitForTimeout(1000);
  await page.locator('li:has-text("YAML editor")').click();
  await page.waitForTimeout(1000);

  // Use the built-in search function in monaco to find the text we are looking for
  // so that it get loaded in the DOM when using the toContainText assert
  await page.locator(".values-editor div.modified").click({ button: "right" });
  await page.locator("text=Command Palette").click();
  await page.getByLabel("input").click();
  await page.getByLabel("input").fill(">find");
  await page
    .locator("div")
    .filter({ hasText: /^Find$/ })
    .nth(1)
    .click();
  await page.getByPlaceholder("Find").fill("replicaCount: ");
  await expect(page.locator(".values-editor div.modified")).toContainText("replicaCount: 2");

  // Set release name
  const releaseNameLocator = page.locator("#releaseName");
  await releaseNameLocator.waitFor();
  await expect(releaseNameLocator).toHaveText("");
  const releaseName = utils.getRandomName("test-07-upgrade");
  console.log(`Creating release "${releaseName}"`);
  await releaseNameLocator.fill(releaseName);

  // Trigger deploy
  await page.locator('cds-button:has-text("Deploy")').click();

  await page.click('cds-button:has-text("Upgrade")');

  await expect(page.locator(".left-menu")).toContainText(olderVersion, { timeout: deployTimeout });
  await page.selectOption('select[name="package-versions"]', defaultVersion);
  await page.waitForTimeout(2000);
  const newSelection = await page.inputValue('select[name="package-versions"]');
  expect(newSelection).toBe(defaultVersion);

  await page.locator('cds-button:has-text("Deploy")').click();

  // Check upgrade result
  // Wait for the app to be deployed and select it from "Applications"
  await expect(page.locator(".left-menu")).toContainText("Up to date", { timeout: deployTimeout });
  await page.waitForTimeout(5000);
  await page.click('a.nav-link:has-text("Applications")');
  await page.waitForTimeout(3000); // Sometimes typing was too fast to get the result shown
  await page.locator("input#search").fill(releaseName);
  await page.waitForTimeout(3000);
  await page.click(`a .card-title:has-text("${releaseName}")`);
  await page.waitForSelector("css=.application-status-pie-chart-number >> text=2", {
    timeout: deployTimeout,
  });
  await page.waitForSelector("css=.application-status-pie-chart-title >> text=Ready", {
    timeout: deployTimeout,
  });

  // Clean up
  await page.locator('cds-button:has-text("Delete")').click();
  await page.locator('cds-modal-actions cds-button:has-text("Delete")').click();
});
