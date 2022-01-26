const { test, expect } = require("@playwright/test");
const { KubeappsOidcLogin } = require("./utils/kubeapps-login");
const { TestUtils } = require("./utils/util-functions");

test("Deploys package with default values in the second cluster", async ({ page }) => {
  test.setTimeout(120000);

  // Log in
  const login = new KubeappsOidcLogin(page);
  await login.doOidcLogin("kubeapps-operator@example.com", "password");

  // Change cluster using ui
  await page.click(".kubeapps-dropdown .kubeapps-nav-link");
  await page.selectOption('select[name="clusters"]', "second-cluster");
  await page.click('cds-button:has-text("Change Context")');

  // Select package to deploy
  await page.click('a.nav-link:has-text("Catalog")');
  await page.click('a:has-text("foo apache chart for CI")');
  await page.click('cds-button:has-text("Deploy") >> nth=0');

  // Deploy package
  const releaseNameLocator = page.locator("#releaseName");
  await releaseNameLocator.waitFor();
  await expect(releaseNameLocator).toHaveText("");
  await releaseNameLocator.type(TestUtils.getRandomName("my-app-01-deploy"));
  await page.locator('cds-button:has-text("Deploy")').click();

  // Assertions
  await page.screenshot({ path: "reports/screenshots/01-multicluster-deploy-pre-assertion.png" });
  await page.waitForSelector("css=.application-status-pie-chart-number >> text=1");
  await page.waitForSelector("css=.application-status-pie-chart-title >> text=Ready");

  // Clean up
  await page.locator('cds-button:has-text("Delete")').click();
  await page.locator('cds-modal-actions button:has-text("Delete")').click();
});
