const utils = require("./lib/utils");

test("Creates a registry", async () => {
  // ODIC login
  var token;
  page.on('response', response => {
    if (response.status() >= 400) {
      console.log("ERROR: ", response.status() + " " + response.url());
    }
    token = response.headers()["authorization"] || token;
  });
  await page.goto(getUrl("/#/c/default/ns/default/config/repos"));
  await page.waitForNavigation();
  await expect(page).toClick("cds-button", { text: "Login via OIDC Provider" });
  await page.waitForNavigation();
  await expect(page).toClick(".dex-container button", { text: "Log in with Email" });
  await page.waitForNavigation();
  await page.type("input[id=\"login\"]", "kubeapps-operator@example.com");
  await page.type("input[id=\"password\"]", "password");
  await page.waitForSelector("#submit-login", { visible: true, timeout: 10000 });
  await page.evaluate((selector) => document.querySelector(selector).click(), "#submit-login");
  await page.waitForSelector(".kubeapps-header-content", { visible: true, timeout: 10000 });
  console.log("Token after OIDC authentication: " + token);
  await page.goto(getUrl("/#/c/default/ns/kubeapps/config/repos"));

  await expect(page).toClick("cds-button", { text: "Add App Repository" });

  await page.type("input[placeholder=\"example\"]", "my-repo");

  await page.type("input[placeholder=\"https://charts.example.com/stable\"]", "https://charts.gitlab.io/");

  // Similar to the above click for an App Repository, the click on
  // the Install Repo doesn't always register (in fact, from the
  // screenshot on failure, it appears to focus the button only (hover css applied)
  await expect(page).toClick("cds-button", { text: "Install Repo" });
  await utils.retryAndRefresh(page, 3, async () => {
    // TODO(andresmgot): In theory, there is no need to refresh but sometimes the repo
    // does not appear
    await expect(page).toClick("a", { text: "my-repo" });
  });

  await utils.retryAndRefresh(page, 3, async () => {
    await expect(page).toMatch("gitlab-runner", { timeout: 10000 });
  });
});
