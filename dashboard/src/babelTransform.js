"use strict";

const babelJest = require("babel-jest");

const hasJsxRuntime = (() => {
  if (process.env.DISABLE_NEW_JSX_TRANSFORM === "true") {
    return false;
  }

  try {
    require.resolve("react/jsx-runtime");
    return true;
  } catch (e) {
    return false;
  }
})();

module.exports = babelJest.createTransformer({
  presets: [
    [
      require.resolve("babel-preset-react-app"),
      {
        runtime: hasJsxRuntime ? "automatic" : "classic",
        // Temporary workaround until this issue get solved:
        // https://github.com/babel/babel/issues/13520
        flow: false,
      },
    ],
  ],
  babelrc: false,
  configFile: false,
});
