const utils = require("./lib/utils");

test("Deploys an Operator", async () => {
  await page.goto(getUrl("/#/c/default/ns/kubeapps/operators"));

  await expect(page).toFillForm("form", {
    token: process.env.ADMIN_TOKEN,
  });

  await page.evaluate(() =>
    document.querySelector("#login-submit-button").click()
  );

  // Browse operator
  await expect(page).toClick("a", { text: "prometheus" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  // Deploy the Operator
  await expect(page).toClick("cds-button", { text: "Deploy" });

  await utils.retryAndRefresh(page, 10, async () => {
    // The CSV takes a bit to get populated
    await expect(page).toMatch("Installed", { timeout: 10000 });
  });

  // Wait for the operator to be ready to be used
  await expect(page).toClick("a", { text: "Catalog" });

  await utils.retryAndRefresh(page, 10, async () => {
    // Filter out charts to search only for the prometheus operator
    await expect(page).toClick("label", { text: "Operators" });

    await expect(page).toMatch("Prometheus");

    await expect(page).toClick(".info-card-header", { text: "Prometheus" });
  });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Installation Values");

  // Update
  await expect(page).toClick("cds-button", { text: "Update" });

  await expect(page).toMatch("creationTimestamp");

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Ready");

  // Delete
  await expect(page).toClick("cds-button", { text: "Delete" });

  await expect(page).toMatch("Are you sure you want to delete the resource?");

  await expect(page).toClick(
    "div.modal-dialog.modal-md > div > div.modal-body > div > div > cds-button:nth-child(2)",
    {
      text: "Delete",
    }
  );

  // Goes back to application list
  await expect(page).toMatch("Applications", { timeout: 60000 });
});
