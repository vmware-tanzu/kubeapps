const axios = require("axios");
const utils = require("./lib/utils");

test("Creates a private registry", async () => {
  var token =
    process.env.USE_MULTICLUSTER_OIDC_ENV === "true" ? undefined : process.env.ADMIN_TOKEN;
  page.on("response", response => {
    // retrieves the token after the oidc flow, note this require "--set-authorization-header=true" flag to be enabled in oauth2proxy
    token = response.headers()["authorization"] || token;
  });

  await utils.login(
    page,
    process.env.USE_MULTICLUSTER_OIDC_ENV,
    "/#/c/default/ns/default/config/repos",
    process.env.ADMIN_TOKEN,
    "kubeapps-operator@example.com",
    "password",
  );

  await expect(page).toMatchElement("cds-button", { text: "Add App Repository" });
  await expect(page).toClick("cds-button", { text: "Add App Repository" });

  const randomNumber = Math.floor(Math.random() * Math.floor(100));
  const repoName = "my-repo-" + randomNumber;
  await page.type('input[placeholder="example"]', repoName);

  await page.type(
    'input[placeholder="https://charts.example.com/stable"]',
    "http://chartmuseum-chartmuseum.kubeapps:8080",
  );

  await expect(page).toClick("label", { text: "Basic Auth" });

  // Credentials from e2e-test.sh
  await page.type('input[placeholder="Username"]', "admin");
  await page.type('input[placeholder="Password"]', "password");

  // Open form to create a new secret
  const secret = "my-repo-secret" + randomNumber;

  await expect(page).toClick("button", { text: "Add new credentials" });

  await page.type('input[placeholder="Secret"]', secret);
  await page.type(
    'input[placeholder="https://index.docker.io/v1/"]',
    "https://index.docker.io/v1/",
  );
  await page.type('input[placeholder="Username"][value=""]', "user");
  await page.type('input[placeholder="Password"][value=""]', "password");
  await page.type('input[placeholder="user@example.com"]', "user@example.com");

  await expect(page).toClick("button", {
    text: "Submit",
  });

  // Select the new secret
  await expect(page).toMatch(secret);
  await expect(page).toSelect("form > cds-form-group > cds-select > select", secret);

  await expect(page).toClick("cds-button", { text: "Install Repo" });

  await expect(page).toClick("a", { text: repoName });

  await utils.retryAndRefresh(page, 3, async () => {
    await expect(page).toMatch("apache");
  });

  await expect(page).toMatchElement("a", { text: "apache", timeout: 60000 });
  await expect(page).toClick("a", { text: "apache" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toSelect('select[name="chart-versions"]', "7.3.15");
  const appName = "my-app" + randomNumber;
  await page.type("#releaseName", appName);

  await expect(page).toMatch(/Deploy.*7.3.15/);

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Update Now", { timeout: 60000 });

  // Now that the deployment has been created, we check that the imagePullSecret
  // has been added. For doing so, we query the kubernetes API to get info of the
  // deployment
  const URL = getUrl("/api/clusters/default/apis/apps/v1/namespaces/default/deployments");

  const cookies = await page.cookies();
  const axiosConfig = {
    headers: {
      Authorization: `${token}`,
      Cookie: `${cookies[0] ? cookies[0].name : ""}=${cookies[0] ? cookies[0].value : ""}`,
    },
  };
  const response = await axios.get(URL, axiosConfig);
  expect(response.status).toEqual(200);

  const deployment = response.data.items.find(deployment => {
    return deployment.metadata.name.match(appName);
  });
  expect(deployment.spec.template.spec.imagePullSecrets).toEqual([{ name: secret }]);

  // Upgrade apache and verify.
  await expect(page).toClick("cds-button", { text: "Upgrade" });

  let retries = 3;
  try {
    await new Promise(r => setTimeout(r, 500));

    let chartVersionElement = await expect(page).toMatchElement(
      '.upgrade-form-version-selector select[name="chart-versions"]',
    );
    let chartVersionElementContent = await chartVersionElement.getProperty("value");
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
  await new Promise(r => setTimeout(r, 1000));

  await expect(page).toSelect(
    '.upgrade-form-version-selector select[name="chart-versions"]',
    "7.3.16",
  );

  await new Promise(r => setTimeout(r, 1000));

  // Ensure that the new value is selected
  chartVersionElement = await expect(page).toMatchElement(
    '.upgrade-form-version-selector select[name="chart-versions"]',
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
