// Copyright 2022-2023 the Kubeapps contributors.
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
  await page.locator('cds-modal-actions cds-button:has-text("Delete")').click();
});
