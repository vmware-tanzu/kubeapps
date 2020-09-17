test("Fails to deploy an application due to missing permissions", async () => {
  await page.goto(getUrl("/#/login"));

  await expect(page).toFillForm("form", {
    token: process.env.VIEW_TOKEN,
  });

  await page.evaluate(() =>
    document.querySelector("#login-submit-button").click()
  );

  await expect(page).toClick("a", { text: "Catalog" });

  await expect(page).toClick("a", { text: "apache", timeout: 60000 });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("missing permissions", { timeout: 20000 });
});
