// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

const { headless } = require("./args");

module.exports = {
  launch: {
    headless,
    args: ["--no-sandbox", "--window-size=1200,780", "--ignore-certificate-errors"],
  },
  browserContext: "incognito",
};
