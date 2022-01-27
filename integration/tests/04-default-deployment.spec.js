const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("./utils/kubeapps-login");
const utils = require("./utils/util-functions");

test("Deploys package with default values in main cluster", async ({ page }) => {
  test.setTimeout(120000);

  // Log in
  const k = new KubeappsLogin(page);
  await k.doLogin("kubeapps-operator@example.com", "password", process.env.ADMIN_TOKEN);

  // Select package to deploy
  await page.click('a.nav-link:has-text("Catalog")');
  await page.locator("input#search").type("apache");
  await page.click('a:has-text("foo apache chart for CI")');
  await page.click('cds-button:has-text("Deploy") >> nth=0');

  // Deploy package
  const releaseNameLocator = page.locator("#releaseName");
  await releaseNameLocator.waitFor();
  await expect(releaseNameLocator).toHaveText("");
  await releaseNameLocator.type(utils.getRandomName("test-04-release"));
  await page.locator('cds-button:has-text("Deploy")').click();

  // Assertions
  await page.waitForSelector("css=.application-status-pie-chart-number >> text=1");
  await page.waitForSelector("css=.application-status-pie-chart-title >> text=Ready");

  // Clean up
  await page.locator('cds-button:has-text("Delete")').click();
  await page.locator('cds-modal-actions button:has-text("Delete")').click();
});
