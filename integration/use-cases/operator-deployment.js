const utils = require("./lib/utils");

test("Deploys an Operator", async () => {
  await page.goto(getUrl("/#/c/default/ns/kubeapps/operators"));

  await expect(page).toFillForm("form", {
    token: process.env.ADMIN_TOKEN,
  });

  await expect(page).toClick("button", { text: "Login" });

  // Browse operator
  await expect(page).toClick("a", { text: "prometheus" });

  await expect(page).toClick("button", { text: "Deploy" });

  // Deploy the Operator
  await expect(page).toClick("button", { text: "Submit" });

  await utils.retryAndRefresh(page, 10, async () => {
    // The CSV takes a bit to get populated
    await expect(page).toMatch("Installed", { timeout: 10000 });
  });

  // Wait for the operator to be ready to be used
  await expect(page).toClick("a", { text: "Catalog" });

  await utils.retryAndRefresh(page, 10, async () => {
    // Filter out charts to search only for the prometheus operator
    await expect(page).toClick(".checkbox", { text: "Charts" });

    await expect(page).toMatch("Prometheus");
  });

  // Deploy Operator instance
  await expect(page).toClick("a", { text: "Catalog" });

  await expect(page).toMatch("Charts");

  // Filter out charts
  await expect(page).toClick(".checkbox", { text: "Charts" });

  await expect(page).toClick("a", { text: "Prometheus" });

  await expect(page).toClick("button", { text: "Submit" });

  await expect(page).toMatch("Installation Values", { timeout: 60000 });

  // Update
  await expect(page).toClick("button", { text: "Update" });

  await expect(page).toClick("button", { text: "Submit" });

  await expect(page).toMatch("Ready", { timeout: 60000 });

  // Delete
  await expect(page).toClick("button", { text: "Delete" });

  await expect(page).toMatch("Are you sure you want to delete this?");

  await expect(page).toClick("button.button.button-primary.button-danger", {
    text: "Delete",
  });

  // Goes back to application list
  await expect(page).toMatch("Deploy App", { timeout: 60000 });
});
