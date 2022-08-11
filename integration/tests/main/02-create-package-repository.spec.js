// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("../utils/kubeapps-login");
const utils = require("../utils/util-functions");

test("Create a new package repository successfully", async ({ page }) => {
  // Log in
  const k = new KubeappsLogin(page);
  await k.doLogin("kubeapps-operator@example.com", "password", process.env.ADMIN_TOKEN);

  // Go to repos page
  await page.click(".dropdown.kubeapps-menu button.kubeapps-nav-link");
  await page.locator("text=Package Repositories").click();
  await expect(page).not.toContain("text=Fetching Package Repositories...");

  // Add new repo
  console.log('Creating package repository "my-repo"');
  await page.locator("text=Add Package Repository >> div").click();
  await page.fill("input#kubeapps-repo-name", "my-repo");
  await page.fill("input#kubeapps-repo-url", "https://charts.gitlab.io/");
  await page.locator("text=Helm Charts").first().click();
  await page.locator("text=Helm Repository").click();
  await page.locator("text=Install Repository >> div").click();

  // Wait for new packages to be indexed
  await page.waitForTimeout(5000);

  // Check if packages show up in catalog
  await page.locator('a:has-text("my-repo")').click();
  await page.waitForSelector('css=.catalog-container .card-title >> text="gitlab-runner"');

  // Clean up
  // Go back to repos page and delete repo
  await page.click(".dropdown.kubeapps-menu button.kubeapps-nav-link");
  await page.locator("text=Package Repositories").click();
  await expect(page).not.toContain("text=Fetching Package Repositories...");
  await page.locator("#delete-repo-my-repo div").click();
  await page.locator('button:has-text("Delete")').click();
});
