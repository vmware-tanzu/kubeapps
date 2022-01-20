const { test, expect } = require("@playwright/test");
const { KubeappsOidcLogin } = require("./utils/kubeapps-login");
const { TestUtils } = require("./utils/util-functions");

test("Create a new private package repository successfully", async ({ page }) => {
  test.setTimeout(60000);

  // Log in
  const login = new KubeappsOidcLogin(page);
  await login.doOidcLogin("kubeapps-operator@example.com", "password");

  // Go to repos page
  await page.goto(login.getUrl("/#/c/default/ns/default/config/repos"));

  // Add new repo
  await page.click('cds-button:has-text("Add App Repository")');
  const repoName = TestUtils.getRandomName("my-repo");
  await page.type("input#kubeapps-repo-name", repoName);
  await page.type("input#kubeapps-repo-url", "http://chartmuseum-chartmuseum.kubeapps:8080");

  // Set credentials
  await page.click('label:has-text("Basic Auth")');
  await page.type("input#kubeapps-repo-username", "admin");
  await page.type("input#kubeapps-repo-password", "password");

  // Create a new secret for Docker repo credentials
  const secretName = TestUtils.getRandomName("my-repo-secret");
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
  const appName = TestUtils.getRandomName("my-app");
  await releaseNameLocator.type(appName);

  // Select version and deploy
  await page.locator('select[name="package-versions"]').selectOption("8.6.2");
  await page.locator('cds-button:has-text("Deploy")').click();

  // Assertions
  await page.waitForSelector("css=.application-status-pie-chart-number >> text=1");
  await page.waitForSelector("css=.application-status-pie-chart-title >> text=Ready");

  /*
    TO-DO: 
      Translate rest of `02-create-private-registry.js` old test logic
      Add clean up
  */
});
