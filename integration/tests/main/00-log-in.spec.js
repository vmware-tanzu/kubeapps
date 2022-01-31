const { test, expect } = require("@playwright/test");
const { KubeappsLogin } = require("../utils/kubeapps-login");
const utils = require("../utils/util-functions");

test.describe("Log in", () => {
  test("Logs in successfully via OIDC", async ({ page }) => {
    const k = new KubeappsLogin(page);
    await k.doLogin("kubeapps-user@example.com", "password", process.env.VIEW_TOKEN);

    //await page.waitForResponse(async response => (await response.status()) === 200);
    await page.waitForSelector('css=h1 >> text="Applications"');
  });
});
