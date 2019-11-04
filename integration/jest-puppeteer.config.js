const { headless } = require("./args");

module.exports = {
  launch: {
    headless,
    args: ["--no-sandbox", "--window-size=1200,780"]
  },
  browserContext: "incognito"
};
