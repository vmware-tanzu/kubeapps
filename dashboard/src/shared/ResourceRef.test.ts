import { Kube } from "./Kube";
import ResourceRef from "./ResourceRef";
import { IResource } from "./types";

describe("ResourceRef", () => {
  describe("constructor", () => {
    it("it returns a ResourceRef with the correct details", () => {
      const r = {
        apiVersion: "apps/v1",
        kind: "Deployment",
        metadata: {
          name: "foo",
          namespace: "bar",
        },
      } as IResource;

      const ref = new ResourceRef(r);
      expect(ref).toBeInstanceOf(ResourceRef);
      expect(ref).toEqual({
        apiVersion: r.apiVersion,
        kind: r.kind,
        name: r.metadata.name,
        namespace: r.metadata.namespace,
      });
    });

    it("sets a default namespace if not in the resource", () => {
      const r = {
        apiVersion: "apps/v1",
        kind: "Deployment",
        metadata: {
          name: "foo",
        },
      } as IResource;

      const ref = new ResourceRef(r, "default");
      expect(ref.namespace).toBe("default");
    });

    it("throws an error if namespace not in the resource or default namespace not set", () => {
      const r = {
        apiVersion: "apps/v1",
        kind: "Deployment",
        metadata: {
          name: "foo",
        },
      } as IResource;

      expect(() => new ResourceRef(r)).toThrowError();
    });

    it("allows the default namespace to be provided", () => {
      const r = {
        apiVersion: "apps/v1",
        kind: "Deployment",
        metadata: {
          name: "foo",
        },
      } as IResource;

      const ref = new ResourceRef(r, "bar");
      expect(ref.namespace).toBe("bar");
    });
  });

  describe("getResourceURL", () => {
    let kubeGetResourceURLMock: jest.Mock;
    beforeEach(() => {
      kubeGetResourceURLMock = jest.fn();
      Kube.getResourceURL = kubeGetResourceURLMock;
    });
    afterEach(() => {
      jest.resetAllMocks();
    });
    it("calls Kube.getResourceURL with the correct arguments", () => {
      const r = {
        apiVersion: "v1",
        kind: "Service",
        metadata: {
          name: "foo",
          namespace: "bar",
        },
      } as IResource;

      const ref = new ResourceRef(r);

      ref.getResourceURL();
      expect(kubeGetResourceURLMock).toBeCalledWith("v1", "services", "bar", "foo");
    });
  });

  describe("watchResourceURL", () => {
    let kubeWatchResourceURLMock: jest.Mock;
    beforeEach(() => {
      kubeWatchResourceURLMock = jest.fn();
      Kube.watchResourceURL = kubeWatchResourceURLMock;
    });
    afterEach(() => {
      jest.resetAllMocks();
    });
    it("calls Kube.watchResourceURL with the correct arguments", () => {
      const r = {
        apiVersion: "v1",
        kind: "Service",
        metadata: {
          name: "foo",
          namespace: "bar",
        },
      } as IResource;

      const ref = new ResourceRef(r);

      ref.watchResourceURL();
      expect(kubeWatchResourceURLMock).toBeCalledWith("v1", "services", "bar", "foo");
    });
  });

  describe("getResource", () => {
    let kubeGetResourceMock: jest.Mock;
    beforeEach(() => {
      kubeGetResourceMock = jest.fn();
      Kube.getResource = kubeGetResourceMock;
    });
    afterEach(() => {
      jest.resetAllMocks();
    });
    it("calls Kube.getResource with the correct arguments", () => {
      const r = {
        apiVersion: "v1",
        kind: "Service",
        metadata: {
          name: "foo",
          namespace: "bar",
        },
      } as IResource;

      const ref = new ResourceRef(r);

      ref.getResource();
      expect(kubeGetResourceMock).toBeCalledWith("v1", "services", "bar", "foo");
    });
  });

  describe("watchResource", () => {
    let kubeWatchResourceMock: jest.Mock;
    beforeEach(() => {
      kubeWatchResourceMock = jest.fn();
      Kube.watchResource = kubeWatchResourceMock;
    });
    afterEach(() => {
      jest.resetAllMocks();
    });
    it("calls Kube.watchResource with the correct arguments", () => {
      const r = {
        apiVersion: "v1",
        kind: "Service",
        metadata: {
          name: "foo",
          namespace: "bar",
        },
      } as IResource;

      const ref = new ResourceRef(r);

      ref.watchResource();
      expect(kubeWatchResourceMock).toBeCalledWith("v1", "services", "bar", "foo");
    });
  });
});
