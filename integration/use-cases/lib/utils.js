module.exports = {
  retryAndRefresh: async (page, retries, toCheck) => {
    let retriesLeft = retries;
    while (retriesLeft > 0) {
      try {
        await toCheck();
        break;
      } catch (e) {
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
};
