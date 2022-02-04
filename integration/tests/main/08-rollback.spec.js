// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("../utils/kubeapps-login");
const utils = require("../utils/util-functions");

test("Rolls back an application", async ({ page }) => {
  test.setTimeout(360000);

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

  // Go to Deploy package
  await page.click('cds-button:has-text("Deploy") >> nth=0');

  // Set release name
  const releaseNameLocator = page.locator("#releaseName");
  await releaseNameLocator.waitFor();
  await expect(releaseNameLocator).toHaveText("");
  const releaseName = utils.getRandomName("test-08-rollback");
  console.log(`Creating release "${releaseName}"`);
  await releaseNameLocator.fill(releaseName);

  // Trigger deploy
  await page.locator('cds-button:has-text("Deploy")').click();

  // Check deployment
  await page.waitForSelector("css=.application-status-pie-chart-number >> text=1", {
    timeout: utils.getDeploymentTimeout(),
  });
  await page.waitForSelector("css=.application-status-pie-chart-title >> text=Ready", {
    timeout: utils.getDeploymentTimeout(),
  });

  // Try rollback
  await page.locator('cds-button:has-text("Rollback")').click();
  await expect(page.locator("cds-modal-content")).toContainText(
    "The application has not been upgraded, it's not possible to rollback.",
  );
  await page.locator('cds-button:has-text("Cancel")').click();

  // Upgrade the app to get another revision
  await page.locator('cds-button:has-text("Upgrade")').click();

  // Increase replicas
  await page.locator("input[type='number']").fill("2");
  await page.click('li:has-text("Changes")');
  await expect(page.locator("section#deployment-form-body-tabs-panel2")).toContainText(
    "replicaCount: 2",
  );
  await page.locator('cds-button:has-text("Deploy")').click();

  // Check upgrade result
  await page.waitForSelector("css=.application-status-pie-chart-number >> text=2", {
    timeout: utils.getDeploymentTimeout(),
  });
  await page.waitForSelector("css=.application-status-pie-chart-title >> text=Ready", {
    timeout: utils.getDeploymentTimeout(),
  });

  //  Rollback to the previous revision (default selected value)
  await page.locator('cds-button:has-text("Rollback")').click();
  await page.waitForSelector("cds-select#revision-selector");
  await expect(page.locator("cds-select#revision-selector cds-control-message")).toContainText(
    "(current: 2)",
  );
  await page.locator('cds-modal-actions cds-button:has-text("Rollback")').click();
  await page.waitForTimeout(5000);

  // Check revision and rollback to a revision (manual selected value)
  await page.locator('cds-button:has-text("Rollback")').click();
  await page.waitForSelector("cds-select#revision-selector");
  await expect(page.locator("cds-select#revision-selector cds-control-message")).toContainText(
    "(current: 3)",
  );
  await page.selectOption("cds-select#revision-selector select", "1");
  await page.locator('cds-modal-actions cds-button:has-text("Rollback")').click();
  await page.waitForTimeout(5000);

  // Check revisions
  await page.waitForSelector("css=.application-status-pie-chart-title >> text=Ready");
  await page.locator('cds-button:has-text("Rollback")').click();
  await page.waitForSelector("cds-select#revision-selector");
  await expect(page.locator("cds-select#revision-selector cds-control-message")).toContainText(
    "(current: 4)",
  );
  await page.locator('cds-button:has-text("Cancel")').click();

  // Clean up
  await page.locator('cds-button:has-text("Delete")').click();
  await page.locator('cds-modal-actions button:has-text("Delete")').click();
});
