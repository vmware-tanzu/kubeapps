const utils = require("./lib/utils");

test("Fails to deploy an application due to missing permissions", async () => {
  await utils.login(
    page,
    document,
    process.env.USE_MULTICLUSTER_OIDC_ENV,
    "/",
    process.env.VIEW_TOKEN,
    "kubeapps-user@example.com",
    "password",
  );

  await expect(page).toClick("a", { text: "Catalog" });
  // wait until load
  await page.evaluate(() => {
    [...document.querySelectorAll(".kubeapps-dropdown-header")].find(element =>
      element.outerText.includes("Current Context"),
    );
  });

  await expect(page).toClick("a", { text: "apache", timeout: 60000 });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  // wait for the loading msg to disappear
  await page.waitForFunction(
    () => !document.querySelector("#root > section > main > div > div > section > h3"),
  );

  await expect(page).toMatch("missing permissions", { timeout: 20000 });
});
