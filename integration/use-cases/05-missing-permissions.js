const utils = require("./lib/utils");
const testName = "05-missing-permissions";

test("Fails to deploy an application due to missing permissions", async () => {
  /* 
  await utils.login(
    page,
    process.env.USE_MULTICLUSTER_OIDC_ENV,
    "/",
    process.env.VIEW_TOKEN,
    "kubeapps-user@example.com",
    "password",
  );

  await utils.doAction("Click on Catalog", expect(page).toClick("a", { text: "Catalog" }));

  await expect(page).toMatchElement("a", { text: "apache", timeout: 60000 });

  await utils.doAction("Click on Apache card", expect(page).toClick("a", { text: "apache" }));

  await utils.doAction("Select Deploy the selection button", expect(page).toClick("cds-button", { text: "Deploy" }));

  await expect(page).toMatchElement("#releaseName", { text: "" });
  const releaseName = utils.getRandomName("my-app-for-05-perms");
  await page.type("#releaseName", releaseName);

  page.waitForTimeout(3000);
  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Missing permissions", { timeout: 60000 });
  */
});
