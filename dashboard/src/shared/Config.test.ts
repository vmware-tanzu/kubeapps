import axios from "axios";
import * as moxios from "moxios";
import Config, { IConfig, SupportedThemes } from "./Config";

describe("Config", () => {
  let defaultJSON: IConfig;
  let initialEnv: any;

  const matchMedia = window.matchMedia;
  beforeEach(() => {
    initialEnv = { ...process.env };
    // Import as "any" to avoid typescript syntax error
    moxios.install(axios as any);

    defaultJSON = require("../../public/config.json");

    moxios.stubRequest("config.json", { status: 200, response: { ...defaultJSON } });
  });

  afterEach(() => {
    process.env = initialEnv;
    moxios.uninstall(axios as any);
    window.matchMedia = matchMedia;
  });

  it("returns default namespace if no override provided", async () => {
    expect(await Config.getConfig()).toEqual(defaultJSON);
  });

  it("returns the stored theme", () => {
    jest.spyOn(window.localStorage.__proto__, "getItem").mockReturnValueOnce(SupportedThemes.dark);
    expect(Config.getTheme()).toBe(SupportedThemes.dark);
  });

  it("returns the light theme by default", () => {
    expect(Config.getTheme()).toBe(SupportedThemes.light);
  });

  it("returns the default browser theme", () => {
    Object.defineProperty(window, "matchMedia", {
      writable: true,
      value: jest.fn().mockImplementation(query => ({
        matches: true,
      })),
    });
    expect(Config.getTheme()).toBe(SupportedThemes.dark);
  });
});
