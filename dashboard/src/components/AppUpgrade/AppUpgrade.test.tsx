import { shallow } from "enzyme";
import * as React from "react";

import { hapi } from "shared/hapi/release";
import itBehavesLike from "../../shared/specs";
import { ForbiddenError, IAppRepository, IChartState } from "../../shared/types";
import ErrorSelector from "../ErrorAlert/ErrorSelector";
import PermissionsErrorPage from "../ErrorAlert/PermissionsErrorAlert";
import UpgradeForm from "../UpgradeForm";
import SelectRepoForm from "../UpgradeForm/SelectRepoForm";
import AppUpgrade from "./AppUpgrade";

const defaultProps = {
  app: {} as hapi.release.Release,
  bindingsWithSecrets: [],
  checkChart: jest.fn(),
  clearRepo: jest.fn(),
  error: undefined,
  fetchChartVersions: jest.fn(),
  fetchRepositories: jest.fn(),
  getApp: jest.fn(),
  getBindings: jest.fn(),
  getChartValues: jest.fn(),
  getChartVersion: jest.fn(),
  kubeappsNamespace: "kubeapps",
  namespace: "default",
  push: jest.fn(),
  releaseName: "foo",
  repo: {} as IAppRepository,
  repoError: undefined,
  repos: [],
  selected: {} as IChartState["selected"],
  upgradeApp: jest.fn(),
  version: "1.0.0",
};

itBehavesLike("aLoadingComponent", { component: AppUpgrade, props: defaultProps });

it("renders the repo selection form if not introduced", () => {
  const wrapper = shallow(
    <AppUpgrade
      {...defaultProps}
      app={
        {
          chart: {
            metadata: {
              name: "bar",
              version: "1.0.0",
            },
          },
          name: "foo",
        } as hapi.release.Release
      }
      repos={[
        {
          metadata: { name: "stable" },
        } as IAppRepository,
      ]}
    />,
  );
  expect(wrapper.find(SelectRepoForm).exists()).toBe(true);
  expect(wrapper.find(ErrorSelector).exists()).toBe(false);
  expect(wrapper.find(UpgradeForm).exists()).toBe(false);
  expect(wrapper).toMatchSnapshot();
});

it("renders an error if it exists", () => {
  const repo = {
    metadata: { name: "stable" },
  } as IAppRepository;
  const wrapper = shallow(
    <AppUpgrade
      {...defaultProps}
      error={new Error("foo doesn't exists")}
      repos={[repo]}
      repo={repo}
    />,
  );
  expect(wrapper.find(ErrorSelector).exists()).toBe(true);
  expect(wrapper.find(SelectRepoForm).exists()).toBe(false);
  expect(wrapper.find(UpgradeForm).exists()).toBe(false);
  expect(wrapper.html()).toContain("Sorry! Something went wrong.");
  expect(wrapper).toMatchSnapshot();
});

it("renders a forbidden error", () => {
  const repo = {
    metadata: { name: "stable" },
  } as IAppRepository;
  const wrapper = shallow(
    <AppUpgrade {...defaultProps} error={new ForbiddenError()} repos={[repo]} repo={repo} />,
  );
  expect(wrapper.find(ErrorSelector).exists()).toBe(true);
  expect(wrapper.find(SelectRepoForm).exists()).toBe(false);
  expect(wrapper.find(UpgradeForm).exists()).toBe(false);
  expect(wrapper.html()).toContain(
    "You don&#x27;t have sufficient permissions to update foo in <span>the <code>default</code> namespace</span>",
  );
  expect(
    wrapper
      .find(ErrorSelector)
      .shallow()
      .find(PermissionsErrorPage)
      .prop("roles")[0],
  ).toMatchObject({
    apiGroup: "kubeapps.com",
    namespace: "kubeapps",
    resource: "apprepositories",
    verbs: ["get"],
  });
});

it("renders the upgrade form when the repo is available", () => {
  const repo = {
    metadata: { name: "stable" },
  } as IAppRepository;
  const wrapper = shallow(
    <AppUpgrade
      {...defaultProps}
      app={
        {
          chart: {
            metadata: {
              name: "bar",
              version: "1.0.0",
            },
          },
          name: "foo",
        } as hapi.release.Release
      }
      repos={[repo]}
      repo={repo}
    />,
  );
  expect(wrapper.find(UpgradeForm).exists()).toBe(true);
  expect(wrapper.find(ErrorSelector).exists()).toBe(false);
  expect(wrapper.find(SelectRepoForm).exists()).toBe(false);
  expect(wrapper).toMatchSnapshot();
});
