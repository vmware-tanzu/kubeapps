import { shallow } from "enzyme";
import { safeDump as yamlSafeDump } from "js-yaml";
import * as React from "react";

import { hapi } from "../../shared/hapi/release";
import { IResource, ForbiddenError, NotFoundError } from "../../shared/types";
import DeploymentStatus from "../DeploymentStatus";
import { PermissionsErrorAlert, NotFoundErrorAlert } from "../ErrorAlert";
import AppControls from "./AppControls";
import AppDetails from "./AppDetails";
import AppNotes from "./AppNotes";
import AppViewComponent, { IAppViewProps } from "./AppView";
import ChartInfo from "./ChartInfo";
import ServiceTable from "./ServiceTable";

describe("AppViewComponent", () => {
  // Generates a Yaml file separated by --- containing every object passed.
  const generateYamlManifest = (items: any[]): string => {
    let yamlManifest = "";
    items.forEach(i => {
      yamlManifest += "---\n" + yamlSafeDump(i);
    });
    return yamlManifest;
  };

  let validProps: IAppViewProps;
  beforeEach(() => {
    const appRelease = hapi.release.Release.create({
      info: hapi.release.Info.create(),
      namespace: "weee",
    });

    validProps = {
      app: appRelease,
      deleteApp: jest.fn(),
      deleteError: undefined,
      error: undefined,
      getApp: jest.fn(),
      namespace: "my-happy-place",
      releaseName: "mr-sunshine",
    };
  });

  it("renders a loading message if info is not present", () => {
    validProps.app.info = null;
    const wrapper = shallow(<AppViewComponent {...validProps} />);
    expect(wrapper.text()).toBe("Loading");
  });

  describe("State initialization", () => {
    const resources = {
      configMap: { apiVersion: "v1", kind: "ConfigMap", metadata: { name: "cm-one" } },
      deployment: {
        apiVersion: "extensions/v1beta1",
        kind: "Deployment",
        metadata: { name: "deployment-one" },
      },
      service: { apiVersion: "v1", kind: "Service", metadata: { name: "svc-one" } },
    };

    /*
      The imported manifest contains one deployment, one service, one config map and some bogus manifests.
      We only set websockets for deployment and services
    */
    it("sets a list of web sockets for its deployments and services", () => {
      const manifest = generateYamlManifest([
        resources.deployment,
        resources.service,
        resources.configMap,
      ]);

      const wrapper = shallow(<AppViewComponent {...validProps} />);
      validProps.app.manifest = manifest;
      // setProps again so we trigger componentWillReceiveProps
      wrapper.setProps(validProps);
      const sockets: WebSocket[] = wrapper.state("sockets");

      expect(sockets.length).toEqual(2);
      expect(sockets[0].url).toBe(
        "ws://localhost/api/kube/apis/apps/v1beta1/namespaces/weee/deployments?watch=true&fieldSelector=metadata.name%3Ddeployment-one",
      );
      expect(sockets[1].url).toBe(
        "ws://localhost/api/kube/api/v1/namespaces/weee/services?watch=true&fieldSelector=metadata.name%3Dsvc-one",
      );
    });

    it("stores other k8s resources directly in the state", () => {
      const wrapper = shallow(<AppViewComponent {...validProps} />);
      const manifest = generateYamlManifest([resources.configMap, resources.deployment]);

      validProps.app.manifest = manifest;
      wrapper.setProps(validProps);

      const otherResources: Map<string, IResource> = wrapper.state("otherResources");
      const configMap = otherResources["ConfigMap/cm-one"];
      expect(Object.keys(otherResources).length).toEqual(1);

      // It sets the websocket for the deployment
      const sockets: WebSocket[] = wrapper.state("sockets");
      expect(sockets.length).toEqual(1);

      expect(configMap).toBeDefined();
      expect(configMap.metadata.name).toEqual("cm-one");
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

      const otherResources: Map<string, IResource> = wrapper.state("otherResources");
      expect(Object.keys(otherResources).length).toBe(0);

      const sockets: WebSocket[] = wrapper.state("sockets");
      expect(sockets.length).toEqual(0);
    });
  });

  describe("renderization", () => {
    it("renders all the elements of an application", () => {
      const wrapper = shallow(<AppViewComponent {...validProps} />);
      expect(wrapper.find(ChartInfo).exists()).toBe(true);
      expect(wrapper.find(DeploymentStatus).exists()).toBe(true);
      expect(wrapper.find(AppControls).exists()).toBe(true);
      expect(wrapper.find(ServiceTable).exists()).toBe(true);
      expect(wrapper.find(AppNotes).exists()).toBe(true);
      expect(wrapper.find(AppDetails).exists()).toBe(true);
    });
    it("renders a forbidden error if it exists", () => {
      const wrapper = shallow(
        <AppViewComponent {...validProps} deleteError={new ForbiddenError()} />,
      );
      expect(wrapper.find(PermissionsErrorAlert).exists()).toBe(true);
      expect(wrapper.find(PermissionsErrorAlert).props()).toMatchObject({
        action: 'delete Application "mr-sunshine"',
        namespace: "my-happy-place",
      });
    });
    it("renders a not-found error if it exists", () => {
      const wrapper = shallow(
        <AppViewComponent {...validProps} deleteError={new NotFoundError()} />,
      );
      expect(wrapper.find(NotFoundErrorAlert).exists()).toBe(true);
      expect(wrapper.find(NotFoundErrorAlert).props()).toMatchObject({
        namespace: "my-happy-place",
        resource: 'Application "mr-sunshine"',
      });
    });
  });
});
