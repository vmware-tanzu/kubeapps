test("Deploys an application with the values by default", async () => {
  // ODIC login
  var token;
  page.on('response', response => {
    if (response.status() >= 400) {
      console.log("ERROR: ", response.status() + " " + response.url());
    }
    token = response.headers()["authorization"] || token;
  });
  await page.goto(getUrl("/#/c/default/ns/default"));
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

  await expect(page).toClick("a", { text: "Catalog" });

  await expect(page).toClick("a", { text: "apache", timeout: 60000 });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Ready", { timeout: 60000 });
});
