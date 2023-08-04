// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CustomComponent } from "RemoteComponent";
import {
  AvailablePackageDetail,
  InstalledPackageDetail,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import ResourceRef from "shared/ResourceRef";
import { getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import CustomAppView from ".";
import { IAppViewResourceRefs } from "../AppView";

const defaultState = {
  config: { remoteComponentsUrl: "" },
} as IStoreState;

const defaultProps = {
  app: {
    availablePackageRef: {
      identifier: "bar",
    },
  } as InstalledPackageDetail,
  resourceRefs: {
    ingresses: [] as ResourceRef[],
    deployments: [
      new ResourceRef("default", "deployments", true, "demo-namespace", {
        apiVersion: "apps/v1",
        kind: "Deployment",
        name: "ssh-server-example",
        namespace: "demo-namespace",
      }),
    ] as ResourceRef[],
    statefulsets: [] as ResourceRef[],
    daemonsets: [] as ResourceRef[],
    otherResources: [
      {
        cluster: "default",
        apiVersion: "v1",
        kind: "PersistentVolumeClaim",
        plural: "persistentvolumeclaims",
        namespaced: true,
        name: "ssh-server-example-root-pv-claim",
        namespace: "demo-namespace",
      } as ResourceRef,
    ] as ResourceRef[],
    services: [
      {
        cluster: "default",
        apiVersion: "v1",
        kind: "Service",
        plural: "services",
        namespaced: true,
        name: "ssh-server-example",
        namespace: "demo-namespace",
      } as ResourceRef,
    ] as ResourceRef[],
    secrets: [] as ResourceRef[],
  } as IAppViewResourceRefs,
  appDetails: {} as AvailablePackageDetail,
};

// Ensure remote-component doesn't trigger external requests during this test.
const xhrMock: Partial<XMLHttpRequest> = {
  open: jest.fn(),
  send: jest.fn(),
  setRequestHeader: jest.fn(),
  readyState: 4,
  status: 200,
  response: "Hello World!",
};

beforeAll((): void => {
  jest.spyOn(window, "XMLHttpRequest").mockImplementation(() => xhrMock as XMLHttpRequest);
});

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
      ...initialState,
      config: { remoteComponentsUrl: "www.thiswebsite.com" },
    } as Partial<IStoreState>),
    <CustomAppView {...defaultProps} />,
  );
  expect(wrapper.find(CustomComponent).prop("url")).toBe("www.thiswebsite.com");
});
