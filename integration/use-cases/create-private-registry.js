const axios = require("axios");
const utils = require("./lib/utils");

test("Creates a private registry", async () => {
  // ODIC login
  var token;
  page.on('response', response => {
    if (response.status() >= 400) {
      console.log("ERROR: ", response.status() + " " + response.url());
    }
    token = response.headers()["authorization"] || token;
  });
  await page.goto(getUrl("/#/c/default/ns/default/config/repos"));
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
  console.log("Token after OIDC authentication: " + token);

  await page.goto(getUrl("/#/c/default/ns/kubeapps/config/repos"));
  await page.waitForFunction(() => !document.querySelector(".margin-t-xxl")); // wait for the loading msg to disappear
  
  await expect(page).toClick("cds-button", { text: "Add App Repository", timeout: 10000 });

  await page.evaluate(() => {
    [...document.querySelectorAll('cds-modal-header')].find(element => element.outerText.includes("Add an App Repository"));
  });

  const randomNumber = Math.floor(Math.random() * Math.floor(100));
  const repoName = "my-repo-" + randomNumber;
  await page.type("input[placeholder=\"example\"]", repoName);

  await page.type(
    "input[placeholder=\"https://charts.example.com/stable\"]",
    "http://chartmuseum-chartmuseum.kubeapps:8080"
  );

  await expect(page).toClick("label", { text: "Basic Auth", timeout: 10000 });
  await expect(page).toClick("label", { text: "Basic Auth", timeout: 10000 });
  await expect(page).toClick("label", { text: "Basic Auth", timeout: 10000 });
  await expect(page).toClick("label", { text: "Basic Auth", timeout: 10000 });


  // Credentials from e2e-test.sh
  await page.type("input[placeholder=\"Username\"]", "admin");
  await page.type("input[placeholder=\"Password\"]", "password");

  // Open form to create a new secret
  const secret = "my-repo-secret" + randomNumber;
  console.log("quiero Add new credentials, estan?"); console.log(await page.evaluate(() => document.body.innerHTML))
  
  await expect(page).toClick("cds-button", { text: "Add new credentials", timeout: 10000 });

  await page.type("input[placeholder=\"Secret\"]", secret);
  await page.type(
    "input[placeholder=\"https://index.docker.io/v1/\"]",
    "https://index.docker.io/v1/"
  );
  await page.type("input[placeholder=\"Username\"][value=\"\"]", "user");
  await page.type("input[placeholder=\"Password\"][value=\"\"]", "password");
  await page.type("input[placeholder=\"user@example.com\"]", "user@example.com");
  await expect(page).toClick(".secondary-input cds-button", { text: "Submit", timeout: 10000 });

  // Select the new secret
  await expect(page).toClick("label", { text: secret, timeout: 10000 });

  await expect(page).toClick("cds-button", { text: "Install Repo", timeout: 10000 });

  await expect(page).toClick("a", { text: repoName, timeout: 10000 });

  await utils.retryAndRefresh(page, 3, async () => {
    await expect(page).toMatch("apache", { timeout: 2000 });
  });

  await expect(page).toClick("a", { text: "apache", timeout: 60000 });

  await expect(page).toClick("cds-button", { text: "Deploy", timeout: 10000 });

  await expect(page).toSelect('select[name="chart-versions"]', "7.3.15");
  const appName = "my-app" + randomNumber;
  await page.type("#releaseName", appName);

  await expect(page).toMatch(/Deploy.*7.3.15/, { timeout: 10000 });

  await expect(page).toClick("cds-button", { text: "Deploy", timeout: 10000 });

  await expect(page).toMatch("Update Now", { timeout: 60000 });

  // Now that the deployment has been created, we check that the imagePullSecret
  // has been added. For doing so, we query the kubernetes API to get info of the
  // deployment
  const URL = getUrl(
    "/api/clusters/default/apis/apps/v1/namespaces/default/deployments"
  );
  const response = await axios.get(URL, {
    headers: { Authorization: `${token}` },
  });
  const deployment = response.data.items.find((deployment) => {
    return deployment.metadata.name.match(appName);
  });
  expect(deployment.spec.template.spec.imagePullSecrets).toEqual([
    { name: secret },
  ]);

  // Upgrade apache and verify.
  await expect(page).toClick("cds-button", { text: "Upgrade", timeout: 10000 });

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
  } catch (e) {
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

  await expect(page).toClick("li", { text: "Changes", timeout: 10000 });

  await expect(page).toMatch("tag: 2.4.43-debian-10-r54");

  await expect(page).toClick("cds-button", { text: "Deploy", timeout: 10000 });

  await expect(page).toMatch("Up to date", { timeout: 60000 });
  await expect(page).toMatch("Ready");
});
