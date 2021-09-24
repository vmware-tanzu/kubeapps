import {
  InstalledPackageDetail,
  AvailablePackageDetail,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import CustomAppView from ".";
import { CustomComponent } from "../../../RemoteComponent";
import { IAppViewResourceRefs } from "../AppView";

const defaultState = {
  config: { remoteComponentsUrl: "" },
};

const defaultProps = {
  app: {
    availablePackageRef: {
      identifier: "bar",
    },
  } as InstalledPackageDetail,
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
  appDetails: {} as AvailablePackageDetail,
};

it("should render a custom app view", () => {
  const wrapper = mountWrapper(getStore(defaultState), <CustomAppView {...defaultProps} />);
  expect(wrapper.find(CustomAppView)).toExist();
});

it("should render the remote component", () => {
  const wrapper = mountWrapper(getStore(defaultState), <CustomAppView {...defaultProps} />);
  expect(wrapper.find(CustomComponent)).toExist();
});

it("should render the remote component with the default URL", () => {
  const wrapper = mountWrapper(getStore(defaultState), <CustomAppView {...defaultProps} />);
  expect(wrapper.find(CustomComponent)).toExist();
  expect(wrapper.find(CustomComponent).prop("url")).toContain("custom_components.js");
});

it("should render the remote component with the URL if set in the config", () => {
  const wrapper = mountWrapper(
    getStore({
      config: { remoteComponentsUrl: "www.thiswebsite.com" },
    }),
    <CustomAppView {...defaultProps} />,
  );
  expect(wrapper.find(CustomComponent).prop("url")).toBe("www.thiswebsite.com");
});
