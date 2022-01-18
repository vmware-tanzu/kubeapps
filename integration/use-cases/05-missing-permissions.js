const utils = require("./lib/utils");
const testName = "05-missing-permissions";

test("Fails to deploy an application due to missing permissions", async () => {
  console.log("05 -> Before log in");
  await utils.login(
    page,
    process.env.USE_MULTICLUSTER_OIDC_ENV,
    "/",
    process.env.VIEW_TOKEN,
    "kubeapps-user@example.com",
    "password",
  );
  console.log("05 -> Logged in");

  await utils.doAction("Click on Catalog", expect(page).toClick("a", { text: "Catalog" }));
  console.log("05 -> Catalog link clicked");

  await expect(page).toMatchElement("a", { text: "apache", timeout: 60000 });
  console.log("05 -> Apache card exists");

  await utils.doAction("Click on Apache card", expect(page).toClick("a", { text: "apache" }));
  console.log("05 -> Apache card clicked");

  await utils.doAction("Select Deploy the selection button", expect(page).toClick("cds-button", { text: "Deploy" }));
  console.log("05 -> Select and deploy button clicked");

  await expect(page).toMatchElement("#releaseName", { text: "" });
  const releaseName = utils.getRandomName("my-app-for-05-perms");
  await page.type("#releaseName", releaseName);
  console.log("05 -> Input release name = " + releaseName);

  page.waitForTimeout(3000);
  await expect(page).toClick("cds-button", { text: "Deploy" });
  console.log("05 -> Clicked deploy");

  await expect(page).toMatch("secrets is forbidden", { timeout: 60000 });
  console.log("05 ->All checks done");
});
