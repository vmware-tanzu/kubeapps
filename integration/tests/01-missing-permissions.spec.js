const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("./utils/kubeapps-login");
const utils = require("./utils/util-functions");

test("Regular user fails to deploy an application due to missing permissions", async ({ page }) => {
  test.setTimeout(60000);
  
  // Log in
  const k = new KubeappsLogin(page);
  await k.doLogin("kubeapps-user@example.com", "password", process.env.VIEW_TOKEN);

  // Select package to deploy
  await page.click('a.nav-link:has-text("Catalog")');
  await page.click('a .card-title:has-text("apache")');
  await page.click('cds-button:has-text("Deploy") >> nth=0');

  // Deploy package
  const releaseNameLocator = page.locator("#releaseName");
  await releaseNameLocator.waitFor();
  await expect(releaseNameLocator).toHaveText("");
  await releaseNameLocator.type(utils.getRandomName("test-01-release"));
  await page.locator('cds-button:has-text("Deploy")').click();

  const errorLocator = page.locator(".alert-items .alert-text");
  await expect(errorLocator).toHaveCount(1);
  await page.waitForTimeout(5000);

  // For some reason, UI is showing different error messages randomly
  // Custom assertion logic
  const errorMsg = await errorLocator.textContent();
  console.log(`Error message on UI = "${errorMsg}"`);

  await page.waitForFunction(msg => {
    return msg.indexOf("secrets is forbidden") > -1 || msg.indexOf("unable to read secret") > -1;
  }, errorMsg);
});
