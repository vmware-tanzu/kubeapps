test("Deploys an Operator", async () => {
  await page.goto(getUrl("/#/ns/kubeapps/operators"));

  await expect(page).toFillForm("form", {
    token: process.env.EDIT_TOKEN,
  });

  await expect(page).toClick("button", { text: "Login" });

  // Browse operator
  await expect(page).toClick("a", { text: "prometheus" });

  await expect(page).toClick("button", { text: "Deploy" });

  // TODO(andresmgot) Fill deployment form when it's ready
  // For now we assume is already installed
  await expect(page).toMatch(
    "Install the operator by running the following command"
  );

  // Close modal
  await expect(page).toClick("h1", { text: "prometheus" });

  // Deploy Operator instance
  await expect(page).toClick("a", { text: "Catalog" });

  await expect(page).toMatch("Charts");

  // Filter out charts
  await expect(page).toClick(".checkbox", { text: "Charts" });

  await expect(page).toClick("a", { text: "Prometheus" });

  await expect(page).toClick("button", { text: "Submit" });

  await expect(page).toMatch("Ready", { timeout: 60000 });

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
