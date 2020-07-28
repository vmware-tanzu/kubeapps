const utils = require("./lib/utils");

test("Creates a registry", async () => {
  await page.goto(getUrl("/#/c/default/ns/kubeapps/config/repos"));

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

  await page.type("#kubeapps-repo-url", "https://charts.gitlab.io/");

  // Similar to the above click for an App Repository, the click on
  // the Install Repo doesn't always register (in fact, from the
  // screenshot on failure, it appears to focus the button only (hover css applied)
  await expect(page).toClick("button", { text: "Install Repo" });
  await expect(page).toClick("a", { text: "my-repo" });

  await utils.retryAndRefresh(page, 3, async () => {
    await expect(page).toMatch("gitlab-runner", { timeout: 2000 });
  });
});
