const utils = require("./lib/utils");

test("Upgrades an application", async () => {
  await utils.login(
    page,
    process.env.USE_MULTICLUSTER_OIDC_ENV,
    "/#/c/default/ns/default/catalog?Repository=bitnami",
    process.env.EDIT_TOKEN,
    "kubeapps-operator@example.com",
    "password",
  );

  await expect(page).toMatch("apache", { timeout: 60000 });

  await expect(page).toClick("a", { text: "apache" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  let latestChartVersion = "";

  await utils.retryAndRefresh(page, 3, async () => {
    await new Promise(r => setTimeout(r, 1000));

    const chartVersionElement = await expect(page).toMatchElement('select[name="chart-versions"]');
    const chartVersionElementContent = await chartVersionElement.getProperty("textContent");
    const chartVersionValue = await chartVersionElementContent.jsonValue();
    latestChartVersion = chartVersionValue.split(" ")[0];
    expect(latestChartVersion).not.toBe("");
  });

  await expect(page).toSelect('select[name="chart-versions"]', "7.3.2");

  await new Promise(r => setTimeout(r, 500));

  await utils.retryAndRefresh(page, 3, async () => {
    await expect(page).toMatch("7.3.2");
  });

  await expect(page).toMatchElement("input[type='number']");
  // Increase the number of replicas
  await page.focus("input[type='number']");
  await page.keyboard.press("Backspace");
  await page.keyboard.type("2");

  await new Promise(r => setTimeout(r, 500));

  // Check that the Changes tab reflects the change
  await expect(page).toClick("li", { text: "Changes" });
  await expect(page).toMatch("replicaCount: 2");

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Update Now", { timeout: 60000 });

  await expect(page).toClick("cds-button", { text: "Upgrade" });

  await new Promise(r => setTimeout(r, 1000));

  // Verify that the form contains the old version
  await utils.retryAndRefresh(page, 3, async () => {
    await expect(page).toMatch("7.3.2");
  });

  await expect(page).toMatchElement("input[type='number']", { value: 2 });

  await utils.retryAndRefresh(page, 3, async () => {
    await expect(page).toSelect('select[name="chart-versions"]', latestChartVersion);

    await new Promise(r => setTimeout(r, 1000));

    // Ensure that the new value is selected
    const chartVersionElement = await expect(page).toMatchElement(
      '.upgrade-form-version-selector select[name="chart-versions"]',
    );
    const chartVersionElementContent = await chartVersionElement.getProperty("value");
    const chartVersionValue = await chartVersionElementContent.jsonValue();
    expect(chartVersionValue).toEqual(latestChartVersion);
  });

  await expect(page).toMatchElement("input[type='number']", { value: 2 });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Up to date", { timeout: 60000 });
});
