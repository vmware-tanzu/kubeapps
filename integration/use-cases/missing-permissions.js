const { setDefaultOptions } = require("expect-puppeteer");

setDefaultOptions({ timeout: 8000 });
jest.setTimeout(120000);

test("Fails to deploy an application due to missing permissions", async () => {
  page.setDefaultTimeout(8000);
  await page.goto(getUrl("/#/login"));

  await expect(page).toFillForm("form", {
    token: process.env.VIEW_TOKEN
  });

  await expect(page).toClick("button", { text: "Login" });

  await expect(page).toClick("a", { text: "Catalog" });

  await expect(page).toClick("a", { text: "aerospike", timeout: 60000 });

  await expect(page).toClick("button", { text: "Deploy" });

  await expect(page).toClick("button", { text: "Submit" });

  await expect(page).toMatch(
    "You don't have sufficient permissions to create",
    { timeout: 20000 }
  );
});
