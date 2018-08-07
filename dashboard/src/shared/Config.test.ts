import axios from "axios";
import MockAdapter from "axios-mock-adapter";
import Config from "./Config";

describe("Config", () => {
  let defaultJSON: any;
  let initialEnv: string;

  beforeAll(() => {
    initialEnv = process.env.REACT_APP_KUBEAPPS_NS || "";
    const mock = new MockAdapter(axios);

    defaultJSON = require("../../public/config.json");

    mock.onGet("/config.json").reply(200, defaultJSON);
  });

  afterEach(() => {
    process.env.REACT_APP_KUBEAPPS_NS = initialEnv;
  });

  it("returns default namespace if no override proced", async () => {
    expect(await Config.getConfig()).toEqual(defaultJSON);
  });

  it("returns the overriden namespace if env variable provided", async () => {
    process.env.REACT_APP_KUBEAPPS_NS = "magic-playground";
    expect(await Config.getConfig()).toEqual({ ...defaultJSON, namespace: "magic-playground" });
  });
});
