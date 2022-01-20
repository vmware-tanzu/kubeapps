const { test, expect } = require("@playwright/test");
const { KubeappsOidcLogin } = require("./utils/kubeapps-login");
const { TestUtils } = require("./utils/util-functions");

test("Regular user fails to deploy an application due to missing permissions", async ({ page }) => {
  // Log in
  const login = new KubeappsOidcLogin(page);
  await login.doOidcLogin("kubeapps-user@example.com", "password");

  // Select package to deploy
  await page.click('a.nav-link:has-text("Catalog")');
  await page.click('a:has-text("apache")');
  await page.click('cds-button:has-text("Deploy") >> nth=0');

  // Deploy package
  const releaseNameLocator = page.locator("#releaseName");
  await releaseNameLocator.waitFor();
  await expect(releaseNameLocator).toHaveText("");
  await releaseNameLocator.type(TestUtils.getRandomName("my-app-for-05-perms"));
  await page.locator('cds-button:has-text("Deploy")').click();

  const errorLocator = page.locator(".alert-items .alert-text");
  await expect(errorLocator).toHaveCount(1);
  await expect(errorLocator).toContainText("unable to read secret");
});
