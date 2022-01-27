const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("./utils/kubeapps-login");
const utils = require("./utils/util-functions");

test("Upgrades an application", async ({ page }) => {
  test.setTimeout(360000);

  // Log in
  const k = new KubeappsLogin(page);
  await k.doLogin("kubeapps-operator@example.com", "password", process.env.EDIT_TOKEN);

  // Go to catalog
  await page.goto(utils.getUrl("/#/c/default/ns/default/catalog?Search=apache&Repository=bitnami"));

  // Select package
  //await page.waitForSelector('.card-title:has-text("kubeapps-apache")');
  await page.click('.card-title:has-text("kubeapps-apache")');

  // Select an older version to be installed
  const defaultVersion = await page.inputValue('select[name="package-versions"]');
  await page.selectOption('select[name="package-versions"]', "8.6.2");
  const olderVersion = await page.inputValue('select[name="package-versions"]');
  expect(defaultVersion).not.toBe(olderVersion);

  // Deploy package
  await page.click('cds-button:has-text("Deploy") >> nth=0');

  // Set replicas
  await page.locator("input[type='number']").fill("2");
  await page.click('li:has-text("Changes")');
  await expect(page.locator("section#deployment-form-body-tabs-panel2")).toContainText(
    "replicaCount: 2",
  );

  // Set release name
  const releaseNameLocator = page.locator("#releaseName");
  await releaseNameLocator.waitFor();
  await expect(releaseNameLocator).toHaveText("");
  await releaseNameLocator.type(utils.getRandomName("test-07-upgrade"));

  // Trigger deploy
  await page.locator('cds-button:has-text("Deploy")').click();
  await utils.takeScreenShot(page, "07-deployed");

  await page.click('cds-button:has-text("Upgrade")');

  await expect(page.locator(".left-menu")).toContainText(olderVersion);
  await page.selectOption('select[name="package-versions"]', defaultVersion);
  await page.waitForTimeout(2000);
  const newSelection = await page.inputValue('select[name="package-versions"]');
  expect(newSelection).toBe(defaultVersion);

  await utils.takeScreenShot(page, "07-test-pre-upgrade");
  await page.locator('cds-button:has-text("Deploy")').click();

  // Check upgrade result
  await expect(page.locator(".left-menu")).toContainText("Up to date");
  await page.waitForSelector("css=.application-status-pie-chart-number >> text=2");
  await page.waitForSelector("css=.application-status-pie-chart-title >> text=Ready");

  // Clean up
  await page.locator('cds-button:has-text("Delete")').click();
  await page.locator('cds-modal-actions button:has-text("Delete")').click();
});
