import CustomAppView from ".";
import { IRelease } from "shared/types";
import { CustomComponent } from "../../../RemoteComponent";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IAppViewResourceRefs } from "../AppView";

const defaultState = {
  config: { remoteComponentsUrl: "" },
};

const DEFAULT_CUSTOM_APP_PROPS = {
  app: {
    chart: {
      metadata: {
        name: "bar",
        appVersion: "0.0.1",
        description: "test chart",
        icon: "icon.png",
        version: "1.0.0",
      },
    },
    name: "foo",
  } as IRelease,
  resourceRefs: {
    ingresses: [],
    deployments: [
      {
        cluster: "default",
        apiVersion: "apps/v1",
        kind: "Deployment",
        plural: "deployments",
        namespaced: true,
        name: "ssh-server-example",
        namespace: "demo-namespace",
      },
    ],
    statefulsets: [],
    daemonsets: [],
    otherResources: [
      {
        cluster: "default",
        apiVersion: "v1",
        kind: "PersistentVolumeClaim",
        plural: "persistentvolumeclaims",
        namespaced: true,
        name: "ssh-server-example-root-pv-claim",
        namespace: "demo-namespace",
      },
    ],
    services: [
      {
        cluster: "default",
        apiVersion: "v1",
        kind: "Service",
        plural: "services",
        namespaced: true,
        name: "ssh-server-example",
        namespace: "demo-namespace",
      },
    ],
    secrets: [],
  } as unknown as IAppViewResourceRefs,
};

it("should render a custom app view", () => {
  const wrapper = mountWrapper(
    getStore(defaultState),
    <CustomAppView {...DEFAULT_CUSTOM_APP_PROPS} />,
  );
  expect(wrapper.find(CustomAppView)).toExist();
});

it("should render the remote component", () => {
  const wrapper = mountWrapper(
    getStore(defaultState),
    <CustomAppView {...DEFAULT_CUSTOM_APP_PROPS} />,
  );
  expect(wrapper.find(CustomComponent)).toExist();
});

it("should render the remote component with the default URL", () => {
  const wrapper = mountWrapper(
    getStore(defaultState),
    <CustomAppView {...DEFAULT_CUSTOM_APP_PROPS} />,
  );
  expect(wrapper.find(CustomComponent)).toExist();
  expect(wrapper.find(CustomComponent).prop("url")).toContain("custom_components.js");
});

it("should render the remote component with the URL if set in the config", () => {
  const wrapper = mountWrapper(
    getStore({
      config: { remoteComponentsUrl: "www.thiswebsite.com" },
    }),
    <CustomAppView {...DEFAULT_CUSTOM_APP_PROPS} />,
  );
  expect(wrapper.find(CustomComponent).prop("url")).toBe("www.thiswebsite.com");
});
