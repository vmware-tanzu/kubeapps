import { shallow } from "enzyme";
import * as React from "react";

import { hapi } from "shared/hapi/release";
import { IAppRepository, IChartState } from "shared/types";
import DeploymentErrors from "../DeploymentForm/DeploymentErrors";
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

it("renders a loading message if apps object is empty", () => {
  const wrapper = shallow(<AppUpgrade {...defaultProps} />);
  expect(wrapper.text()).toBe("Loading");
});

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
  expect(wrapper.find(DeploymentErrors).exists()).toBe(false);
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
  expect(wrapper.find(DeploymentErrors).exists()).toBe(true);
  expect(wrapper.find(SelectRepoForm).exists()).toBe(false);
  expect(wrapper.find(UpgradeForm).exists()).toBe(false);
  expect(wrapper).toMatchSnapshot();
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
  expect(wrapper.find(DeploymentErrors).exists()).toBe(false);
  expect(wrapper.find(SelectRepoForm).exists()).toBe(false);
  expect(wrapper).toMatchSnapshot();
});
