import * as moxios from "moxios";
import { App } from "./App";
import { axios } from "./Auth";
import { IAppOverview } from "./types";

describe("App", () => {
  beforeEach(() => {
    moxios.install(axios);
  });
  afterEach(() => {
    moxios.uninstall(axios);
  });
  describe("getResourceURL", () => {
    it("returns the root API URL if no params are given", () => {
      expect(App.getResourceURL()).toBe("/api/tiller-deploy/v1/releases");
    });
    it("returns namespaced URLs", () => {
      expect(App.getResourceURL("default")).toBe(
        "/api/tiller-deploy/v1/namespaces/default/releases",
      );
    });
    it("returns a single release URL", () => {
      expect(App.getResourceURL("default", "foo")).toBe(
        "/api/tiller-deploy/v1/namespaces/default/releases/foo",
      );
    });
    it("returns a URL with a query", () => {
      expect(App.getResourceURL("default", undefined, "statuses=foo")).toBe(
        "/api/tiller-deploy/v1/namespaces/default/releases?statuses=foo",
      );
    });
  });

  describe("listApps", () => {
    it("should request all the releases if no namespace is given", async () => {
      const apps = [{ releaseName: "foo" } as IAppOverview];
      moxios.stubRequest("/api/tiller-deploy/v1/releases", {
        response: { data: apps },
        status: 200,
      });
      expect(await App.listApps()).toEqual(apps);
    });
    it("should request the releases of a namespace", async () => {
      const apps = [{ releaseName: "foo" } as IAppOverview];
      moxios.stubRequest("/api/tiller-deploy/v1/namespaces/default/releases", {
        response: { data: apps },
        status: 200,
      });
      expect(await App.listApps("default")).toEqual(apps);
    });
    it("should request the releases of a namespace with any status", async () => {
      const apps = [{ releaseName: "foo" } as IAppOverview];
      moxios.stubRequest("/api/tiller-deploy/v1/namespaces/default/releases?statuses=all", {
        response: { data: apps },
        status: 200,
      });
      expect(await App.listApps("default", true)).toEqual(apps);
    });
  });
});
