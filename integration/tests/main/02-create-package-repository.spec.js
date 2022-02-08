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
  await page.click('a.dropdown-menu-link:has-text("App Repositories")');
  await page.waitForTimeout(3000);

  // Add new repo
  console.log('Creating repository "my-repo"');
  await page.click('cds-button:has-text("Add App Repository")');
  await page.fill("input#kubeapps-repo-name", "my-repo");
  await page.fill("input#kubeapps-repo-url", "https://charts.gitlab.io/");
  await page.click('cds-button:has-text("Install Repo")');

  // Wait for new packages to be indexed
  await page.waitForTimeout(5000);

  // Check if packages show up in catalog
  await page.click('a:has-text("my-repo")');
  await page.waitForSelector('css=.catalog-container .card-title >> text="gitlab-runner"');

  // Clean up
  // Go back to repos page and delete repo
  await page.click(".dropdown.kubeapps-menu button.kubeapps-nav-link");
  await page.click('a.dropdown-menu-link:has-text("App Repositories")');
  await page.waitForTimeout(3000);
  await page.click("cds-button#delete-repo-my-repo");
  await page.locator('cds-modal-actions button:has-text("Delete")').click();
});
