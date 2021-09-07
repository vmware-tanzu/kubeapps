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

  await expect(page).toMatchElement("a", { text: "Apache HTTP Server", timeout: 60000 });
  await expect(page).toClick("a", { text: "Apache HTTP Server" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatchElement("#releaseName", { text: "" });
  await page.type("#releaseName", utils.getRandomName("my-app"));

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Ready", { timeout: 60000 });
});
