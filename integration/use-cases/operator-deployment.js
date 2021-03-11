const utils = require("./lib/utils");
const {
  screenshotsFolder,
} = require("../args");
const path = require("path");

// The operator may take some minutes to be created
jest.setTimeout(360000);

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

  await utils.retryAndRefresh(page, 2, async () => {
    // Sometimes this fails with: TypeError: Cannot read property 'click' of null
    await expect(page).toClick("cds-button", { text: "Deploy" });
  });

  // Deploy the Operator
  await expect(page).toClick("cds-button", { text: "Deploy" });

  await utils.retryAndRefresh(page, 2, async () => {
    // The CSV takes a bit to get populated
    await expect(page).toMatch("Installed");
  });

  // Wait for the operator to be ready to be used
  await expect(page).toClick("a", { text: "Catalog" });

  await utils.retryAndRefresh(page, 30, async () => {
    await expect(page).toMatch("Operators", {timeout: 10000});

    // Filter out charts to search only for the prometheus operator
    await expect(page).toClick("label", { text: "Operators" });

    await expect(page).toMatch("Prometheus");

    await expect(page).toClick(".info-card-header", { text: "Prometheus" });
  });

  await utils.retryAndRefresh(page, 2, async () => {
    // Found the error "prometheuses.monitoring.coreos.com not found in the definition of prometheusoperator"
    await expect(page).toMatch("Deploy", {timeout: 10000});
  });


  await utils.retryAndRefresh(page, 5, async () => {
    await expect(page).toClick("cds-button", { text: "Deploy" });

    await expect(page).toMatch("Installation Values", {timeout: 20000});
  }, "operator-view");

  // Update
  await expect(page).toClick("cds-button", { text: "Update" });

  await utils.retryAndRefresh(page, 2, async () => {
    await expect(page).toMatch("creationTimestamp", {timeout: 10000});
  });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await utils.retryAndRefresh(page, 2, async () => {
    await expect(page).toMatch("Installation Values", {timeout: 10000});
  });

  // Delete
  await expect(page).toClick("cds-button", { text: "Delete" });

  await expect(page).toMatch("Are you sure you want to delete the resource?");

  try {
    // TODO(andresmgot): Remove this line once 2.3 is released
    await expect(page).toClick(
      "div.modal-dialog.modal-md > div > div.modal-body > div > div > cds-button:nth-child(2)",
      {
        text: "Delete",
      }
    );  
  } catch(e) {
    await expect(page).toClick(
      "#root > section > main > div > div > section > cds-modal > cds-modal-actions > button.btn.btn-danger",
      {
        text: "Delete",
      }
    );
  }

  // Goes back to application list
  await expect(page).toMatch("Applications", { timeout: 60000 });
});
