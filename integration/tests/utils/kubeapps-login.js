const { TestUtils } = require("./util-functions");

exports.KubeappsOidcLogin = class KubeappsOidcLogin {
  constructor(page) {
    this.page = page;
    this.utils = new TestUtils();
  }

  getUrl = path => this.utils.getUrl(path);

  async doOidcLogin(username, pwd) {
    console.log(`Logging in Kubeapps via OIDC in host: ${this.utils.getUrl("/")}`);

    // Go to Home page
    await this.page.goto(this.utils.getUrl("/"));
    await this.page.waitForLoadState("networkidle");

    // Click to Log in with OIDC provider
    await this.page.click("text=Login via OIDC Provider");
    await this.page.waitForLoadState("networkidle");

    // Click to Log in with Email
    await this.page.click("text=Log in with Email");
    await this.page.waitForLoadState("networkidle");

    // Type in credentials
    await this.page.fill('input[id="login"]', username);
    await this.page.fill('input[id="password"]', pwd);
    await this.page.click("#submit-login");
    await this.page.waitForLoadState("networkidle");
    await this.page.waitForLoadState("domcontentloaded");

    await this.page.screenshot({ path: "screenshots/login-submit-grant.png" });

    // Confirm Grant Access
    await this.page.locator('button.dex-btn:has-text("Grant Access")').click();
    await this.page.waitForLoadState("networkidle");
    await this.page.waitForLoadState("domcontentloaded");
  }

  async doTokenLogin(token) {
    // Go to Home page
    await this.page.goto(this.getUrl("/"));
    await this.page.waitForLoadState("networkidle");

    const formLocator = page.locator("form");
    await formLocator.type("input[name=token]", token);
    await this.page.click("#login-submit-button");

    await this.page.waitForLoadState("networkidle");
    await this.page.waitForLoadState("domcontentloaded");
  }
};
