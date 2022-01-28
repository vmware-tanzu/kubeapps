const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("./utils/kubeapps-login");
const utils = require("./utils/util-functions");

test("Create a new private package repository successfully", async ({ page }) => {
  test.setTimeout(120000);

  // Log in
  const k = new KubeappsLogin(page);
  await k.doLogin("kubeapps-operator@example.com", "password", process.env.ADMIN_TOKEN);

  // Go to repos page
  await page.goto(utils.getUrl("/#/c/default/ns/default/config/repos"));

  // Add new repo
  await page.click('cds-button:has-text("Add App Repository")');
  const repoName = utils.getRandomName("my-repo");
  console.log(`Creating repository "${repoName}"`);
  await page.type("input#kubeapps-repo-name", repoName);
  await page.type("input#kubeapps-repo-url", "http://chartmuseum-chartmuseum.kubeapps:8080");

  // Set credentials
  await page.click('label:has-text("Basic Auth")');
  await page.type("input#kubeapps-repo-username", "admin");
  await page.type("input#kubeapps-repo-password", "password");

  // Create a new secret for Docker repo credentials
  const secretName = utils.getRandomName("my-repo-secret");
  await page.click('.docker-creds-subform-button button:has-text("Add new credentials")');
  await page.type("input#kubeapps-docker-cred-secret-name", secretName);
  await page.type("input#kubeapps-docker-cred-server", "https://index.docker.io/v1/");
  await page.type("input#kubeapps-docker-cred-username", "user");
  await page.type("input#kubeapps-docker-cred-password", "password");
  await page.type("input#kubeapps-docker-cred-email", "user@example.com");
  await page.click('.docker-creds-subform button:has-text("Submit")');

  // Select the newly created secret
  await page.selectOption("form cds-form-group cds-select select", secretName);

  await page.click('cds-button:has-text("Install Repo")');

  // Wait for new packages to be indexed
  await page.waitForTimeout(5000);

  // Check if our package shows up in catalog
  await page.reload({ waitUntil: "networkidle" });
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
  await releaseNameLocator.type(appName);

  // Select version and deploy
  await page.locator('select[name="package-versions"]').selectOption("8.6.2");
  await page.locator('cds-button:has-text("Deploy")').click();

  // Assertions
  await page.waitForTimeout(5000);
  await page.waitForSelector("css=.application-status-pie-chart-number >> text=1");
  await page.waitForSelector("css=.application-status-pie-chart-title >> text=Ready");
  
  // Now that the deployment has been created, we check that the imagePullSecret
  // has been added. For doing so, we query the resources API to get info of the
  // deployment
  const axInstance = await utils.getAxiosInstance(page);
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
  await expect(page.locator(".left-menu")).toContainText("Up to date");
  await page.waitForSelector("css=.application-status-pie-chart-number >> text=1");
  await page.waitForSelector("css=.application-status-pie-chart-title >> text=Ready");

  // Clean up
  await page.locator('cds-button:has-text("Delete")').click();
  await page.locator('cds-modal-actions button:has-text("Delete")').click();

});
