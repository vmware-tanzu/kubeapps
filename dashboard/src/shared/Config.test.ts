import axios from "axios";
import MockAdapter from "axios-mock-adapter";
import Config, { IConfig } from "./Config";

describe("Config", () => {
  let defaultJSON: IConfig;
  let initialEnv: any;

  beforeEach(() => {
    initialEnv = { ...process.env };
    const mock = new MockAdapter(axios);

    defaultJSON = require("../../public/config.json");

    mock.onGet("/config.json").reply(200, defaultJSON);
  });

  afterEach(() => {
    process.env = initialEnv;
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
