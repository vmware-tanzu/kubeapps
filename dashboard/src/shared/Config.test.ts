import axios from "axios";
import * as moxios from "moxios";
import Config, { IConfig } from "./Config";

describe("Config", () => {
  let defaultJSON: IConfig;
  let initialEnv: any;

  beforeEach(() => {
    initialEnv = { ...process.env };
    moxios.install(axios);

    defaultJSON = require("../../public/config.json");

    moxios.stubRequest("config.json", { status: 200, response: defaultJSON });
  });

  afterEach(() => {
    process.env = initialEnv;
    moxios.uninstall(axios);
  });

  it("returns default namespace if no override provided", async () => {
    expect(await Config.getConfig()).toEqual(defaultJSON);
  });

  it("returns the overriden namespace if env variable provided", async () => {
    process.env.REACT_APP_KUBEAPPS_NS = "magic-playground";
    expect(await Config.getConfig()).toEqual({ ...defaultJSON, namespace: "magic-playground" });
  });

  it("does not returns the overriden namespace if NODE_ENV=production", async () => {
    process.env.NODE_ENV = "production";
    process.env.REACT_APP_KUBEAPPS_NS = "magic-playground";
    expect(await Config.getConfig()).toEqual(defaultJSON);
  });
});
