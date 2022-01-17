const utils = require("./lib/utils");
const testName = "07-upgrade";
const { screenshotsFolder } = require("../args");
const path = require("path");

test("Upgrades an application", async () => {
  await utils.login(
    page,
    process.env.USE_MULTICLUSTER_OIDC_ENV,
    "/#/c/default/ns/default/catalog?Repository=bitnami",
    process.env.EDIT_TOKEN,
    "kubeapps-operator@example.com",
    "password",
  );

  await expect(page).toMatchElement("a", { text: "foo apache chart for CI", timeout: 60000 });
  await expect(page).toClick("a", { text: "foo apache chart for CI" });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  let initialPackageVersion = "";
  let currentPackageVersion = "";

  await utils.retryAndRefresh(
    page,
    3,
    async () => {
      // get the latest pkg version (the first one)
      const latestPackageVersionElements = await page.$$('select[name="package-versions"] option', {
        delay: 2000,
      });
      const latestPackageVersionElementContent = await latestPackageVersionElements[0].getProperty(
        "textContent",
      );
      const latestPackageVersionValue = await latestPackageVersionElementContent.jsonValue();
      latestPackageVersion = latestPackageVersionValue.split(" ")[0];

      // get an older version to be installed (the second one)
      const initialPackageVersionElements = await page.$$(
        'select[name="package-versions"] option',
        {
          delay: 2000,
        },
      );
      const initialPackageVersionElementContent =
        await initialPackageVersionElements[1].getProperty("textContent");
      const initialPackageVersionValue = await initialPackageVersionElementContent.jsonValue();
      initialPackageVersion = initialPackageVersionValue.split(" ")[0];

      expect(initialPackageVersion).not.toBe("");
    },
    testName,
  );

  console.log(
    `apache initialPackageVersion: ${initialPackageVersion}, latestPackageVersion: ${latestPackageVersion}`,
  );

  // select the initialPackageVersion
  await expect(page).toSelect('select[name="package-versions"]', initialPackageVersion);
  await new Promise(r => setTimeout(r, 500));

  await utils.retryAndRefresh(
    page,
    3,
    async () => {
      // Check if the page contains the selected version
      await expect(page).toMatch(initialPackageVersion);

      let screenshotFilename = `../${screenshotsFolder}/${testName}-check-initial-selected-version.png`;
      console.log(`Saving screenshot to ${screenshotFilename}`);
      await page.screenshot({
        path: path.join(__dirname, screenshotFilename),
      });
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
  await page.type("#releaseName", utils.getRandomName("my-app-for-07-upgrade"));

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await utils.retryAndRefresh(
    page,
    2,
    async () => {
      // Since we installed an older version, an update message should appear
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
      await expect(page).toMatch(initialPackageVersion);
    },
    testName,
  );

  await expect(page).toMatchElement("input[type='number']", { value: 2 });

  await utils.retryAndRefresh(
    page,
    3,
    async () => {
      // Select the latest pkg version
      await expect(page).toSelect('select[name="package-versions"]', latestPackageVersion);
      // Immediately after selecting the version, the checked selection is still
      // the unchanged. Not sure if it's due to the browser DOM needing
      // milliseconds to update, or react's own rendering, but without the
      // following 0.5s delay, we pick up the previously selected version
      // (initialPackageVersion) as the checked one.
      await new Promise(r => setTimeout(r, 500));

      // get the current pkg version (the selected one after being upgraded)
      const upgradedPackageVersionElement = await expect(page).toMatchElement(
        'select[name="package-versions"] option:checked',
      );
      const upgradedPackageVersionElementContent = await upgradedPackageVersionElement.getProperty(
        "textContent",
      );
      const upgradedPackageVersionValue = await upgradedPackageVersionElementContent.jsonValue();
      upgradedPackageVersion = upgradedPackageVersionValue.split(" ")[0];

      // If the upgrade was successful, the upgraded version should match the latest version
      expect(upgradedPackageVersion).toEqual(latestPackageVersion);
    },
    testName,
  );

  await expect(page).toMatchElement("input[type='number']", { value: 2 });

  await expect(page).toClick("cds-button", { text: "Deploy" });

  await expect(page).toMatch("Up to date", { timeout: 60000 });
});
