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
  await page.click('a.dropdown-menu-link:has-text("Package Repositories")');
  await page.waitForTimeout(3000);

  // Add new repo
  await page.click('cds-button:has-text("Add Package Repository")');
  const repoName = utils.getRandomName("private-img-repo");
  console.log(`Creating package repository "${repoName}"`);
  await page.fill("input#kubeapps-repo-name", repoName);
  await page.fill("input#kubeapps-repo-url", "http://chartmuseum-chartmuseum.kubeapps:8080");

  // Set credentials for repository
  await page.click('label:has-text("Basic Auth")');
  await page.fill("input#kubeapps-repo-username", "admin");
  await page.fill("input#kubeapps-repo-password", "password");

  // Create a new secret for Docker repo credentials
  const secretName = utils.getRandomName("my-repo-secret");
  await page.click('.docker-creds-subform-button button:has-text("Add new credentials")');
  await page.fill("input#kubeapps-docker-cred-secret-name", secretName);
  await page.fill("input#kubeapps-docker-cred-server", process.env.DOCKER_REGISTRY_URL);
  await page.fill("input#kubeapps-docker-cred-username", process.env.DOCKER_USERNAME);
  await page.fill("input#kubeapps-docker-cred-password", process.env.DOCKER_PASSWORD);
  await page.click('.docker-creds-subform button:has-text("Submit")');

  // Select the newly created secret
  await page.selectOption("form cds-form-group cds-select select", secretName);

  await page.click('cds-button:has-text("Install Repo")');

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
