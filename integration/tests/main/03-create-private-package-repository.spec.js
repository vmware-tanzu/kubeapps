// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("../utils/kubeapps-login");
const utils = require("../utils/util-functions");

test("Create a new private package repository successfully", async ({ page }) => {
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
  await page.click('a.dropdown-menu-link:has-text("App Repositories")');
  await page.waitForTimeout(3000);

  // Add new repo
  await page.click('cds-button:has-text("Add App Repository")');
  const repoName = utils.getRandomName("my-repo");
  console.log(`Creating repository "${repoName}"`);
  await page.fill("input#kubeapps-repo-name", repoName);
  await page.fill("input#kubeapps-repo-url", "http://chartmuseum-chartmuseum.kubeapps:8080");

  // Set credentials
  await page.click('label:has-text("Basic Auth")');
  await page.fill("input#kubeapps-repo-username", "admin");
  await page.fill("input#kubeapps-repo-password", "password");

  // Create a new secret for Docker repo credentials
  const secretName = utils.getRandomName("my-repo-secret");
  await page.click('.docker-creds-subform-button button:has-text("Add new credentials")');
  await page.fill("input#kubeapps-docker-cred-secret-name", secretName);
  await page.fill("input#kubeapps-docker-cred-server", "https://index.docker.io/v1/");
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
  await page.click('a:has-text("foo apache chart for CI")');

  // Deploy package
  await page.click('cds-button:has-text("Deploy")');
  await page.selectOption('select[name="package-versions"]', "8.6.2");
  const releaseNameLocator = page.locator("#releaseName");
  await releaseNameLocator.waitFor();
  await expect(releaseNameLocator).toHaveText("");
  const appName = utils.getRandomName("test-03-release");
  console.log(`Creating release "${appName}"`);
  await releaseNameLocator.fill(appName);

  // Select version and deploy
  await page.locator('select[name="package-versions"]').selectOption("8.6.2");
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

  // Now that the deployment has been created, we check that the imagePullSecret
  // has been added. For doing so, we query the resources API to get info of the
  // deployment
  const axInstance = await utils.getAxiosInstance(page, k.token);
  const resourceResp = await axInstance.get(
    `/apis/plugins/resources/v1alpha1/helm.packages/v1alpha1/c/default/ns/default/${appName}`,
  );
  expect(resourceResp.status).toEqual(200);

  let deployment;
  resourceResp.data
    .trim()
    .split(/\r?\n/)
    .forEach(r => {
      // Axios doesn't provide streaming responses, so splitting on new line works
      // but gives us a string, not JSON, and may leave a blank line at the end.
      const response = JSON.parse(r);
      const resourceRef = response.result?.resourceRef;
      if (resourceRef.kind === "Deployment" && resourceRef.name.match(appName)) {
        deployment = JSON.parse(response.result?.manifest);
      }
    });

  expect(deployment?.spec?.template?.spec?.imagePullSecrets).toEqual([{ name: secretName }]);

  // Prepare and verify the upgrade
  await page.waitForSelector('cds-button:has-text("Upgrade")');
  await page.click('cds-button:has-text("Upgrade")');

  // Check first current installed version
  await page.waitForSelector('select[name="package-versions"]');
  const packageVersionValue = await page.inputValue('select[name="package-versions"]');
  expect(packageVersionValue).toEqual("8.6.2");

  // Select new version
  await page.selectOption('select[name="package-versions"]', "8.6.3");

  // Ensure that the new value is selected
  await page.waitForSelector('select[name="package-versions"]');
  const newPackageVersionValue = await page.inputValue('select[name="package-versions"]');
  expect(newPackageVersionValue).toEqual("8.6.3");
  await page.click('li:has-text("Changes")');
  await expect(page.locator("section#deployment-form-body-tabs-panel2")).toContainText(
    "tag: 2.4.48-debian-10-r75",
  );

  // Deploy upgrade
  await page.click('cds-button:has-text("Deploy")');

  // Check upgrade result
  await page.waitForSelector(".left-menu");
  // Wait for the app to be deployed and select it from "Applications"
  await expect(page.locator(".left-menu")).toContainText("Up to date", { timeout: deployTimeout });
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
