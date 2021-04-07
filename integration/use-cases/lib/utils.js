const { screenshotsFolder } = require("../../args");
const path = require("path");

module.exports = {
  retryAndRefresh: async (page, retries, toCheck, testName) => {
    let retriesLeft = retries;
    while (retriesLeft > 0) {
      try {
        await toCheck();
        break;
      } catch (e) {
        if (testName) {
          await page.screenshot({
            path: path.join(
              __dirname,
              `../../${screenshotsFolder}/${testName}-${retries - retriesLeft}.png`,
            ),
          });
        }
        if (retriesLeft === 1) {
          // Unable to get it done
          throw e;
        }
        // Refresh since the chart will get a bit of time to populate
        try {
          await page.reload({
            waitUntil: ["domcontentloaded"],
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
  login: async (page, isOIDC, uri, token, username, password) => {
    await page.goto(getUrl(uri));
    if (isOIDC === "true") {
      await page.waitForNavigation();
      await expect(page).toClick("cds-button", {
        text: "Login via OIDC Provider",
      });
      await page.waitForNavigation({ waitUntil: "domcontentloaded" });
      await expect(page).toMatchElement(".dex-container button", {
        text: "Log in with Email",
      });
      await expect(page).toClick(".dex-container button", {
        text: "Log in with Email",
      });
      await page.waitForNavigation();
      await page.type('input[id="login"]', username);
      await page.type('input[id="password"]', password);
      await page.waitForSelector("#submit-login", {
        visible: true,
        timeout: 10000,
      });
      await page.click("#submit-login");
      await page.waitForSelector(".kubeapps-header-content", {
        visible: true,
        timeout: 10000,
      });
      if (uri !== "/") {
        await page.goto(getUrl(uri));
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
};
