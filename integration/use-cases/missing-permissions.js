test("Fails to deploy an application due to missing permissions", async () => {
  // ODIC login
  await page.goto(getUrl("/"));
  await page.waitForNavigation();
  await expect(page).toClick("cds-button", { text: "Login via OIDC Provider" });
  await page.waitForNavigation();
  await expect(page).toClick(".dex-container button", { text: "Log in with Email" });
  await page.waitForNavigation();
  await page.type("input[id=\"login\"]", "kubeapps-user@example.com");
  await page.type("input[id=\"password\"]", "password");
  await page.waitForSelector("#submit-login", { visible: true, timeout: 10000 });
  await page.evaluate((selector) => document.querySelector(selector).click(), "#submit-login");
  await page.waitForSelector(".kubeapps-header-content", { visible: true, timeout: 10000 });

  await expect(page).toClick("a", { text: "Catalog" });
  // wait until load
  await page.evaluate(() => {
    [...document.querySelectorAll('.kubeapps-dropdown-header')].find(element => element.outerText.includes("Current Context"));
  });

  await expect(page).toClick("a", { text: "apache", timeout: 60000 });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  // wait for the loading msg to disappear
  await page.waitForFunction(() => !document.querySelector("#root > section > main > div > div > section > h3"));

  await expect(page).toMatch("missing permissions", { timeout: 20000 });
});
