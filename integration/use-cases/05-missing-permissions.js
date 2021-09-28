const utils = require("./lib/utils");
const testName = "05-missing-permissions";

test("Fails to deploy an application due to missing permissions", async () => {
  await utils.login(
    page,
    process.env.USE_MULTICLUSTER_OIDC_ENV,
    "/",
    process.env.VIEW_TOKEN,
    "kubeapps-user@example.com",
    "password",
  );

  await expect(page).toClick("a", { text: "Catalog" });
  await utils.retryAndRefresh(
    page,
    3,
    async () => {
      await expect(page).toMatchElement("a", { text: "apache", timeout: 60000 });
    },
    testName,
  );

  await expect(page).toClick("a", { text: "apache" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatchElement("#releaseName", { text: "" });
  await page.type("#releaseName", utils.getRandomName("my-app-for-05-perms"));

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await utils.retryAndRefresh(
    page,
    3,
    async () => {
      await expect(page).toMatch("Missing permissions", { timeout: 60000 });
    },
    testName,
  );
});
