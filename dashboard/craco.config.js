const webpack = require("webpack");
const NodePolyfillPlugin = require('node-polyfill-webpack-plugin');

module.exports = {
  webpack: {
    configure: {
      plugins: [
        new NodePolyfillPlugin(), // add the required polyfills (not included in webpack 5)
        new webpack.ProvidePlugin({
          process: 'process/browser.js',
          Buffer: ['buffer', 'Buffer'],
        }),
      ],
      ignoreWarnings: [/Failed to parse source map/], // ignore source map warnings
    },
  },
};
