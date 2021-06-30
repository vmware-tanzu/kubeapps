/* eslint-disable */
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
        // Temporary workaround until one of these issues get solved:
        // https://github.com/babel/babel/issues/13520
        // https://github.com/facebook/create-react-app/issues/11159
        // Note:
        // Flow is a static type checker, but we don't have any raw js files requiring it.
        // It is enabled by default, so we need to explicitly disable it because it is
        // causing some parsing errors when reading named imports like "import foo as as".
        // More context here: https://github.com/kubeapps/kubeapps/pull/3042#issuecomment-870053634
        flow: false,
      },
    ],
  ],
  babelrc: false,
  configFile: false,
});
