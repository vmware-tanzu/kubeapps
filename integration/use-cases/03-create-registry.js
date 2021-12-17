const utils = require("./lib/utils");
const testName = "03-create-registry";

test("Creates a registry", async () => {
  await utils.login(
    page,
    process.env.USE_MULTICLUSTER_OIDC_ENV,
    "/#/c/default/ns/kubeapps/config/repos",
    process.env.ADMIN_TOKEN,
    "kubeapps-operator@example.com",
    "password",
  );

  await expect(page).toMatchElement("cds-button", { text: "Add App Repository" });
  await expect(page).toClick("cds-button", { text: "Add App Repository" });

  await page.type('input[placeholder="example"]', "my-repo");

  await page.type(
    'input[placeholder="https://charts.example.com/stable"]',
    "https://charts.gitlab.io/",
  );

  // Similar to the above click for an App Repository, the click on
  // the Install Repo doesn't always register (in fact, from the
  // screenshot on failure, it appears to focus the button only (hover css applied)
  await expect(page).toClick("cds-button", { text: "Install Repo" });
  await utils.retryAndRefresh(
    page,
    3,
    async () => {
      // TODO(andresmgot): In theory, there is no need to refresh but sometimes the repo
      // does not appear
      await expect(page).toClick("a", { text: "my-repo" });
    },
    testName,
  );

  await utils.retryAndRefresh(
    page,
    3,
    async () => {
      await expect(page).toMatch("gitlab-runner");
    },
    testName,
  );
});
