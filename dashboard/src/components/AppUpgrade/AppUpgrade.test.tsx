import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";
import { BrowserRouter } from "react-router-dom";

import { hapi } from "shared/hapi/release";
import itBehavesLike from "../../shared/specs";
import { ForbiddenError, IAppRepository, IChartState, IRelease } from "../../shared/types";
import { ErrorSelector, MessageAlert, PermissionsErrorAlert } from "../ErrorAlert";
import ErrorPageHeader from "../ErrorAlert/ErrorAlertHeader";
import SelectRepoForm from "../SelectRepoForm";
import UpgradeForm from "../UpgradeForm";
import AppUpgrade from "./AppUpgrade";

const defaultProps = {
  app: {} as hapi.release.Release,
  appsIsFetching: false,
  reposIsFetching: false,
  repoName: "",
  repoNamespace: "chart-namespace",
  isFetching: false,
  checkChart: jest.fn(),
  clearRepo: jest.fn(),
  appsError: undefined,
  chartsError: undefined,
  fetchChartVersions: jest.fn(),
  fetchRepositories: jest.fn(),
  getAppWithUpdateInfo: jest.fn(),
  getChartVersion: jest.fn(),
  deployed: {} as IChartState["deployed"],
  getDeployedChartVersion: jest.fn(),
  kubeappsNamespace: "kubeapps",
  namespace: "default",
  push: jest.fn(),
  goBack: jest.fn(),
  releaseName: "foo",
  repo: {} as IAppRepository,
  repoError: undefined,
  repos: [],
  selected: {} as IChartState["selected"],
  upgradeApp: jest.fn(),
  version: "1.0.0",
};

beforeEach(() => {
  jest.resetAllMocks();
});

itBehavesLike("aLoadingComponent", {
  component: AppUpgrade,
  props: { ...defaultProps, isFetching: true },
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
          updateInfo: { repository: {} },
        } as IRelease
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
        appsError={new Error("foo doesn't exists")}
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
    const role = {
      apiGroup: "kubeapps.com",
      namespace: "kubeapps",
      resource: "apprepositories",
      verbs: ["get"],
    };
    const wrapper = mount(
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
            updateInfo: { repository: {} },
          } as IRelease
        }
        repoError={new ForbiddenError(JSON.stringify([role]))}
      />,
    );
    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper.find(UpgradeForm)).not.toExist();

    expect(wrapper.find(ErrorPageHeader).text()).toContain(
      "You don't have sufficient permissions to view App Repositories in the kubeapps namespace",
    );
    expect(wrapper.find(PermissionsErrorAlert).prop("roles")[0]).toMatchObject(role);
  });

  it("renders a warning message if there are no repositories", () => {
    const wrapper = mount(
      <BrowserRouter>
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
              updateInfo: { repository: {} },
            } as IRelease
          }
          repos={[]}
        />
      </BrowserRouter>,
    );

    expect(wrapper.find(SelectRepoForm).find(MessageAlert)).toExist();
    expect(wrapper.find(UpgradeForm)).not.toExist();

    expect(
      wrapper
        .find(MessageAlert)
        .children()
        .text(),
    ).toContain("Chart repositories not found");
  });
});

it("renders the upgrade form when the repo is available", () => {
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
          updateInfo: { repository: {} },
        } as IRelease
      }
    />,
  );
  wrapper.setProps({ repoName: "foobar" });
  expect(wrapper.find(UpgradeForm)).toExist();
  expect(wrapper.find(ErrorSelector)).not.toExist();
  expect(wrapper.find(SelectRepoForm)).not.toExist();
  expect(wrapper).toMatchSnapshot();
});

it("skips the repo selection form if the app contains upgrade info", () => {
  const repo = {
    metadata: { name: "stable" },
  } as IAppRepository;
  const app = {
    chart: {
      metadata: {
        name: "bar",
        version: "1.0.0",
      },
    },
    name: "foo",
    updateInfo: {
      upToDate: true,
      chartLatestVersion: "1.1.0",
      appLatestVersion: "1.1.0",
      repository: { name: "stable", url: "" },
    },
  } as IRelease;
  const wrapper = shallow(<AppUpgrade {...defaultProps} repos={[repo]} />);
  wrapper.setProps({ app });
  expect(wrapper.find(UpgradeForm)).toExist();
  expect(wrapper.find(ErrorSelector)).not.toExist();
  expect(wrapper.find(SelectRepoForm)).not.toExist();
  expect(wrapper).toMatchSnapshot();
});

describe("when receiving new props", () => {
  it("should request the deployed chart when the app and repo are populated", () => {
    const app = {
      chart: {
        metadata: {
          name: "bar",
          version: "1.0.0",
        },
      },
    } as IRelease;
    const getDeployedChartVersion = jest.fn();
    const wrapper = shallow(
      <AppUpgrade {...defaultProps} getDeployedChartVersion={getDeployedChartVersion} />,
    );
    wrapper.setProps({ repoName: "stable", app });
    wrapper.update();
    expect(getDeployedChartVersion).toHaveBeenCalledWith(defaultProps.repoNamespace, "stable/bar", "1.0.0");
  });

  it("should request the deployed chart when the repo is populated later", () => {
    const app = {
      chart: {
        metadata: {
          name: "bar",
          version: "1.0.0",
        },
      },
    } as IRelease;
    const getDeployedChartVersion = jest.fn();
    const wrapper = shallow(
      <AppUpgrade {...defaultProps} app={app} getDeployedChartVersion={getDeployedChartVersion} />,
    );
    expect(getDeployedChartVersion).not.toHaveBeenCalled();
    wrapper.setProps({ repoName: "stable" });
    wrapper.update();
    expect(getDeployedChartVersion).toHaveBeenCalledWith(defaultProps.repoNamespace, "stable/bar", "1.0.0");
  });

  it("a new app should re-trigger the deployed chart retrieval", () => {
    const app = {
      chart: {
        metadata: {
          name: "bar",
          version: "1.0.0",
        },
      },
    } as IRelease;
    const getDeployedChartVersion = jest.fn();
    const wrapper = shallow(
      <AppUpgrade {...defaultProps} getDeployedChartVersion={getDeployedChartVersion} />,
    );
    wrapper.setProps({ repoName: "stable", app });
    expect(getDeployedChartVersion).toHaveBeenCalledWith(defaultProps.repoNamespace, "stable/bar", "1.0.0");

    const app2 = {
      chart: {
        metadata: {
          name: "foobar",
          version: "1.0.0",
        },
      },
    } as IRelease;
    wrapper.setProps({ app: app2 });
    expect(getDeployedChartVersion).toHaveBeenCalledWith(defaultProps.repoNamespace, "stable/foobar", "1.0.0");
  });
});
