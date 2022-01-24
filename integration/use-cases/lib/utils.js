const { screenshotsFolder } = require("../../args");
const path = require("path");
const WAIT_EVENT_NETWORK = "networkidle0";
const WAIT_EVENT_DOM = "domcontentloaded";

module.exports = {
  takeScreenShot: async (fileName) => {
    let screenshotFilename = `../../${screenshotsFolder}/${fileName}.png`;
    console.log(`Saving screenshot to ${screenshotFilename}`);
    await page.screenshot({
      path: path.join(__dirname, screenshotFilename),
    });
  },
  retryAndRefresh: async (page, retries, toCheck, testName) => {
    let retriesLeft = retries;
    while (retriesLeft > 0) {
      try {
        await toCheck();
        break;
      } catch (e) {
        testName = testName || "unknown";
        await module.exports.takeScreenShot(`${testName}-${retries - retriesLeft}`);
        if (retriesLeft === 1) {
          // Unable to get it done
          throw e;
        }
        // Refresh since the package will get a bit of time to populate
        try {
          await page.reload({
            waitUntil: [WAIT_EVENT_NETWORK],
            timeout: 20000,
          });
        } catch (e) {
          // The reload may fail, ignore this try
          retriesLeft++;
        }
      } finally {
        retriesLeft--;
      }
    }
  },
  doAction: async (name, action) => {
    await Promise.all([
      action,
      page.waitForNavigation({ waitUntil: WAIT_EVENT_NETWORK }),
      page.waitForNavigation({ waitUntil: WAIT_EVENT_DOM })
    ]).catch(function(e) {
      console.log(`ERROR (${name}): ${e.message}`);
      module.exports.takeScreenShot(name.replace(/\s/g, ''));
      throw e;
    });
  },
  login: async (page, isOIDC, uri, token, username, password) => {
    let doAction = module.exports.doAction;
    await doAction("Go to Home", page.goto(getUrl(uri)));
    if (isOIDC === "true") {
      console.log("Log in using OIDC")
      await doAction("Click to start login", page.click(".login-submit-button"));

      // DEX: Choose email as login method
      page.waitForSelector('.dex-container button');
      await expect(page).toMatchElement(".dex-container button", {
        text: "Log in with Email",
      });
      expect(page).toClick(".dex-container button", {
        text: "Log in with Email",
      });

      await page.waitForNavigation({ waitUntil: WAIT_EVENT_NETWORK });
      await page.waitForSelector("#submit-login", {
        visible: true,
        timeout: 10000,
      });
      await page.type('input[id="login"]', username);
      await page.type('input[id="password"]', password);
      await doAction("Click submit user and password", page.click("#submit-login"));

      // DEX: click on the new "Grant Access" confirmation.
      await page.waitForSelector('.dex-container button[type="submit"]', {
        text: "Grant Access",
        visible: true,
        timeout: 10000,
      });
      await expect(page).toMatchElement('.dex-container button[type="submit"]', {
        text: "Grant Access",
      });
      await doAction("Click submit Grant Access in DEX", page.click('.dex-container button[type="submit"]'));

      // Navigation back in Kubeapps
      await page.waitForSelector(".kubeapps-header-content", {
        visible: true,
        timeout: 10000,
      });
      if (uri !== "/") {
        await doAction("Go back to Home", page.goto(getUrl(uri)));
      }
    } else {
      await expect(page).toFillForm("form", {
        token: token,
      });
      await page.waitForSelector("#login-submit-button", {
        visible: true,
        timeout: 10000,
      });
      await page.click("#login-submit-button");
    }
  },
  getRandomName: base => {
    const randomNumber = Math.floor(Math.random() * Math.floor(10000));
    const name = base + "-" + randomNumber;
    return name;
  },
};
