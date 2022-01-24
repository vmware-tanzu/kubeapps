// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import axios from "axios";
import * as moxios from "moxios";
import Config, { IConfig, SupportedThemes } from "./Config";

describe("Config", () => {
  let defaultJSON: IConfig;
  let initialEnv: any;

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
    jest.restoreAllMocks();
  });

  it("returns default namespace if no override provided", async () => {
    expect(await Config.getConfig()).toEqual(defaultJSON);
  });
});

describe("Themes", () => {
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
    jest.restoreAllMocks();
  });

  it("returns the stored theme", () => {
    jest.spyOn(window.localStorage.__proto__, "getItem").mockReturnValue(SupportedThemes.dark);
    expect(Config.getTheme({} as IConfig)).toBe(SupportedThemes.dark);
  });

  it("returns the light theme by default", () => {
    expect(Config.getTheme({} as IConfig)).toBe(SupportedThemes.light);
  });

  it("returns the system theme", () => {
    const defaultJSONTheme = { ...defaultJSON, theme: SupportedThemes.dark };
    expect(Config.getTheme(defaultJSONTheme)).toBe(SupportedThemes.dark);
  });

  it("returns the default browser theme", () => {
    Object.defineProperty(window, "matchMedia", {
      writable: true,
      value: jest.fn().mockImplementation(() => ({
        matches: true,
      })),
    });
    expect(Config.getTheme({} as IConfig)).toBe(SupportedThemes.dark);
  });

  it("returns the theme according to the preference (user>system>browser>fallback)", () => {
    // User preference = dark
    jest.spyOn(window.localStorage.__proto__, "getItem").mockReturnValue(SupportedThemes.dark);

    // System preference = light
    const defaultJSONTheme = { ...defaultJSON, theme: SupportedThemes.light };

    // Browser preference = dark
    Object.defineProperty(window, "matchMedia", {
      writable: true,
      value: jest.fn().mockImplementation(() => ({
        matches: true,
      })),
    });
    expect(Config.getTheme(defaultJSONTheme)).toBe(SupportedThemes.dark);
  });

  it("returns the theme according to the preference (system>browser>fallback)", () => {
    // System preference = light
    const defaultJSONTheme = { ...defaultJSON, theme: SupportedThemes.light };

    // Browser preference = dark
    Object.defineProperty(window, "matchMedia", {
      writable: true,
      value: jest.fn().mockImplementation(() => ({
        matches: true,
      })),
    });
    expect(Config.getTheme(defaultJSONTheme)).toBe(SupportedThemes.light);
  });

  it("returns the theme according to the preference (browser>fallback)", () => {
    // System preference = N/A
    const defaultJSONTheme = { ...defaultJSON, theme: "" };

    // Browser preference = dark
    Object.defineProperty(window, "matchMedia", {
      writable: true,
      value: jest.fn().mockImplementation(() => ({
        matches: true,
      })),
    });
    expect(Config.getTheme(defaultJSONTheme)).toBe(SupportedThemes.dark);
  });

  it("returns the theme according to the preference (fallback)", () => {
    // System preference = N/A
    const defaultJSONTheme = { ...defaultJSON, theme: "" };

    expect(Config.getTheme(defaultJSONTheme)).toBe(SupportedThemes.light);
  });
});
