import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import { hapi } from "shared/hapi/release";
import itBehavesLike from "../../shared/specs";
import { ForbiddenError, IAppRepository, IChartState } from "../../shared/types";
import { ErrorSelector, PermissionsErrorAlert } from "../ErrorAlert";
import ErrorPageHeader from "../ErrorAlert/ErrorAlertHeader";
import UpgradeForm from "../UpgradeForm";
import SelectRepoForm from "../UpgradeForm/SelectRepoForm";
import AppUpgrade from "./AppUpgrade";

const defaultProps = {
  app: {} as hapi.release.Release,
  checkChart: jest.fn(),
  clearRepo: jest.fn(),
  error: undefined,
  fetchChartVersions: jest.fn(),
  fetchRepositories: jest.fn(),
  getApp: jest.fn(),
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
  expect(wrapper.find(SelectRepoForm)).toExist();
  expect(wrapper.find(ErrorSelector)).not.toExist();
  expect(wrapper.find(UpgradeForm)).not.toExist();
  expect(wrapper).toMatchSnapshot();
});

context("when an error exists", () => {
  it("renders a generic error message", () => {
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

    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper.find(SelectRepoForm)).not.toExist();
    expect(wrapper.find(UpgradeForm)).not.toExist();

    expect(wrapper.html()).toContain("Sorry! Something went wrong.");
    expect(wrapper).toMatchSnapshot();
  });

  it("renders a forbidden message", () => {
    const repo = {
      metadata: { name: "stable" },
    } as IAppRepository;
    const role = {
      apiGroup: "kubeapps.com",
      namespace: "kubeapps",
      resource: "apprepositories",
      verbs: ["get"],
    };
    const wrapper = mount(
      <AppUpgrade
        {...defaultProps}
        error={new ForbiddenError(JSON.stringify([role]))}
        repos={[repo]}
        repo={repo}
      />,
    );

    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper.find(SelectRepoForm)).not.toExist();
    expect(wrapper.find(UpgradeForm)).not.toExist();

    expect(wrapper.find(ErrorPageHeader).text()).toContain(
      "You don't have sufficient permissions to update foo in the default namespace",
    );
    expect(wrapper.find(PermissionsErrorAlert).prop("roles")[0]).toMatchObject(role);
  });

  it("renders a forbidden message for the repositories", () => {
    const role = {
      apiGroup: "kubeapps.com",
      namespace: "kubeapps",
      resource: "apprepositories",
      verbs: ["get"],
    };
    const wrapper = mount(<AppUpgrade {...defaultProps} repoError={new ForbiddenError()} />);

    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper.find(SelectRepoForm)).not.toExist();
    expect(wrapper.find(UpgradeForm)).not.toExist();

    expect(wrapper.find(ErrorPageHeader).text()).toContain(
      "You don't have sufficient permissions to view App Repositories in the kubeapps namespace",
    );
    expect(wrapper.find(PermissionsErrorAlert).prop("roles")[0]).toMatchObject(role);
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
  expect(wrapper.find(UpgradeForm)).toExist();
  expect(wrapper.find(ErrorSelector)).not.toExist();
  expect(wrapper.find(SelectRepoForm)).not.toExist();
  expect(wrapper).toMatchSnapshot();
});
