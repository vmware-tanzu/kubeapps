test("Upgrades an application", async () => {
  await page.goto(getUrl("/#/c/default/ns/default/catalog/bitnami"));

  await expect(page).toFillForm("form", {
    token: process.env.EDIT_TOKEN
  });

  await expect(page).toClick("button", { text: "Login" });

  await expect(page).toClick("a", { text: "apache" });

  await expect(page).toClick("button", { text: "Deploy" });

  const chartVersionElement = await expect(page).toMatchElement(
    "#chartVersion"
  );
  const chartVersionElementContent = await chartVersionElement.getProperty(
    "textContent"
  );
  const chartVersionValue = await chartVersionElementContent.jsonValue();
  const latestChartVersion = chartVersionValue.split(" ")[0];

  await expect(page).toSelect("#chartVersion", "7.3.2");
  await expect(page).toMatchElement("#replicaCount-1");

  // Increase the number of replicas
  await page.focus("#replicaCount-1");
  await page.keyboard.press("Backspace");
  await page.keyboard.type("2");

  // Check that the Changes tab reflects the change
  await expect(page).toClick("li", { text: "Changes" });
  await expect(page).toMatch("replicaCount: 2");

  await expect(page).toClick(".button-primary");

  await expect(page).toMatch("Update Available", { timeout: 60000 });

  await expect(page).toClick(".upgrade-button");

  await expect(page).toMatchElement("#replicaCount-1", { value: 2 });

  await expect(page).toSelect("#chartVersion", latestChartVersion);

  await expect(page).toMatchElement("#replicaCount-1", { value: 2 });

  // From comments at https://github.com/puppeteer/puppeteer/issues/3347, try using a
  // selector rather than element / text for click event.
  await expect(page).toClick(".button-primary");

  await expect(page).toMatch("Up to date", { timeout: 10000 });
});
