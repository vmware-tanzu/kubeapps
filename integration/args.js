// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

module.exports = {
  // Endpoint is required!
  endpoint: process.env.INTEGRATION_ENTRYPOINT,
  waitTimeout: process.env.INTEGRATION_WAIT_TIMEOUT || 60000,
  headless: process.env.INTEGRATION_HEADLESS != "false",
  retryAttempts: process.env.INTEGRATION_RETRY_ATTEMPTS || 0,
  screenshotsFolder: process.env.INTEGRATION_SCREENSHOTS_FOLDER || "reports/screenshots",
};
