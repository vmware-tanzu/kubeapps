// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// global-setup.js
const utils = require("./tests/utils/util-functions");

module.exports = async config => {
  const timeDisplayFactor = 60000;
  console.log("Setting up Playwright tests");
  console.log(`>> Global timeout: ${config.globalTimeout / timeDisplayFactor} mins`);
  config.projects.forEach(project => {
    console.log(
      `>> Project ${project.name} test timeout: ${project.timeout / timeDisplayFactor} mins`,
    );
  });
  console.log(`>> Deployments timeout: ${utils.getDeploymentTimeout() / timeDisplayFactor} mins`);
};
