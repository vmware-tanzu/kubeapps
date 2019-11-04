jest.setTimeout(120000);

test("Deploys an application with the values by default", async () => {
  page.setDefaultTimeout(8000);
  await page.goto(getUrl("/#/login"));

  await expect(page).toFillForm("form", {
    token: process.env.ADMIN_TOKEN
  });

  await expect(page).toClick("button", { text: "Login" });

  await expect(page).toClick("a", { text: "Catalog" });

  await expect(page).toClick("a", { text: "aerospike", timeout: 60000 });

  await expect(page).toClick("button", { text: "Deploy" });

  await expect(page).toClick("button", { text: "Submit" });

  await expect(page).toMatch("Ready", { timeout: 60000 });
});
