// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("../utils/kubeapps-login");
const utils = require("../utils/util-functions");

test("Deploy a chart using a private container image", async ({ page }) => {
  const deployTimeout = utils.getDeploymentTimeout();

  // Log in
  const k = new KubeappsLogin(page);
  await k.doLogin("kubeapps-operator@example.com", "password", process.env.ADMIN_TOKEN);

  // Change namespace to non-kubeapps
  await page.click(".kubeapps-dropdown .kubeapps-nav-link");
  await page.selectOption('select[name="namespaces"]', "default");
  await page.click('cds-button:has-text("Change Context")');

  // Go to repos page
  await page.click(".dropdown.kubeapps-menu button.kubeapps-nav-link");
  await page.locator("text=Package Repositories").click();
  await expect(page).not.toContain("text=Fetching Package Repositories...");

  // Add new repo
  const repoName = utils.getRandomName("my-repo");
  console.log(`Creating package repository "${repoName}"`);
  await page.locator("text=Add Package Repository >> div").click();
  await page.fill("input#kubeapps-repo-name", repoName);
  await page.fill(
    "input#kubeapps-repo-url",
    "http://chartmuseum.chart-museum.svc.cluster.local:8080",
  );
  await page.locator("text=Helm Charts").first().click();
  await page.locator("text=Helm Repository").click();

  // Set credentials for repository
  await page.locator("#panel-auth cds-accordion-header div >> nth=0").first().click();
  // Basic auth
  await page.locator("text=Basic Auth").click();
  await page.locator('[id="kubeapps-repo-username"]').fill("admin");
  await page.locator('[id="kubeapps-repo-password"]').fill("password");

  // Set the Docker repo credentials
  await page.locator("text=Docker Registry Credentials").nth(1).click();

  await page
    .locator('[id="kubeapps-imagePullSecrets-cred-server"]')
    .fill(process.env.DOCKER_REGISTRY_URL);
  await page
    .locator('[id="kubeapps-imagePullSecrets-cred-username"]')
    .fill(process.env.DOCKER_USERNAME);
  await page
    .locator('[id="kubeapps-imagePullSecrets-cred-password"]')
    .fill(process.env.DOCKER_PASSWORD);
  await page.locator('[id="kubeapps-imagePullSecrets-cred-email"]').fill("test@example.com");
  await page.locator("text=Install Repository >> div").click();

  // For now, disabling the this test case temporarily.

  // Wait for new packages to be indexed
  await page.waitForTimeout(5000);

  // Check if our package shows up in catalog
  await page.click(`a:has-text("${repoName}")`);
  await page.click('a:has-text("simplechart")');

  // Deploy package
  await page.click('cds-button:has-text("Deploy")');
  await page.selectOption('select[name="package-versions"]', "0.1.0");
  const releaseNameLocator = page.locator("#releaseName");
  await releaseNameLocator.waitFor();
  await expect(releaseNameLocator).toHaveText("");
  const appName = utils.getRandomName("test-10-release");
  console.log(`Creating release "${appName}"`);
  await releaseNameLocator.fill(appName);

  // Select version and deploy
  await page.locator('cds-button:has-text("Deploy")').click();

  // Assertions
  // Wait for the app to be deployed and select it from "Applications"
  await page.waitForTimeout(5000);
  await page.click('a.nav-link:has-text("Applications")');
  await page.waitForTimeout(3000); // Sometimes typing was too fast to get the result shown
  await page.locator("input#search").fill(appName);
  await page.waitForTimeout(3000);
  await page.click(`a .card-title:has-text("${appName}")`);

  await page.waitForSelector("css=.application-status-pie-chart-number >> text=1", {
    timeout: deployTimeout,
  });
  await page.waitForSelector("css=.application-status-pie-chart-title >> text=Ready", {
    timeout: deployTimeout,
  });

  // Clean up
  await page.locator('cds-button:has-text("Delete")').click();
  await page.locator('cds-modal-actions button:has-text("Delete")').click();
});
