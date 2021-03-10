const utils = require("./lib/utils");

test("Creates a registry", async () => {
  await page.goto(getUrl("/#/c/default/ns/kubeapps/config/repos"));

  await expect(page).toFillForm("form", {
    token: process.env.ADMIN_TOKEN,
  });

  await page.evaluate(() =>
    document.querySelector("#login-submit-button").click()
  );

  await expect(page).toClick("cds-button", { text: "Add App Repository" });

  await page.type("input[placeholder=\"example\"]", "my-repo");

  await page.type("input[placeholder=\"https://charts.example.com/stable\"]", "https://charts.gitlab.io/");

  // Similar to the above click for an App Repository, the click on
  // the Install Repo doesn't always register (in fact, from the
  // screenshot on failure, it appears to focus the button only (hover css applied)
  await expect(page).toClick("cds-button", { text: "Install Repo" });
  await utils.retryAndRefresh(page, 3, async () => {
    // TODO(andresmgot): In theory, there is no need to refresh but sometimes the repo
    // does not appear
    await expect(page).toClick("a", { text: "my-repo" });
  });

  await utils.retryAndRefresh(page, 3, async () => {
    await expect(page).toMatch("gitlab-runner", { timeout: 10000 });
  });
});
