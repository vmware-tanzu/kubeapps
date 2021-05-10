const utils = require("./lib/utils");

test("Upgrades an application", async () => {
  await utils.login(
    page,
    process.env.USE_MULTICLUSTER_OIDC_ENV,
    "/#/c/default/ns/default/catalog?Repository=bitnami",
    process.env.EDIT_TOKEN,
    "kubeapps-operator@example.com",
    "password",
  );

  // Deploy the app
  await expect(page).toMatchElement("a", { text: "apache", timeout: 60000 });
  await expect(page).toClick("a", { text: "apache" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatchElement("#releaseName", { text: "" });
  await page.type("#releaseName", utils.getRandomName("my-app"));

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Ready", { timeout: 60000 });

  // Try to rollback when the app hasn't been upgraded
  await expect(page).toClick("cds-button", { text: "Rollback" });
  await expect(page).toMatchElement("cds-modal-content p", { text: "The application has not been upgraded, it's not possible to rollback." });
  await expect(page).toClick("cds-button", { text: "Cancel" });

  // Upgrade the app to get another revision
  await expect(page).toClick("cds-button", { text: "Upgrade" });

  // Increase the number of replicas
  await expect(page).toMatchElement("input[type='number']");
  await page.focus("input[type='number']");
  await page.keyboard.press("Backspace");
  await page.keyboard.type("2");

  await new Promise(r => setTimeout(r, 500));

  await expect(page).toClick("li", { text: "Changes" });
  await expect(page).toMatch("replicaCount: 2");
  await expect(page).toMatchElement("input[type='number']", { value: 2 });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  // Rollback to the previous revision (default selected value)
  await expect(page).toClick("cds-button", { text: "Rollback" });

  await expect(page).toMatch("(current: 2)");
  await expect(page).toClick("cds-modal-actions cds-button", { text: "Rollback" });


  // Check revision and rollback to a revision (manual selected value)
  await page.waitForTimeout(1000)
  await expect(page).toClick("cds-button", { text: "Rollback" });
  await expect(page).toMatch("(current: 3)");

  await expect(page).toSelect("cds-select > select", "1");
  await expect(page).toClick("cds-modal-actions cds-button", { text: "Rollback" });

  // Check revision
  await page.waitForTimeout(1000)
  await expect(page).toClick("cds-button", { text: "Rollback" });
  await expect(page).toMatch("(current: 4)");

});
