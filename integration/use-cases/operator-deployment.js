const utils = require("./lib/utils");
const path = require("path");

// The operator may take some minutes to be created
jest.setTimeout(360000);

test("Deploys an Operator", async () => {
  // ODIC login
  await page.goto(getUrl("/#/c/default/ns/kubeapps/operators"));
  await page.waitForNavigation();
  await expect(page).toClick("cds-button", { text: "Login via OIDC Provider" });
  await page.waitForNavigation();
  await expect(page).toClick(".dex-container button", { text: "Log in with Email" });
  await page.waitForNavigation();
  await page.type("input[id=\"login\"]", "kubeapps-operator@example.com");
  await page.type("input[id=\"password\"]", "password");
  await page.waitForSelector("#submit-login", { visible: true, timeout: 10000 });
  await page.evaluate((selector) => document.querySelector(selector).click(), "#submit-login");
  await page.waitForSelector(".kubeapps-header-content", { visible: true, timeout: 10000 });
  await page.goto(getUrl("/#/c/default/ns/kubeapps/operators"));

  // wait for the loading msg to disappear
  await page.waitForFunction(() => !document.querySelector(".margin-t-xxl"));

  // Browse operator
  await expect(page).toClick("a", { text: "prometheus" });

  await utils.retryAndRefresh(page, 3, async () => {
    // Sometimes this fails with: TypeError: Cannot read property 'click' of null
    await expect(page).toClick("cds-button", { text: "Deploy" });
  });

  const isAlreadyDeployed = await page.evaluate(() => document.querySelector('cds-button[disabled]') !== null);

  if (!isAlreadyDeployed) {
    // Deploy the Operator
    await expect(page).toClick("cds-button", { text: "Deploy" });

    // wait for the loading msg to disappear
    await page.waitForFunction(() => !document.querySelector(".margin-t-xxl"));

    await utils.retryAndRefresh(page, 4, async () => {
      // The CSV takes a bit to get populated
      await expect(page).toMatch("Installed", { timeout: 10000 });
    });
  } else {
    console.log("Warning: the operator has already been deployed")
  }

  // Wait for the operator to be ready to be used
  await expect(page).toClick("a", { text: "Catalog" });

  await utils.retryAndRefresh(page, 30, async () => {
    await expect(page).toMatch("Operators", { timeout: 10000 });

    // Filter out charts to search only for the prometheus operator
    await expect(page).toClick("label", { text: "Operators" });

    await expect(page).toMatch("Prometheus");

    await expect(page).toClick(".info-card-header", { text: "Prometheus" });
  });

  await utils.retryAndRefresh(page, 2, async () => {
    // Found the error "prometheuses.monitoring.coreos.com not found in the definition of prometheusoperator"
    await expect(page).toMatch("Deploy", { timeout: 10000 });
  });

  await utils.retryAndRefresh(page, 5, async () => {
    await expect(page).toClick("cds-button", { text: "Deploy" });

    await expect(page).toMatch("Installation Values", { timeout: 20000 });
  }, "operator-view");

  // Update
  await expect(page).toClick("cds-button", { text: "Update" });

  await utils.retryAndRefresh(page, 2, async () => {
    await expect(page).toMatch("creationTimestamp", { timeout: 10000 });
  });
  // 
  await expect(page).toClick("cds-button", { text: "Deploy" });

  await utils.retryAndRefresh(page, 2, async () => {
    await expect(page).toMatch("Installation Values", { timeout: 10000 });
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
  } catch (e) {
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
