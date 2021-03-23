const utils = require("./lib/utils");

test("Upgrades an application", async () => {
  // ODIC login
  await Promise.all([
    await page.goto(getUrl("/#/c/default/ns/default/catalog?Repository=bitnami")),
    await page.waitForNavigation(),
    await expect(page).toClick("cds-button", { text: "Login via OIDC Provider" }),
    await page.waitForNavigation(),
    await expect(page).toClick(".dex-container button", { text: "Log in with Email" }),
    await page.waitForNavigation(),
    await page.type("input[id=\"login\"]", "kubeapps-operator@example.com"),
    await page.type("input[id=\"password\"]", "password"),
    await page.click("#submit-login"),
    await page.waitForNavigation({ waitUntil: 'networkidle2' }),
    await page.goto(getUrl("/#/c/default/ns/default/config/repos")),
    await page.waitForNavigation(),
    await page.goto(getUrl("/#/c/default/ns/default/catalog?Repository=bitnami")),
  ]);

  await expect(page).toMatch("apache", { timeout: 60000 });

  await expect(page).toClick("a", { text: "apache" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  let latestChartVersion = "";

  await utils.retryAndRefresh(page, 3, async () => {
    await new Promise((r) => setTimeout(r, 1000));

    const chartVersionElement = await expect(page).toMatchElement(
      'select[name="chart-versions"]'
    );
    const chartVersionElementContent = await chartVersionElement.getProperty(
      "textContent"
    );
    const chartVersionValue = await chartVersionElementContent.jsonValue();
    latestChartVersion = chartVersionValue.split(" ")[0];
    expect(latestChartVersion).not.toBe("");
  });

  await expect(page).toSelect('select[name="chart-versions"]', "7.3.2");

  await new Promise((r) => setTimeout(r, 500));

  await utils.retryAndRefresh(page, 3, async () => {
    await expect(page).toMatch("7.3.2", { timeout: 10000 });
  });

  await expect(page).toMatchElement("input[type='number']");
  // Increase the number of replicas
  await page.focus("input[type='number']");
  await page.keyboard.press("Backspace");
  await page.keyboard.type("2");

  await new Promise((r) => setTimeout(r, 500));

  // Check that the Changes tab reflects the change
  await expect(page).toClick("li", { text: "Changes" });
  await expect(page).toMatch("replicaCount: 2");

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Update Now", { timeout: 60000 });

  await expect(page).toClick("cds-button", { text: "Upgrade" });

  await new Promise((r) => setTimeout(r, 1000));

  // Verify that the form contains the old version
  await utils.retryAndRefresh(page, 3, async () => {
    await expect(page).toMatch("7.3.2", { timeout: 10000 });
  });

  await expect(page).toMatchElement("input[type='number']", { value: 2 });

  await utils.retryAndRefresh(page, 3, async () => {
    await expect(page).toSelect(
      'select[name="chart-versions"]',
      latestChartVersion
    );

    await new Promise((r) => setTimeout(r, 1000));

    // Ensure that the new value is selected
    const chartVersionElement = await expect(page).toMatchElement(
      '.upgrade-form-version-selector select[name="chart-versions"]'
    );
    const chartVersionElementContent = await chartVersionElement.getProperty("value");
    const chartVersionValue = await chartVersionElementContent.jsonValue();
    expect(chartVersionValue).toEqual(latestChartVersion);
  });

  await expect(page).toMatchElement("input[type='number']", { value: 2 });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Up to date", { timeout: 60000 });
});
