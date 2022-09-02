// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

const webpack = require("webpack");
const NodePolyfillPlugin = require('node-polyfill-webpack-plugin');

module.exports = {
  webpack: {
    configure: {
      plugins: [
        // add the required polyfills (not included in webpack 5)
        new NodePolyfillPlugin({
          // Allow using console.log
          excludeAliases: ['console'],
        }),
        new webpack.ProvidePlugin({
          process: 'process/browser.js',
          Buffer: ['buffer', 'Buffer'],
        }),
      ],
      ignoreWarnings: [/Failed to parse source map/], // ignore source map warnings
    },
  },
};
