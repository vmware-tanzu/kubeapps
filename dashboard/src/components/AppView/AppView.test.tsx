import { shallow } from "enzyme";
import context from "jest-plugin-context";
import { safeDump as yamlSafeDump, YAMLException } from "js-yaml";
import * as React from "react";

import AccessURLTable from "../../containers/AccessURLTableContainer";
import ApplicationStatus from "../../containers/ApplicationStatusContainer";
import ApplicationStatusContainer from "../../containers/ApplicationStatusContainer";
import { hapi } from "../../shared/hapi/release";
import ResourceRef from "../../shared/ResourceRef";
import itBehavesLike from "../../shared/specs";
import { ForbiddenError, IResource, NotFoundError } from "../../shared/types";
import { ErrorSelector } from "../ErrorAlert";
import PermissionsErrorPage from "../ErrorAlert/PermissionsErrorAlert";
import AppControls from "./AppControls";
import AppNotes from "./AppNotes";
import AppViewComponent, { IAppViewProps } from "./AppView";
import ChartInfo from "./ChartInfo";
import ResourceTable from "./ResourceTable";

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

  const validProps: IAppViewProps = {
    app: appRelease,
    deleteApp: jest.fn(),
    deleteError: undefined,
    error: undefined,
    getAppWithUpdateInfo: jest.fn(),
    namespace: "my-happy-place",
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

  context("when app info is null", () => {
    itBehavesLike("aLoadingComponent", {
      component: AppViewComponent,
      props: {
        ...validProps,
        app: { ...validProps.app, info: null },
      },
    });
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

      const wrapper = shallow(<AppViewComponent {...validProps} />);
      validProps.app.manifest = manifest;
      // setProps again so we trigger componentWillReceiveProps
      wrapper.setProps(validProps);

      expect(wrapper.state("deployRefs")).toEqual([
        new ResourceRef(resources.deployment, appRelease.namespace),
      ]);
      expect(wrapper.state("serviceRefs")).toEqual([
        new ResourceRef(resources.service, appRelease.namespace),
      ]);
      expect(wrapper.state("ingressRefs")).toEqual([
        new ResourceRef(resources.ingress, appRelease.namespace),
      ]);
      expect(wrapper.state("secretRefs")).toEqual([
        new ResourceRef(resources.secret, appRelease.namespace),
      ]);
    });

    it("stores other k8s resources directly in the state", () => {
      const wrapper = shallow(<AppViewComponent {...validProps} />);
      const manifest = generateYamlManifest([
        resources.deployment,
        resources.service,
        resources.configMap,
        resources.secret,
      ]);

      validProps.app.manifest = manifest;
      wrapper.setProps(validProps);

      const otherResources: ResourceRef[] = wrapper.state("otherResources");
      const configMap = otherResources[0];
      // It should skip deployments, services and secrets from "other resources"
      expect(otherResources.length).toEqual(1);

      expect(configMap).toBeDefined();
      expect(configMap.name).toEqual("cm-one");
    });

    it("does not store empty resources, bogus or without kind attribute", () => {
      const wrapper = shallow(<AppViewComponent {...validProps} />);
      const manifest = generateYamlManifest([
        { apiVersion: "v1", metadata: { name: "cm-one" } },
        {},
        "# This is a comment",
        " ",
      ]);

      validProps.app.manifest = manifest;
      wrapper.setProps(validProps);

      expect(wrapper.state("otherResources")).toEqual([]);
      expect(wrapper.state("deployRefs")).toEqual([]);
      expect(wrapper.state("serviceRefs")).toEqual([]);
      expect(wrapper.state("ingressRefs")).toEqual([]);
      expect(wrapper.state("secretRefs")).toEqual([]);
    });

    // See https://github.com/kubeapps/kubeapps/issues/632
    it("supports manifests with duplicated keys", () => {
      const wrapper = shallow(<AppViewComponent {...validProps} />);
      const manifest = `
      apiVersion: v1
      metadata:
        name: cm-one
        labels:
          chart: cm-1.2.3
          chart: cm-1.2.3
`;

      validProps.app.manifest = manifest;
      expect(() => {
        wrapper.setProps(validProps);
      }).not.toThrow(YAMLException);
    });
  });

  describe("renderization", () => {
    it("renders all the elements of an application", () => {
      const wrapper = shallow(<AppViewComponent {...validProps} />);
      expect(wrapper.find(ChartInfo).exists()).toBe(true);
      expect(wrapper.find(ApplicationStatus).exists()).toBe(true);
      expect(wrapper.find(AppControls).exists()).toBe(true);
      expect(wrapper.find(AppNotes).exists()).toBe(true);
      expect(wrapper.find(ResourceTable).exists()).toBe(true);
      expect(wrapper.find(AccessURLTable).exists()).toBe(true);
    });

    it("renders an error if error prop is set", () => {
      const wrapper = shallow(<AppViewComponent {...validProps} error={new NotFoundError()} />);
      const err = wrapper.find(ErrorSelector);
      expect(err.exists()).toBe(true);
      expect(err.html()).toContain("Application mr-sunshine not found");
    });

    it("renders a forbidden delete-error if if the deleteError prop is a ForbiddenError", () => {
      const wrapper = shallow(
        <AppViewComponent {...validProps} deleteError={new ForbiddenError()} />,
      );
      const err = wrapper.find(ErrorSelector);
      expect(err.exists()).toBe(true);
      expect(
        err
          .shallow()
          .find(PermissionsErrorPage)
          .props(),
      ).toMatchObject({
        action: "delete Application mr-sunshine",
        namespace: "my-happy-place",
      });
    });

    it("renders a not-found delete-error if the deleteError prop is a NotFound error", () => {
      const wrapper = shallow(
        <AppViewComponent {...validProps} deleteError={new NotFoundError()} />,
      );
      const err = wrapper.find(ErrorSelector);
      expect(err.exists()).toBe(true);
      expect(err.html()).toContain("Application mr-sunshine not found");
    });
  });

  it("forwards services/ingresses", () => {
    const wrapper = shallow(<AppViewComponent {...validProps} />);
    const ingress = {
      isFetching: false,
      item: {
        metadata: {
          name: "foo",
        },
        spec: {
          rules: [
            {
              host: "foo.bar",
              http: {
                paths: [{ path: "/ready" }],
              },
            },
          ],
        },
      },
    };
    const ingressRefs = [new ResourceRef(ingress.item as IResource, "default")];
    const service = {
      isFetching: false,
      item: {
        apiVersion: "v1",
        kind: "Service",
        metadata: {
          name: "foo",
        },
        spec: {},
      },
    };
    const serviceRefs = [new ResourceRef(service.item as IResource, "default")];

    wrapper.setState({ ingressRefs, serviceRefs });

    const accessURLTable = wrapper.find(AccessURLTable);
    expect(accessURLTable).toExist();
    expect(accessURLTable.props()).toMatchObject({ ingressRefs, serviceRefs });

    const svcTable = wrapper.find(ResourceTable).findWhere(t => t.prop("title") === "Services");
    expect(svcTable).toExist();
    expect(svcTable.prop("resourceRefs")).toEqual(serviceRefs);
  });

  it("forwards deployments", () => {
    const wrapper = shallow(<AppViewComponent {...validProps} />);
    const deployment = {
      metadata: {
        name: "foo",
      },
      spec: {},
    };
    const deployRefs = [new ResourceRef(deployment as IResource, "default")];

    wrapper.setState({ deployRefs });

    const depTable = wrapper
      .find(ResourceTable)
      .filterWhere(e => e.prop("title") === "Deployments");
    expect(depTable).toExist();
    expect(depTable.prop("resourceRefs")).toEqual(deployRefs);
  });

  it("forwards statefulsets", () => {
    const wrapper = shallow(<AppViewComponent {...validProps} />);
    const r = {
      metadata: {
        name: "foo",
      },
      spec: {},
    };
    const ref = [new ResourceRef(r as IResource, "default")];

    wrapper.setState({ statefulSetRefs: ref });

    const depTable = wrapper
      .find(ResourceTable)
      .filterWhere(e => e.prop("title") === "StatefulSets");
    expect(depTable).toExist();
    expect(depTable.prop("resourceRefs")).toEqual(ref);
  });

  it("forwards daemonsets", () => {
    const wrapper = shallow(<AppViewComponent {...validProps} />);
    const r = {
      metadata: {
        name: "foo",
      },
      spec: {},
    };
    const ref = [new ResourceRef(r as IResource, "default")];

    wrapper.setState({ daemonSetRefs: ref });

    const depTable = wrapper.find(ResourceTable).filterWhere(e => e.prop("title") === "DaemonSets");
    expect(depTable).toExist();
    expect(depTable.prop("resourceRefs")).toEqual(ref);
  });

  it("forwards other resources", () => {
    const wrapper = shallow(<AppViewComponent {...validProps} />);
    const otherResource = {
      metadata: {
        name: "foo",
      },
      spec: {},
    };
    const otherResources = [otherResource];

    wrapper.setState({ otherResources });

    const orTable = wrapper
      .find(ResourceTable)
      .filterWhere(e => e.prop("title") === "Other Resources");
    expect(orTable).toExist();
    expect(orTable.prop("resourceRefs")).toEqual([otherResource]);
  });

  it("renders a list of resources", () => {
    const obj = { kind: "ClusterRole", metadata: { name: "foo" } } as IResource;
    const list = {
      kind: "List",
      items: [obj, resources.deployment],
    };
    const manifest = generateYamlManifest([resources.service, list]);

    const wrapper = shallow(<AppViewComponent {...validProps} />);
    validProps.app.manifest = manifest;
    // setProps again so we trigger componentWillReceiveProps
    wrapper.setProps(validProps);

    expect(wrapper.state()).toMatchObject({
      deployRefs: [new ResourceRef(resources.deployment, appRelease.namespace)],
      serviceRefs: [new ResourceRef(resources.service, appRelease.namespace)],
      otherResources: [new ResourceRef(obj, appRelease.namespace)],
    });
  });

  it("renders a list of roles", () => {
    const obj = { kind: "ClusterRole", metadata: { name: "foo" } } as IResource;
    const list = {
      kind: "RoleList",
      items: [obj, resources.deployment],
    };
    const manifest = generateYamlManifest([resources.service, list]);

    const wrapper = shallow(<AppViewComponent {...validProps} />);
    validProps.app.manifest = manifest;
    // setProps again so we trigger componentWillReceiveProps
    wrapper.setProps(validProps);

    expect(wrapper.state()).toMatchObject({
      deployRefs: [new ResourceRef(resources.deployment, appRelease.namespace)],
      serviceRefs: [new ResourceRef(resources.service, appRelease.namespace)],
      otherResources: [new ResourceRef(obj, appRelease.namespace)],
    });
  });

  it("forwards statefulsets and daemonsets", () => {
    const r = [resources.statefulset, resources.daemonset];
    const manifest = generateYamlManifest(r);
    const wrapper = shallow(<AppViewComponent {...validProps} />);
    validProps.app.manifest = manifest;
    // setProps again so we trigger componentWillReceiveProps
    wrapper.setProps(validProps);

    const applicationStatus = wrapper.find(ApplicationStatusContainer);
    expect(applicationStatus).toExist();

    expect(applicationStatus.prop("statefulsetRefs")).toEqual([
      new ResourceRef(resources.statefulset, appRelease.namespace),
    ]);
    expect(applicationStatus.prop("daemonsetRefs")).toEqual([
      new ResourceRef(resources.daemonset, appRelease.namespace),
    ]);
  });
});
