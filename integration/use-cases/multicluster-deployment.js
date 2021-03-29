test("Deploys an application with the values by default", async () => {
  // ODIC login
  await page.goto(getUrl("/"));
  await page.waitForNavigation();
  await expect(page).toClick("cds-button", { text: "Login via OIDC Provider" });
  await page.waitForNavigation();
  await expect(page).toClick(".dex-container button", {
    text: "Log in with Email",
  });
  await page.waitForNavigation();
  await page.type('input[id="login"]', "kubeapps-operator@example.com");
  await page.type('input[id="password"]', "password");
  await page.waitForSelector("#submit-login", {
    visible: true,
    timeout: 10000,
  });
  await page.evaluate(
    (selector) => document.querySelector(selector).click(),
    "#submit-login"
  );
  await page.waitForSelector(".kubeapps-header-content", {
    visible: true,
    timeout: 10000,
  });

  // Change cluster using ui
  await expect(page).toClick(".kubeapps-nav-link");

  await page.select('select[name="clusters"]', "second-cluster");

  await expect(page).toClick("cds-button", { text: "Change Context" });

  await expect(page).toClick("a", { text: "Catalog" });

  await expect(page).toClick("a", { text: "apache", timeout: 60000 });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  // wait for the loading msg to disappear
  await page.waitForFunction(
    () =>
      !document.querySelector(
        "#root > section > main > div > div > section > h3"
      ),
    { timeout: 60000 }
  );

  await expect(page).toMatch("Ready", { timeout: 60000 });
});
