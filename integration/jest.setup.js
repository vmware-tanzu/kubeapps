const { endpoint } = require("./args");
const { setDefaultOptions } = require("expect-puppeteer");

// Change timeout for Puppeteer page.waitForXXX functions from 0.5s to 30s
setDefaultOptions({ timeout: 30000 });

// endpoint argument is mandatory
if (endpoint == null || endpoint == "") {
  console.error("The INTEGRATION_ENDPOINT environment variable is mandatory");
  process.exit(1);
}

// Initialize globals
global.endpoint = endpoint;

// Helper to get the proper endpoint
global.getUrl = path => `${global.endpoint}${path}`;

// Timeout for a test
jest.setTimeout(160000);
