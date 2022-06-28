// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("../utils/kubeapps-login");
const utils = require("../utils/util-functions");

test.describe("Limited user simple deployments", () => {
  test("Regular user can deploy and delete packages in its own namespace from global repo without secrets", async ({
    page,
  }) => {
    // Log in as admin to create a repo without password
    const k = new KubeappsLogin(page);
    await k.doLogin("kubeapps-operator@example.com", "password", process.env.ADMIN_TOKEN);

    // Change namespace using UI
    await page.click(".kubeapps-dropdown .kubeapps-nav-link");
    await page.selectOption('select[name="namespaces"]', "kubeapps-repos-global");
    await page.click('cds-button:has-text("Change Context")');

    // Go to repos page
    await page.click(".dropdown.kubeapps-menu button.kubeapps-nav-link");
    await page.locator("text=Package Repositories").click();
    await expect(page).not.toContain("text=Fetching Package Repositories...");

    // Add new repo
    const repoName = utils.getRandomName("repo-09");
    console.log(`Creating package repository "${repoName}"`);
    await page.locator("text=Add Package Repository >> div").click();
    await page.fill("input#kubeapps-repo-name", repoName);
    await page.fill(
      "input#kubeapps-repo-url",
      "https://prometheus-community.github.io/helm-charts",
    );
    await page.locator("text=Helm Charts").first().click();
    await page.locator("text=Helm Repository").click();
    await page.locator("text=Install Repository >> div").click();

    await page.waitForLoadState("networkidle");
    await page.click(`div.page-content table td a:has-text("${repoName}")`);

    // Log out admin and log in regular user
    await k.doLogout();
    await k.doLogin("kubeapps-user@example.com", "password", process.env.VIEW_TOKEN);

    // Switch to user's namespace using UI
    await page.click(".kubeapps-dropdown .kubeapps-nav-link");
    await page.selectOption('select[name="namespaces"]', "kubeapps-user-namespace");
    await page.click('cds-button:has-text("Change Context")');

    // Select package to deploy
    await page.click('a.nav-link:has-text("Catalog")');
    await page.locator("input#search").fill("alertmanager");
    await page.waitForTimeout(3000);
    await page.click('a:has-text("alertmanager")');
    await page.click('cds-button:has-text("Deploy") >> nth=0');

    // Deploy package
    const releaseNameLocator = page.locator("#releaseName");
    await releaseNameLocator.waitFor();
    await expect(releaseNameLocator).toHaveText("");
    const releaseName = utils.getRandomName("test-09-release");
    console.log(`Creating release "${releaseName}"`);
    await releaseNameLocator.fill(releaseName);
    await page.locator('cds-button:has-text("Deploy")').click();

    // Check that package is deployed
    await page.waitForSelector("css=.application-status-pie-chart-number >> text=1", {
      timeout: utils.getDeploymentTimeout(),
    });
    await page.waitForSelector("css=.application-status-pie-chart-title >> text=Ready", {
      timeout: utils.getDeploymentTimeout(),
    });

    // Delete deployment
    await page.locator('cds-button:has-text("Delete")').click();
    await page.locator('cds-modal-actions button:has-text("Delete")').click();
    await page.waitForTimeout(10000);

    // Search for package deployed
    await page.click('a.nav-link:has-text("Applications")');
    await page.waitForTimeout(3000); // Sometimes typing was too fast to get the result shown
    await page.locator("input#search").fill("alertmanager");
    await page.waitForTimeout(3000);
    const packageLocator = page.locator('a:has-text("alertmanager")');
    await expect(packageLocator).toHaveCount(0);
  });
});
