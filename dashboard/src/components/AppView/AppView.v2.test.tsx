import { safeDump as yamlSafeDump } from "js-yaml";
import * as React from "react";

import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper.v2";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { IResource } from "shared/types";
import ApplicationStatusContainer from "../../containers/ApplicationStatusContainer";
import { hapi } from "../../shared/hapi/release";
import ResourceRef from "../../shared/ResourceRef";
import AccessURLTable from "./AccessURLTable/AccessURLTable.v2";
import AppNotes from "./AppNotes.v2";
import AppViewComponent from "./AppView.v2";
import ChartInfo from "./ChartInfo/ChartInfo.v2";
import ResourceTabs from "./ResourceTabs";

describe("AppViewComponent", () => {
  // Generates a Yaml file separated by --- containing every object passed.
  const generateYamlManifest = (items: any[]): string => {
    let yamlManifest = "";
    items.forEach(i => {
      yamlManifest += "---\n" + yamlSafeDump(i);
    });
    return yamlManifest;
  };

  const appRelease = hapi.release.Release.create({
    info: hapi.release.Info.create(),
    namespace: "weee",
  });

  const validProps = {
    app: appRelease,
    deleteApp: jest.fn(),
    deleteError: undefined,
    error: undefined,
    getAppWithUpdateInfo: jest.fn(),
    namespace: "my-happy-place",
    cluster: "default",
    releaseName: "mr-sunshine",
    push: jest.fn(),
  };

  const resources = {
    configMap: { apiVersion: "v1", kind: "ConfigMap", metadata: { name: "cm-one" } },
    deployment: {
      apiVersion: "apps/v1beta1",
      kind: "Deployment",
      metadata: { name: "deployment-one" },
    } as IResource,
    service: { apiVersion: "v1", kind: "Service", metadata: { name: "svc-one" } } as IResource,
    ingress: {
      apiVersion: "extensions/v1beta1",
      kind: "Ingress",
      metadata: { name: "ingress-one" },
    } as IResource,
    secret: {
      apiVersion: "v1",
      kind: "Secret",
      metadata: { name: "secret-one" },
      type: "Opaque",
    } as IResource,
    daemonset: {
      apiVersion: "apps/v1beta1",
      kind: "DaemonSet",
      metadata: { name: "daemonset-one" },
    } as IResource,
    statefulset: {
      apiVersion: "apps/v1beta1",
      kind: "StatefulSet",
      metadata: { name: "statefulset-one" },
    } as IResource,
  };

  it("renders a loading wrapper", () => {
    const props = {
      ...validProps,
      app: { ...validProps.app, info: null },
    } as any;

    const wrapper = mountWrapper(defaultStore, <AppViewComponent {...props} />);
    expect(wrapper.find(LoadingWrapper)).toExist();
  });

  describe("State initialization", () => {
    /*
      The imported manifest contains one deployment, one service, one config map and some bogus manifests.
    */
    it("sets ResourceRefs for its deployments, services, ingresses and secrets", () => {
      const manifest = generateYamlManifest([
        resources.deployment,
        resources.service,
        resources.configMap,
        resources.ingress,
        resources.secret,
      ]);

      const props = {
        ...validProps,
        app: {
          ...validProps.app,
          manifest,
        } as any,
      };
      const wrapper = mountWrapper(defaultStore, <AppViewComponent {...props} />);

      const tabs = wrapper.find(ResourceTabs);
      expect(tabs.prop("deployments")).toEqual([
        new ResourceRef(resources.deployment, validProps.cluster, appRelease.namespace),
      ]);
      expect(tabs.prop("services")).toEqual([
        new ResourceRef(resources.service, validProps.cluster, appRelease.namespace),
      ]);
      expect(tabs.prop("secrets")).toEqual([
        new ResourceRef(resources.secret, validProps.cluster, appRelease.namespace),
      ]);
    });

    it("stores other k8s resources", () => {
      const manifest = generateYamlManifest([
        resources.deployment,
        resources.service,
        resources.configMap,
        resources.secret,
      ]);

      const props = {
        ...validProps,
        app: {
          ...validProps.app,
          manifest,
        } as any,
      };
      const wrapper = mountWrapper(defaultStore, <AppViewComponent {...props} />);

      const tabs = wrapper.find(ResourceTabs);
      const otherResources: ResourceRef[] = tabs.prop("otherResources");
      const configMap = otherResources[0];
      // It should skip deployments, services and secrets from "other resources"
      expect(otherResources.length).toEqual(1);

      expect(configMap).toBeDefined();
      expect(configMap.name).toEqual("cm-one");
    });

    it("does not store empty resources, bogus or without kind attribute", () => {
      const manifest = generateYamlManifest([
        { apiVersion: "v1", metadata: { name: "cm-one" } },
        {},
        "# This is a comment",
        " ",
      ]);

      const props = {
        ...validProps,
        app: {
          ...validProps.app,
          manifest,
        } as any,
      };
      const wrapper = mountWrapper(defaultStore, <AppViewComponent {...props} />);

      const tabs = wrapper.find(ResourceTabs);
      expect(tabs.prop("otherResources")).toEqual([]);
      expect(tabs.prop("deployments")).toEqual([]);
      expect(tabs.prop("services")).toEqual([]);
      expect(tabs.prop("secrets")).toEqual([]);
    });

    // See https://github.com/kubeapps/kubeapps/issues/632
    it("supports manifests with duplicated keys", () => {
      const manifest = `
      apiVersion: v1
      metadata:
        name: cm-one
        labels:
          chart: cm-1.2.3
          chart: cm-1.2.3
`;

      const props = {
        ...validProps,
        app: {
          ...validProps.app,
          manifest,
        } as any,
      };
      expect(() => {
        mountWrapper(defaultStore, <AppViewComponent {...props} />);
      }).not.toThrow();
    });

    it("supports manifests with YAML type casting", () => {
      const manifest = `
      apiVersion: v1
      kind: Deployment
      metadata:
        name: !!string foo
`;

      const props = {
        ...validProps,
        app: {
          ...validProps.app,
          manifest,
        } as any,
      };
      expect(() => {
        const wrapper = mountWrapper(defaultStore, <AppViewComponent {...props} />);
        const tabs = wrapper.find(ResourceTabs);
        expect(tabs.prop("deployments")[0].name).toEqual("foo");
      }).not.toThrow();
    });
  });

  describe("renderization", () => {
    it("renders all the elements of an application", () => {
      const wrapper = mountWrapper(defaultStore, <AppViewComponent {...validProps} />);
      expect(wrapper.find(ChartInfo)).toExist();
      expect(wrapper.find(ApplicationStatusContainer)).toExist();
      expect(wrapper.find(".control-buttons")).toExist();
      expect(wrapper.find(AppNotes)).toExist();
      expect(wrapper.find(ResourceTabs)).toExist();
      expect(wrapper.find(AccessURLTable)).toExist();
    });

    it("renders an error if error prop is set", () => {
      const wrapper = mountWrapper(
        defaultStore,
        <AppViewComponent {...validProps} error={new Error("Boom!")} />,
      );
      const err = wrapper.find(Alert);
      expect(err).toExist();
      expect(err.html()).toContain("Found an error: Boom!");
    });

    it("renders a delete-error", () => {
      const wrapper = mountWrapper(
        defaultStore,
        <AppViewComponent {...validProps} deleteError={new Error("Boom!")} />,
      );
      const err = wrapper.find(Alert);
      expect(err).toExist();
      expect(err.html()).toContain("Unable to delete the application. Received: Boom!");
    });
  });

  it("renders a list of resources", () => {
    const obj = { kind: "ClusterRole", metadata: { name: "foo" } } as IResource;
    const list = {
      kind: "List",
      items: [obj, resources.deployment],
    };
    const manifest = generateYamlManifest([resources.service, list]);

    const props = {
      ...validProps,
      app: {
        ...validProps.app,
        manifest,
      } as any,
    };
    const wrapper = mountWrapper(defaultStore, <AppViewComponent {...props} />);

    const tabs = wrapper.find(ResourceTabs);
    expect(tabs.props()).toMatchObject({
      deployments: [
        new ResourceRef(resources.deployment, validProps.cluster, appRelease.namespace),
      ],
      services: [new ResourceRef(resources.service, validProps.cluster, appRelease.namespace)],
      otherResources: [new ResourceRef(obj, validProps.cluster, appRelease.namespace)],
    });
  });

  it("renders a list of roles", () => {
    const obj = { kind: "ClusterRole", metadata: { name: "foo" } } as IResource;
    const list = {
      kind: "RoleList",
      items: [obj, resources.deployment],
    };
    const manifest = generateYamlManifest([resources.service, list]);

    const props = {
      ...validProps,
      app: {
        ...validProps.app,
        manifest,
      } as any,
    };
    const wrapper = mountWrapper(defaultStore, <AppViewComponent {...props} />);

    const tabs = wrapper.find(ResourceTabs);
    expect(tabs.props()).toMatchObject({
      deployments: [
        new ResourceRef(resources.deployment, validProps.cluster, appRelease.namespace),
      ],
      services: [new ResourceRef(resources.service, validProps.cluster, appRelease.namespace)],
      otherResources: [new ResourceRef(obj, validProps.cluster, appRelease.namespace)],
    });
  });

  it("forwards statefulsets and daemonsets to the application status", () => {
    const r = [resources.statefulset, resources.daemonset];
    const manifest = generateYamlManifest(r);
    const props = {
      ...validProps,
      app: {
        ...validProps.app,
        manifest,
      } as any,
    };
    const wrapper = mountWrapper(defaultStore, <AppViewComponent {...props} />);

    const applicationStatus = wrapper.find(ApplicationStatusContainer);
    expect(applicationStatus).toExist();

    expect(applicationStatus.prop("statefulsetRefs")).toEqual([
      new ResourceRef(resources.statefulset, validProps.cluster, appRelease.namespace),
    ]);
    expect(applicationStatus.prop("daemonsetRefs")).toEqual([
      new ResourceRef(resources.daemonset, validProps.cluster, appRelease.namespace),
    ]);
  });
});
