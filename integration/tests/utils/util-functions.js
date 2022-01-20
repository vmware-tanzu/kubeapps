exports.TestUtils = class TestUtils {
  constructor(page) {
    this.host = process.env.INTEGRATION_ENTRYPOINT;
  }

  static async waitFor(page) {
    await page.waitForLoadState("networkidle");
    await page.waitForLoadState("domcontentloaded");
  }

  static getRandomName(base) {
    const randomNumber = Math.floor(Math.random() * Math.floor(100000));
    return base + "-" + randomNumber;
  }

  getUrl = path => `${this.host}${path}`;
};
