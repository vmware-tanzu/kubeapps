// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("../utils/kubeapps-login");
const utils = require("../utils/util-functions");

test("Deploys an Operator", async ({ page }) => {
  test.setTimeout(360000);

  // Log in
  const k = new KubeappsLogin(page);
  await k.doLogin("kubeapps-operator@example.com", "password", process.env.ADMIN_TOKEN);

  // Go to operators page
  await page.goto(utils.getUrl("/#/c/default/ns/kubeapps/operators"));
  await page.waitForSelector('h1:has-text("Operators")');
  await page.waitForFunction('document.querySelector("cds-progress-circle") === null');

  // Select operator to deploy
  await page.locator("input#search").fill("prometheus");
  await page.waitForTimeout(3000);
  // using locator with "has" instead of "hasText" to search by this exact name (and exclude others like "Red Hat, Inc.")
  await page.locator("cds-checkbox", { has: page.locator('text="Red Hat"') }).click();

  await page.click('a:has-text("prometheus")');
  await page.click('cds-button:has-text("Deploy") >> nth=0');
  await page.click('cds-button:has-text("Deploy")');

  // Wait for operators to be deployed
  await page.waitForTimeout(utils.getDeploymentTimeout());

  // Wait for the operator to be ready to be used
  await page.click('a.nav-link:has-text("Catalog")');

  // Select operators in catalog
  await page.waitForSelector('label:has-text("Operators")');
  await page.click('label:has-text("Operators")');

  // Deploy Prometheus
  await page.click('a .card-title:has-text("Prometheus")');
  await page.click('cds-button:has-text("Deploy")');
  await expect(page.locator("css=.kubeapps-main-container")).toContainText("Installation Values");

  // Update
  await page.click('cds-button:has-text("Update")');
  await page.click('cds-button:has-text("Deploy")');
  await expect(page.locator("css=.kubeapps-main-container")).toContainText("Installation Values");

  // Delete
  await page.locator('cds-button:has-text("Delete")').click();
  await page.locator('cds-modal-actions button:has-text("Delete")').click();

  await page.waitForSelector('css=h1 >> text="Applications"');
});
