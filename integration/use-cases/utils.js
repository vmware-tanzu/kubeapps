module.exports = {
  retryAndRefresh: async (page, retries, toCheck) => {
    let retriesLeft = retries;
    while (retriesLeft > 0) {
      try {
        await toCheck();
        break;
      } catch (e) {
        // Refresh since the chart will get a bit of time to populate
        await page.reload({
          waitUntil: ["networkidle0", "domcontentloaded"],
          timeout: 20000,
        });
      } finally {
        retriesLeft--;
        console.log("retrying");
      }
    }
  },
};
