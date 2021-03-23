test("Deploys an application with the values by default", async () => {
  // ODIC login
  await Promise.all([
    await page.goto(getUrl("/#/login")),
    await page.waitForNavigation(),
    await expect(page).toClick("cds-button", { text: "Login via OIDC Provider" }),
    await page.waitForNavigation(),
    await expect(page).toClick(".dex-container button", { text: "Log in with Email" }),
    await page.waitForNavigation(),
    await page.type("input[id=\"login\"]", "kubeapps-operator@example.com"),
    await page.type("input[id=\"password\"]", "password"),
    await expect(page).toClick("#submit-login", { text: "Login" }),
    await page.waitForNavigation({ waitUntil: 'networkidle2' }),
    await page.goto(getUrl("/#/c/default/ns/default/config/repos")),
    await page.waitForNavigation(),
    await page.goto(getUrl("/#/login")),
  ]);

  await expect(page).toClick("a", { text: "Catalog" });

  await expect(page).toClick("a", { text: "apache", timeout: 60000 });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Ready", { timeout: 60000 });
});
