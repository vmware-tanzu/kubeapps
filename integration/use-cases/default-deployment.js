test("Deploys an application with the values by default", async () => {
  await page.goto(getUrl("/#/login"));

  await expect(page).toFillForm("form", {
    token: process.env.ADMIN_TOKEN,
  });

  // I am not sure why, but clicking on the Login button causes an error
  // "Node is either not visible or not an HTMLElement"
  // https://github.com/puppeteer/puppeteer/issues/2977
  await page.evaluate(() =>
    document.querySelector("#login-submit-button").click()
  );

  await expect(page).toClick("a", { text: "Catalog" });

  await expect(page).toClick("a", { text: "apache", timeout: 60000 });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Ready", { timeout: 60000 });
});
