require("expect-puppeteer");
const { endpoint } = require("./args");

// endpoint argument is mandatory
if (endpoint == null || endpoint == "") {
  console.error("The INTEGRATION_ENDPOINT environment variable is mandatory");
  process.exit(1);
}

// Initialize globals
global.endpoint = endpoint;

// Helper to get the proper endpoint
global.getUrl = path => `${global.endpoint}${path}`;
