const axios = require("axios");
const utils = require("./lib/utils");

test("Creates a private registry", async () => {
  await page.goto(getUrl("/#/c/default/ns/default/config/repos"));

  await expect(page).toFillForm("form", {
    token: process.env.ADMIN_TOKEN,
  });

  await expect(page).toClick("button", { text: "Login" });

  await expect(page).toClick("button", { text: "Add App Repository" });

  try {
    await expect(page).toMatch("Install Repo");
  } catch (e) {
    // The Modal sometimes fail to be opened, click the button again
    await expect(page).toClick("button", { text: "Add App Repository" });

    await expect(page).toMatch("Install Repo");
  }

  await page.type("#kubeapps-repo-name", "my-repo");

  await page.type(
    "#kubeapps-repo-url",
    "http://chartmuseum-chartmuseum.kubeapps:8080"
  );

  await expect(page).toClick("label", { text: "Basic Auth" });

  // Credentials from e2e-test.sh
  await page.type("#kubeapps-repo-username", "admin");
  await page.type("#kubeapps-repo-password", "password");

  // Open form to create a new secret
  const randomNumber = Math.floor(Math.random() * Math.floor(100));
  const secret = "my-repo-secret" + randomNumber;
  await expect(page).toClick("button", { text: "Add new credentials" });
  await page.type("#kubeapps-docker-cred-secret-name", secret);
  await page.type(
    "#kubeapps-docker-cred-server",
    "https://index.docker.io/v1/"
  );
  await page.type("#kubeapps-docker-cred-username", "user");
  await page.type("#kubeapps-docker-cred-password", "password");
  await page.type("#kubeapps-docker-cred-email", "user@example.com");
  await expect(page).toClick("button", { text: "Submit" });

  // Select the new secret
  await expect(page).toClick("label", { text: secret });

  // Similar to the above click for an App Repository, the click on
  // the Install Repo doesn't always register (in fact, from the
  // screenshot on failure, it appears to focus the button only (hover css applied)
  await expect(page).toClick("button", { text: "Install Repo" });
  await expect(page).toClick("a", { text: "my-repo" });

  await utils.retryAndRefresh(page, 3, async () => {
    await expect(page).toMatch("apache", { timeout: 2000 });
  });

  await expect(page).toClick("a", { text: "apache", timeout: 60000 });

  await expect(page).toClick("button", { text: "Deploy" });

  await expect(page).toSelect("#chartVersion", "7.3.15");
  const appName = "my-app" + randomNumber;
  await page.type("#releaseName", appName);

  await expect(page).toClick("button", { text: "Submit" });

  await expect(page).toMatch("Update Available", { timeout: 60000 });

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
  await expect(page).toClick(".upgrade-button");

  await expect(page).toSelect("#chartVersion", "7.3.16");

  await expect(page).toClick("li", { text: "Changes" });

  await expect(page).toMatch("tag: 2.4.43-debian-10-r54");

  await expect(page).toClick(".button-primary");

  await expect(page).toMatch("Up to date", { timeout: 60000 });
  await expect(page).toMatch("Ready", { timeout: 60000 });
});
