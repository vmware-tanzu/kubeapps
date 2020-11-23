test("Upgrades an application", async () => {
  await page.goto(getUrl("/#/c/default/ns/default/catalog?Repository=bitnami"));

  await expect(page).toFillForm("form", {
    token: process.env.EDIT_TOKEN,
  });

  await page.evaluate(() =>
    document.querySelector("#login-submit-button").click()
  );

  await expect(page).toClick("a", { text: "apache" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  const chartVersionElement = await expect(page).toMatchElement(
    'select[name="chart-versions"]'
  );
  const chartVersionElementContent = await chartVersionElement.getProperty(
    "textContent"
  );
  const chartVersionValue = await chartVersionElementContent.jsonValue();
  const latestChartVersion = chartVersionValue.split(" ")[0];

  await expect(page).toSelect('select[name="chart-versions"]', "7.3.2");
  await expect(page).toMatchElement("input[type='number']");

  // Increase the number of replicas
  await page.focus("input[type='number']");
  await page.keyboard.press("Backspace");
  await page.keyboard.type("2");

  // Check that the Changes tab reflects the change
  await expect(page).toClick("li", { text: "Changes" });
  await expect(page).toMatch("replicaCount: 2");

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Update Now", { timeout: 60000 });

  await expect(page).toClick("cds-button", { text: "Upgrade" });

  await expect(page).toMatchElement("input[type='number']", { value: 2 });

  await expect(page).toSelect(
    'select[name="chart-versions"]',
    latestChartVersion
  );

  await expect(page).toMatchElement("input[type='number']", { value: 2 });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Up to date", { timeout: 60000 });
});
