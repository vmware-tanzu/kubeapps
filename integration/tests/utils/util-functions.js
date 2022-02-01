// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

const path = require("path");
const fs = require("fs");
const axios = require("axios");
const https = require('https');

module.exports = {

  getRandomName: base => {
    const randomNumber = Math.floor(Math.random() * Math.floor(100000));
    return base + "-" + randomNumber;
  },

  getUrl: path => `${process.env.INTEGRATION_ENTRYPOINT}${path}`,

  goTo: async (page, url) => {
    try {
      console.log(">>> Navigating to: " + url);
      await page.goto(url, { timeout: 5000, waitUntil: "networkidle" });
    } catch (err) {
      console.log(`>>> Error navigating to ${url}. Retrying...`);
      await page.goto(url, { waitUntil: "networkidle" });
    }
  },

  getAxiosInstance: async (page) => {
    const cookies = await page.context().cookies(page.url());
    const agent = new https.Agent({  
      rejectUnauthorized: false
    });
    const axiosConfig = {
      baseURL: `${process.env.INTEGRATION_ENTRYPOINT}`,
      headers: {
        Cookie: `${cookies[0] ? cookies[0].name : ""}=${cookies[0] ? cookies[0].value : ""}`,
        Accept: "application/json"
      },
      httpsAgent: agent,
      timeout: 30000
    };
    return await axios.create(axiosConfig);
  }
};
