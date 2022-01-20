const { test, expect } = require("@playwright/test");
const { KubeappsOidcLogin } = require("./utils/kubeapps-login");
const { TestUtils } = require("./utils/util-functions");

test("Create a new package repository successfully", async ({ page }) => {
  test.setTimeout(60000);

  // Log in
  const login = new KubeappsOidcLogin(page);
  await login.doOidcLogin("kubeapps-operator@example.com", "password");

  // Go to repos page
  await page.goto(login.getUrl("/#/c/default/ns/kubeapps/config/repos"));

  // Add new repo
  await page.click('cds-button:has-text("Add App Repository")');
  await page.type("input#kubeapps-repo-name", "my-repo");
  await page.type("input#kubeapps-repo-url", "https://charts.gitlab.io/");
  await page.click('cds-button:has-text("Install Repo")');

  // Wait for new packages to be indexed
  await page.waitForTimeout(5000);

  // Check if packages show up in catalog
  await page.reload({ waitUntil: "networkidle" });
  await page.click('a:has-text("my-repo")');
  await page.waitForSelector('css=.catalog-container .card-title >> text="gitlab-runner"');
});
