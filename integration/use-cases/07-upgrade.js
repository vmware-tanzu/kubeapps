const utils = require("./lib/utils");
const testName = "07-upgrade";

test("Upgrades an application", async () => {
  await utils.login(
    page,
    process.env.USE_MULTICLUSTER_OIDC_ENV,
    "/#/c/default/ns/default/catalog?Repository=bitnami",
    process.env.EDIT_TOKEN,
    "kubeapps-operator@example.com",
    "password",
  );

  await expect(page).toMatchElement("a", { text: "Apache HTTP Server", timeout: 60000 });
  await expect(page).toClick("a", { text: "Apache HTTP Server" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  let currentPackageVersion = "";

  await utils.retryAndRefresh(
    page,
    3,
    async () => {
      // wait to load every pkg version and get the current version
      await new Promise(r => setTimeout(r, 3000));

      // get the current pkg version (the selected one)
      const currentPackageVersionElement = await expect(page).toMatchElement(
        'select[name="chart-versions"] option:checked',
      );
      const currentPackageVersionElementContent = await currentPackageVersionElement.getProperty(
        "textContent",
      );
      const currentPackageVersionValue = await currentPackageVersionElementContent.jsonValue();
      currentPackageVersion = currentPackageVersionValue.split(" ")[0];

      // get the latest pkg version (the first one)
      const latestPackageVersionElements = await page.$$('select[name="chart-versions"] option');
      const latestPackageVersionElementContent = await latestPackageVersionElements[0].getProperty(
        "textContent",
      );
      const latestPackageVersionValue = await latestPackageVersionElementContent.jsonValue();
      latestPackageVersion = latestPackageVersionValue.split(" ")[0];

      expect(currentPackageVersion).not.toBe("");
    },
    testName,
  );

  // select the same pkg version
  await expect(page).toSelect('select[name="chart-versions"]', currentPackageVersion, {
    delay: 3000,
  });

  await new Promise(r => setTimeout(r, 500));

  await utils.retryAndRefresh(
    page,
    3,
    async () => {
      await expect(page).toMatch(currentPackageVersion);
    },
    testName,
  );

  await expect(page).toMatchElement("input[type='number']");
  // Increase the number of replicas
  await page.focus("input[type='number']");
  await page.keyboard.press("Backspace");
  await page.keyboard.type("2");

  await new Promise(r => setTimeout(r, 500));

  // Check that the Changes tab reflects the change
  await expect(page).toClick("li", { text: "Changes" });
  await expect(page).toMatch("replicaCount: 2");

  await expect(page).toMatchElement("#releaseName", { text: "" });
  await page.type("#releaseName", utils.getRandomName("my-app"));

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await utils.retryAndRefresh(
    page,
    2,
    async () => {
      await expect(page).toMatch("Update Now", { timeout: 60000 });
    },
    testName,
  );

  await expect(page).toClick("cds-button", { text: "Upgrade" });

  await new Promise(r => setTimeout(r, 1000));

  // Verify that the form contains the old version
  await utils.retryAndRefresh(
    page,
    3,
    async () => {
      await expect(page).toMatch(currentPackageVersion);
    },
    testName,
  );

  await expect(page).toMatchElement("input[type='number']", { value: 2 });

  await utils.retryAndRefresh(
    page,
    3,
    async () => {
      // select the latest pkg version
      await expect(page).toSelect('select[name="chart-versions"]', latestChartVersion, {
        delay: 3000,
      });
      // wait to load every pkg version and get the current version
      await new Promise(r => setTimeout(r, 3000));

      // get the current pkg version (the selected one after being upgraded)
      const upgradedPackageVersionElement = await expect(page).toMatchElement(
        'select[name="chart-versions"] option:checked',
      );
      const upgradedPackageVersionElementContent = await upgradedPackageVersionElement.getProperty(
        "textContent",
      );
      const upgradedPackageVersionValue = await upgradedPackageVersionElementContent.jsonValue();
      currentPackageVersion = upgradedPackageVersionValue.split(" ")[0];
      expect(upgradedPackageVersionValue).toEqual(latestPackageVersion);
    },
    testName,
  );

  await expect(page).toMatchElement("input[type='number']", { value: 2 });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Up to date", { timeout: 60000 });
});
