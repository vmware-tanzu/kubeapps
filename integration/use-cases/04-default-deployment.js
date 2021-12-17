const utils = require("./lib/utils");

test("Deploys an application with the values by default", async () => {
  await utils.login(
    page,
    process.env.USE_MULTICLUSTER_OIDC_ENV,
    "/",
    process.env.ADMIN_TOKEN,
    "kubeapps-operator@example.com",
    "password",
  );

  await expect(page).toClick("a", { text: "Catalog" });

  await expect(page).toMatchElement("a", { text: "foo apache chart for CI", timeout: 60000 });
  await expect(page).toClick("a", { text: "foo apache chart for CI" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatchElement("#releaseName", { text: "" });
  await page.type("#releaseName", utils.getRandomName("my-app-for-04-deploy"));

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Ready", { timeout: 60000 });
});
