const { test, expect } = require("@playwright/test");
const { KubeappsOidcLogin } = require("./utils/kubeapps-login");

test.describe("Log in", () => {
  test("Logs in successfully via OIDC", async ({ page }) => {
    const login = new KubeappsOidcLogin(page);
    await login.doOidcLogin("kubeapps-user@example.com", "password");

    //await page.waitForResponse(async response => (await response.status()) === 200);
    await page.waitForSelector('css=h2 >> text="Welcome To Kubeapps"');

    await page.screenshot({ path: "screenshots/login-submit-grant-post.png" });
  });
});
