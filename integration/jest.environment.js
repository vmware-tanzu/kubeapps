const path = require("path");
const fs = require("fs");
const waitOn = require("wait-on");
const PuppeteerEnvironment = require("jest-environment-puppeteer");
require("jest-circus");

const { retryAttempts, endpoint, waitTimeout, screenshotsFolder } = require("./args");

// Create an environment to store a screenshot of the page if the current test
// failed.
class ScreenshotOnFailureEnvironment extends PuppeteerEnvironment {
  async generateScreenshotsFolder() {
    try {
      // Create the report folder if it's not there
      if (!fs.existsSync(screenshotsFolder)) {
        await fs.promises.mkdir(screenshotsFolder, { recursive: true });
      }
    } catch (err) {
      console.error(`The ${screenshotsFolder} folder couldn't be created`);
      process.exit(1);
    }
  }

  async waitOnService() {
    try {
      // Check the server is up before running the test suite
      console.log(
        `Waiting ${endpoint} to be ready before running the tests
        (${waitTimeout / 1000}s)`,
      );
      await waitOn({
        resources: [endpoint],
        timeout: waitTimeout,
      });
      console.log(`${endpoint} is ready!`);
    } catch (err) {
      console.error(`The ${endpoint} URL is not accessible due to:`);
      console.error(err);
      process.exit(1);
    }
  }

  async setup() {
    await this.generateScreenshotsFolder();
    await this.waitOnService();
    await super.setup();
    await this.global.page.setViewport({
      width: 1200,
      height: 780,
      deviceScaleFactor: 1,
    });
    await this.global.page.setDefaultTimeout(30000); // 30s is the default value
    this.global.page
      .on("console", async message => {
        if (message.type() === "error") {
          console.error(`${message.type().toUpperCase()} ${message.text()}`);
          // Code from https://github.com/puppeteer/puppeteer/issues/3397#issuecomment-429325514
          const args = await message.args();
          args.forEach(async arg => {
            const val = await arg.jsonValue();
            // value is serializable
            if (JSON.stringify(val) !== JSON.stringify({})) console.log(val);
            // value is unserializable (or an empty oject)
            else {
              const { type, subtype, description } = arg._remoteObject;
              console.log(`type: ${type}, subtype: ${subtype}, description:\n ${description}`);
            }
          });
        }
      })
      .on("pageerror", ({ message }) => console.log(message))
      .on("requestfailed", request =>
        console.log(`${request.failure().errorText} ${request.url()}`),
      )
      .on("response", response => {
        if (response.status() >= 400) {
          console.log(`${response.status()} in ${response.url()}`);
        }
      });
  }

  async teardown() {
    // Wait a few seconds before tearing down the page so we
    // have time to take screenshots and handle other events
    await this.global.page.waitForTimeout(2000);
    await super.teardown();
  }

  async handleTestEvent(event, state) {
    if (event.name == "test_fn_failure") {
      if (state.currentlyRunningTest.invocations > retryAttempts) {
        const testName = state.currentlyRunningTest.name.toLowerCase().replace(/ /g, "-");
        // Take a screenshot at the point of failure
        await this.global.page.screenshot({
          path: path.join(__dirname, `${screenshotsFolder}/${testName}.png`),
        });
      }
    }
  }
}

module.exports = ScreenshotOnFailureEnvironment;
