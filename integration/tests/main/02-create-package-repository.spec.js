const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("../utils/kubeapps-login");
const utils = require("../utils/util-functions");

test("Create a new package repository successfully", async ({ page }) => {
  test.setTimeout(60000);

  // Log in
  const k = new KubeappsLogin(page);
  await k.doLogin("kubeapps-operator@example.com", "password", process.env.ADMIN_TOKEN);

  // Go to repos page
  await page.goto(utils.getUrl("/#/c/default/ns/kubeapps/config/repos"));

  // Add new repo
  console.log('Creating repository "my-repo"');
  await page.click('cds-button:has-text("Add App Repository")');
  await page.type("input#kubeapps-repo-name", "my-repo");
  await page.type("input#kubeapps-repo-url", "https://charts.gitlab.io/");
  await page.click('cds-button:has-text("Install Repo")');

  // Wait for new packages to be indexed
  await page.waitForTimeout(5000);

  // Check if packages show up in catalog
  await page.click('a:has-text("my-repo")');
  await page.waitForSelector('css=.catalog-container .card-title >> text="gitlab-runner"');

  // Clean up
  const axInstance = await utils.getAxiosInstance(page);
  const response = await axInstance.delete(
    "/api/v1/clusters/default/namespaces/kubeapps/apprepositories/my-repo",
  );
  expect(response.status).toEqual(200);
});
