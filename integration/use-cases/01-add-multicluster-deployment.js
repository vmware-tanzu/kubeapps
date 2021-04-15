const utils = require("./lib/utils");

test("Deploys an application with the values by default", async () => {
  await utils.login(
    page,
    process.env.USE_MULTICLUSTER_OIDC_ENV,
    "/",
    "",
    "kubeapps-operator@example.com",
    "password",
  );

  // Change cluster using ui
  await expect(page).toClick(".kubeapps-nav-link");

  await page.select('select[name="clusters"]', "second-cluster");

  await expect(page).toClick("cds-button", { text: "Change Context" });

  await expect(page).toClick("a", { text: "Catalog" });

  await expect(page).toMatchElement("a", { text: "apache", timeout: 60000 });
  await expect(page).toClick("a", { text: "apache" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatchElement("#releaseName", { text: "" });
  await page.type("#releaseName", utils.getRandomName("my-app"));

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Ready", { timeout: 60000 });
});
