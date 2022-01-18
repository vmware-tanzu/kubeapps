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
  console.log("05 -> Logged in");

  await expect(page).toClick("a", { text: "Catalog" });
  console.log("05 -> Catalog link clicked");
  await utils.retryAndRefresh(
    page,
    3,
    async () => {
      await expect(page).toMatchElement("a", { text: "apache", timeout: 60000 });
      console.log("05 -> Apache card existing");
    },
    testName,
  );

  await expect(page).toClick("a", { text: "apache" });
  console.log("05 -> Apache card clicked");

  await expect(page).toClick("cds-button", { text: "Deploy" });
  console.log("05 -> Deploy button clicked");

  await expect(page).toMatchElement("#releaseName", { text: "" });
  await page.type("#releaseName", utils.getRandomName("my-app-for-05-perms"));
  console.log("05 -> Input release name");

  await expect(page).toClick("cds-button", { text: "Deploy" });
  console.log("05 -> Clicked deploy");

  await utils.retryAndRefresh(
    page,
    3,
    async () => {
      console.log("05 -> Final check");
      await expect(page).toMatch("unable to read secret", { timeout: 60000 });
    },
    testName,
  );
});
