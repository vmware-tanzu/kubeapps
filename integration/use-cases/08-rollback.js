const utils = require("./lib/utils");
const testName = "08-rollback";

test("Rolls back an application", async () => {
  await utils.login(
    page,
    process.env.USE_MULTICLUSTER_OIDC_ENV,
    "/#/c/default/ns/default/catalog?Repository=bitnami",
    process.env.EDIT_TOKEN,
    "kubeapps-operator@example.com",
    "password",
  );

  // Deploy the app
  await expect(page).toMatchElement("a", { text: "foo apache chart for CI", timeout: 60000 });
  await expect(page).toClick("a", { text: "foo apache chart for CI" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatchElement("#releaseName", { text: "" });
  await page.type("#releaseName", utils.getRandomName("my-app"));

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Ready", { timeout: 60000 });

  // Try to rollback when the app hasn't been upgraded
  await expect(page).toClick("cds-button", { text: "Rollback" });
  await expect(page).toMatchElement("cds-modal-content p", {
    text: "The application has not been upgraded, it's not possible to rollback.",
  });
  await expect(page).toClick("cds-button", { text: "Cancel" });

  // Upgrade the app to get another revision
  await expect(page).toClick("cds-button", { text: "Upgrade" });

  // Increase the number of replicas
  await utils.retryAndRefresh(
    page,
    3,
    async () => {
      await expect(page).toMatchElement("input[type='number']");
      // Increase the number of replicas
      await page.focus("input[type='number']");
      await page.keyboard.press("Backspace");
      await page.keyboard.type("2");

      await new Promise(r => setTimeout(r, 500));

      // Check that the Changes tab reflects the change
      await expect(page).toClick("li", { text: "Changes" });
      await expect(page).toMatch("replicaCount: 2");
    },
    testName,
  );

  await expect(page).toClick("cds-button", { text: "Deploy" });

  // Rollback to the previous revision (default selected value)
  await page.waitForTimeout(2000);
  await expect(page).toMatchElement(".application-status-pie-chart h5", { text: "Ready" });
  await expect(page).toClick("cds-button", { text: "Rollback" });
  await expect(page).not.toMatch("Loading");
  await expect(page).toMatch("(current: 2)");
  await expect(page).toClick("cds-modal-actions cds-button", { text: "Rollback" });

  // Check revision and rollback to a revision (manual selected value)
  await page.waitForTimeout(2000);
  await expect(page).toClick("cds-button", { text: "Rollback" });
  await expect(page).not.toMatch("Loading");
  await expect(page).toMatch("(current: 3)");

  await expect(page).toSelect("cds-select > select", "1");
  await expect(page).toClick("cds-modal-actions cds-button", { text: "Rollback" });

  // Check revisions
  await page.waitForTimeout(2000);
  await expect(page).toMatchElement(".application-status-pie-chart h5", { text: "Ready" });
  await expect(page).toClick("cds-button", { text: "Rollback" });
  await expect(page).not.toMatch("Loading");
  await expect(page).toMatch("(current: 4)");
});
