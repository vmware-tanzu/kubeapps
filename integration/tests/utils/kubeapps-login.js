// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

const utils = require("./util-functions");

exports.KubeappsLogin = class KubeappsLogin {
  constructor(page) {
    this.page = page;
    this.oidc = process.env.USE_MULTICLUSTER_OIDC_ENV === "true";
    this.token = "";
  }

  isOidc = () => this.oidc;

  setToken(token) {
    let bearerMatch = token.match(/Bearer/g);
    if (bearerMatch && bearerMatch.length > 1) {
      token = token.split(",")[0];
    }
    this.token = token.trim().startsWith("Bearer") ? token.trim().substring(7) : token;
  }

  async doLogin(username, pwd, token) {
    if (this.oidc) {
      this.page.on("response", response => {
        // retrieves the token after the oidc flow, note this require "--set-authorization-header=true" flag to be enabled in oauth2proxy
        let oidcToken = response.headers()["authorization"];
        if (oidcToken) this.setToken(oidcToken);
      });
      await this.doOidcLogin(username, pwd);
    } else if (token) {
      this.setToken(token);
      await this.doTokenLogin(token);
    } else {
      console.log("ERROR: No valid login data provided.");
      return;
    }
    console.log("Logged into Kubeapps!");
  }

  async doLogout() {
    console.log("Logging out of Kubeapps");
    await this.page.click(".dropdown.kubeapps-menu .kubeapps-nav-link");
    await this.page.click('cds-button:has-text("Log out")');
    await this.page.waitForLoadState("networkidle");
    console.log("Logged out of Kubeapps");
  }

  async doOidcLogin(username, pwd) {
    console.log(`Logging in Kubeapps via OIDC in host: ${utils.getUrl("/")}`);

    // Go to Home page
    await this.page.goto(utils.getUrl("/"));
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

    // Confirm Grant Access
    await this.page.locator('button.dex-btn:has-text("Grant Access")').click();
    await this.page.waitForLoadState("networkidle");
    await this.page.waitForLoadState("domcontentloaded");
  }

  async doTokenLogin(token) {
    console.log(`Logging in Kubeapps using token in host: ${utils.getUrl("/")}`);

    // Go to Home page
    await this.page.goto(utils.getUrl("/"));
    await this.page.waitForLoadState("networkidle");

    const inputLocator = this.page.locator("form input[name=token]");
    await inputLocator.fill(token);
    await this.page.click("#login-submit-button");

    await this.page.waitForLoadState("networkidle");
    await this.page.waitForLoadState("domcontentloaded");
  }
};
