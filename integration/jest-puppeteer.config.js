const { headless } = require("./args");

module.exports = {
  launch: {
    headless,
    args: ["--no-sandbox"]
  }
};
