import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";
import { safeDump as yamlSafeDump, YAMLException } from "js-yaml";
import * as moxios from "moxios";
import * as React from "react";

import { axios } from "../../shared/Auth";
import { hapi } from "../../shared/hapi/release";
import itBehavesLike from "../../shared/specs";
import { ForbiddenError, IIngressSpec, IResource, NotFoundError } from "../../shared/types";
import DeploymentStatus from "../DeploymentStatus";
import { ErrorSelector } from "../ErrorAlert";
import PermissionsErrorPage from "../ErrorAlert/PermissionsErrorAlert";
import AccessURLTable from "./AccessURLTable";
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

  const appRelease = hapi.release.Release.create({
    info: hapi.release.Info.create(),
    namespace: "weee",
  });

  const validProps: IAppViewProps = {
    app: appRelease,
    deleteApp: jest.fn(),
    deleteError: undefined,
    error: undefined,
    getApp: jest.fn(),
    namespace: "my-happy-place",
    releaseName: "mr-sunshine",
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

  context("when otherResources is null", () => {
    itBehavesLike("aLoadingComponent", {
      component: AppViewComponent,
      props: validProps,
      state: { otherResources: null },
    });
  });

  describe("State initialization", () => {
    const resources = {
      configMap: { apiVersion: "v1", kind: "ConfigMap", metadata: { name: "cm-one" } },
      deployment: {
        apiVersion: "apps/v1beta1",
        kind: "Deployment",
        metadata: { name: "deployment-one" },
      },
      service: { apiVersion: "v1", kind: "Service", metadata: { name: "svc-one" } },
      ingress: {
        apiVersion: "extensions/v1beta1",
        kind: "Ingress",
        metadata: { name: "ingress-one" },
      },
      secret: { apiVersion: "v1", kind: "Secret", metadata: { name: "secret-one" } },
    };

    beforeEach(() => {
      moxios.install(axios);
    });

    afterEach(() => {
      moxios.uninstall(axios);
    });

    /*
      The imported manifest contains one deployment, one service, one config map and some bogus manifests.
      We only set websockets for deployment and services
    */
    it("sets a list of web sockets for its deployments and services", () => {
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
      const sockets: WebSocket[] = wrapper.state("sockets");

      expect(sockets.length).toEqual(3);
      expect(sockets[0].url).toBe(
        "ws://localhost/api/kube/apis/apps/v1beta1/namespaces/weee/deployments?watch=true&fieldSelector=metadata.name%3Ddeployment-one",
      );
      expect(sockets[1].url).toBe(
        "ws://localhost/api/kube/api/v1/namespaces/weee/services?watch=true&fieldSelector=metadata.name%3Dsvc-one",
      );
      expect(sockets[2].url).toBe(
        "ws://localhost/api/kube/apis/extensions/v1beta1/namespaces/weee/ingresses?watch=true&fieldSelector=metadata.name%3Dingress-one",
      );
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

      const otherResources: { [r: string]: IResource } = wrapper.state("otherResources");
      const configMap = otherResources["ConfigMap/cm-one"];
      // It should skip deployments, services and secrets from "other resources"
      expect(Object.keys(otherResources).length).toEqual(1);

      // It sets the websocket for the deployment
      const sockets: WebSocket[] = wrapper.state("sockets");
      expect(sockets.length).toEqual(2);

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

    it("requests a secret", done => {
      const secretURL = `/api/kube/api/v1/namespaces/${appRelease.namespace}/secrets/${
        resources.secret.metadata.name
      }`;
      moxios.stubRequest(secretURL, {
        response: JSON.stringify(resources.secret),
        status: 200,
      });

      const wrapper = shallow(<AppViewComponent {...validProps} />);
      const manifest = generateYamlManifest([
        resources.deployment,
        resources.service,
        resources.configMap,
        resources.secret,
      ]);

      validProps.app.manifest = manifest;
      wrapper.setProps(validProps);

      // Wait for the async call to be performed
      setTimeout(() => {
        expect(moxios.requests.mostRecent().url).toBe(secretURL);
        expect(wrapper.state("secrets")).toMatchObject({
          [`Secret/${resources.secret.metadata.name}`]: resources.secret,
        });
        done();
      }, 1);
    });
  });

  describe("renderization", () => {
    it("renders all the elements of an application", () => {
      const wrapper = mount(<AppViewComponent {...validProps} />);
      const service = {
        metadata: { name: "foo" },
        spec: { type: "loadBalancer", ports: [{ port: 8080 }] },
        status: { ingress: [{ loadBalancer: { ip: "1.2.3.4" } }] },
      } as IResource;
      const services = {};
      services[service.metadata.name] = service;
      wrapper.setState({ services });
      expect(wrapper.find(ChartInfo).exists()).toBe(true);
      expect(wrapper.find(DeploymentStatus).exists()).toBe(true);
      expect(wrapper.find(AppControls).exists()).toBe(true);
      expect(wrapper.find(ServiceTable).exists()).toBe(true);
      expect(wrapper.find(AppNotes).exists()).toBe(true);
      expect(wrapper.find(AppDetails).exists()).toBe(true);
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

    it("renders an URL table if an Ingress exists", () => {
      const wrapper = mount(<AppViewComponent {...validProps} />);
      const ingress = {
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
        } as IIngressSpec,
      } as IResource;
      const ingresses = {};
      ingresses[ingress.metadata.name] = ingress;
      wrapper.setState({ ingresses });
      const urlTable = wrapper.find(AccessURLTable);
      expect(urlTable).toExist();
      expect(urlTable.text()).toContain("Ingress");
      expect(urlTable.text()).toContain("http://foo.bar/ready");
    });
  });
});
