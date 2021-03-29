import { login, retryAndRefresh } from "./lib/utils";

test("Deploys an application with the values by default", async () => {
  await login(
    page,
    document,
    process.env.USE_MULTICLUSTER_OIDC_ENV,
    "/",
    process.env.ADMIN_TOKEN,
    "kubeapps-operator@example.com",
    "password"
  );

  // Change cluster using ui
  await expect(page).toClick(".kubeapps-nav-link");

  await page.select('select[name="clusters"]', "second-cluster");

  await expect(page).toClick("cds-button", { text: "Change Context" });

  await expect(page).toClick("a", { text: "Catalog" });

  await expect(page).toClick("a", { text: "apache", timeout: 60000 });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  // wait for the loading msg to disappear
  await page.waitForFunction(
    () =>
      !document.querySelector(
        "#root > section > main > div > div > section > h3"
      ),
    { timeout: 60000 }
  );

  await expect(page).toMatch("Ready", { timeout: 60000 });
});
