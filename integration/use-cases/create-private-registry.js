const axios = require("axios");
const utils = require("./lib/utils");

test("Creates a private registry", async () => {
  await page.goto(getUrl("/#/c/default/ns/default/config/repos"));

  await expect(page).toFillForm("form", {
    token: process.env.ADMIN_TOKEN,
  });

  await page.evaluate(() =>
    document.querySelector("#login-submit-button").click()
  );

  await expect(page).toClick("cds-button", { text: "Add App Repository" });

  const randomNumber = Math.floor(Math.random() * Math.floor(100));
  const repoName = "my-repo-" + randomNumber;
  await page.type("input[placeholder=\"example\"]", repoName);

  await page.type(
    "input[placeholder=\"https://charts.example.com/stable\"]",
    "http://chartmuseum-chartmuseum.kubeapps:8080"
  );

  await expect(page).toClick("label", { text: "Basic Auth" });

  // Credentials from e2e-test.sh
  await page.type("input[placeholder=\"Username\"]", "admin");
  await page.type("input[placeholder=\"Password\"]", "password");

  // Open form to create a new secret
  const secret = "my-repo-secret" + randomNumber;
  try {
    // TODO(andresmgot): Remove this line once 2.3 is released
    await expect(page).toClick("cds-button", { text: "Add new credentials" });
  } catch(e) {
    await expect(page).toClick(".btn-info-outline", { text: "Add new credentials" });
  }
  await page.type("input[placeholder=\"Secret\"]", secret);
  await page.type(
    "input[placeholder=\"https://index.docker.io/v1/\"]",
    "https://index.docker.io/v1/"
  );
  await page.type("input[placeholder=\"Username\"][value=\"\"]", "user");
  await page.type("input[placeholder=\"Password\"][value=\"\"]", "password");
  await page.type("input[placeholder=\"user@example.com\"]", "user@example.com");
  try {
    // TODO(andresmgot): Remove this line once 2.3 is released
    await expect(page).toClick(".secondary-input cds-button", { text: "Submit" });
  } catch(e) {
    await expect(page).toClick(".btn-info-outline", { text: "Submit" });
  }

  // Select the new secret
  await expect(page).toClick("label", { text: secret });

  await expect(page).toClick("cds-button", { text: "Install Repo" });

  await expect(page).toClick("a", { text: repoName });

  await utils.retryAndRefresh(page, 3, async () => {
    await expect(page).toMatch("apache", { timeout: 2000 });
  });

  await expect(page).toClick("a", { text: "apache", timeout: 60000 });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toSelect('select[name="chart-versions"]', "7.3.15");
  const appName = "my-app" + randomNumber;
  await page.type("#releaseName", appName);

  await expect(page).toMatch(/Deploy.*7.3.15/, { timeout: 10000 });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Update Now", { timeout: 60000 });

  // Now that the deployment has been created, we check that the imagePullSecret
  // has been added. For doing so, we query the kubernetes API to get info of the
  // deployment
  const URL = getUrl(
    "/api/clusters/default/apis/apps/v1/namespaces/default/deployments"
  );
  const response = await axios.get(URL, {
    headers: { Authorization: `Bearer ${process.env.ADMIN_TOKEN}` },
  });
  const deployment = response.data.items.find((deployment) => {
    return deployment.metadata.name.match(appName);
  });
  expect(deployment.spec.template.spec.imagePullSecrets).toEqual([
    { name: secret },
  ]);

  // Upgrade apache and verify.
  await expect(page).toClick("cds-button", { text: "Upgrade" });

  let retries = 3;
  try {
    await new Promise((r) => setTimeout(r, 500));

    let chartVersionElement = await expect(page).toMatchElement(
      '.upgrade-form-version-selector select[name="chart-versions"]'
    );
    let chartVersionElementContent = await chartVersionElement.getProperty(
      "value"
    );
    let chartVersionValue = await chartVersionElementContent.jsonValue();
    expect(chartVersionValue).toEqual("7.3.15");
  } catch(e) {
    retries--;
    if (!retries) {
      throw e;
    }
  }

  // TODO(andresmgot): Avoid race condition for selecting the latest version
  // but going back to the previous version
  await new Promise((r) => setTimeout(r, 1000));

  await expect(page).toSelect(
    '.upgrade-form-version-selector select[name="chart-versions"]',
    "7.3.16"
  );

  await new Promise((r) => setTimeout(r, 1000));

  // Ensure that the new value is selected
  chartVersionElement = await expect(page).toMatchElement(
    '.upgrade-form-version-selector select[name="chart-versions"]'
  );
  chartVersionElementContent = await chartVersionElement.getProperty("value");
  chartVersionValue = await chartVersionElementContent.jsonValue();
  expect(chartVersionValue).toEqual("7.3.16");

  await expect(page).toClick("li", { text: "Changes" });

  await expect(page).toMatch("tag: 2.4.43-debian-10-r54");

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Up to date", { timeout: 60000 });
  await expect(page).toMatch("Ready");
});
