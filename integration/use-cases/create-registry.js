const { setDefaultOptions } = require("expect-puppeteer");

setDefaultOptions({ timeout: 8000 });
jest.setTimeout(120000);

test("Deploys an application with the values by default", async () => {
  page.setDefaultTimeout(8000);
  await page.goto(getUrl("/#/login"));

  await expect(page).toFillForm("form", {
    token: process.env.ADMIN_TOKEN
  });

  await expect(page).toClick("button", { text: "Login" });

  // Double click to show configuration
  await expect(page).toClick("a", { text: "Configuration" });
  await expect(page).toClick("a", { text: "Configuration" });

  await expect(page).toClick("a", { text: "App Repositories" });

  await expect(page).toClick("button", { text: "Add App Repository" });

  await page.type("#kubeapps-repo-name", "my-repo");

  await page.type("#kubeapps-repo-url", "https://charts.gitlab.io/");

  await expect(page).toClick("button", { text: "Install Repo" });

  await expect(page).toClick("a", { text: "my-repo" });

  let retries = 3;
  while (retries > 0) {
    try {
      await expect(page).toMatch("gitlab", { timeout: 2000 });
      break;
    } catch (e) {
      // Refresh since the chart will get a bit of time to populate
      await page.reload({ waitUntil: ["networkidle0", "domcontentloaded"] });
    } finally {
      retries--;
    }
  }
});
