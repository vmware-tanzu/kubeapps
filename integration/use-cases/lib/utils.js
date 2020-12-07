const {
  screenshotsFolder,
} = require("../../args");
const path = require("path");

module.exports = {
  retryAndRefresh: async (page, retries, toCheck,  testName) => {
    let retriesLeft = retries;
    while (retriesLeft > 0) {
      try {
        await toCheck();
        break;
      } catch (e) {
        if(testName) {
          await page.screenshot({
            path: path.join(__dirname, `../../${screenshotsFolder}/${testName}-${retries - retriesLeft}.png`),
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
  click: async (page, document, selector) => {
    page.evaluate(() => document.querySelector(selector).click());
  },
};
