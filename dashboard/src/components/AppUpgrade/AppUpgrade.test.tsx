import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

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
  isFetching: false,
  checkChart: jest.fn(),
  clearRepo: jest.fn(),
  error: undefined,
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

  it("renders a warning message if there are no repositories", () => {
    const wrapper = shallow(<AppUpgrade {...defaultProps} />);

    expect(wrapper.find(MessageAlert)).toExist();
    expect(wrapper.find(SelectRepoForm)).not.toExist();
    expect(wrapper.find(UpgradeForm)).not.toExist();

    expect(
      wrapper
        .find(MessageAlert)
        .children()
        .text(),
    ).toContain("Chart repositories not found");
  });

  it("renders an error message if the app information is missing some metadata", () => {
    const repo = {
      metadata: { name: "stable" },
    } as IAppRepository;
    const wrapper = mount(
      <AppUpgrade
        {...defaultProps}
        repos={[repo]}
        app={
          {
            chart: {
              metadata: {},
            },
            name: "foo",
          } as IRelease
        }
      />,
    );

    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper.find(SelectRepoForm)).not.toExist();
    expect(wrapper.find(UpgradeForm)).not.toExist();

    expect(wrapper.find(ErrorSelector).text()).toContain(
      "Unable to obtain the required information to upgrade",
    );
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
        } as IRelease
      }
      repos={[repo]}
    />,
  );
  wrapper.setProps({ repo });
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
  it("should set the source repository in the state if present as a property", () => {
    const repo = { metadata: { name: "stable" } };
    const wrapper = shallow(<AppUpgrade {...defaultProps} />);
    wrapper.setProps({ repo });
    expect(wrapper.state("repo")).toEqual(repo);
  });

  it("should request the deployed chart when the app and repo are populated", () => {
    const repo = { metadata: { name: "stable" } };
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
    wrapper.setProps({ repo, app });
    expect(getDeployedChartVersion).toHaveBeenCalledWith("stable/bar", "1.0.0");
  });

  it("should request the deployed chart when the repo is populated later", () => {
    const repo = { metadata: { name: "stable" } };
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
    wrapper.setProps({ repo });
    expect(getDeployedChartVersion).toHaveBeenCalledWith("stable/bar", "1.0.0");
  });

  it("a new app should re-trigger the deployed chart retrieval", () => {
    const repo = { metadata: { name: "stable" } };
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
    wrapper.setProps({ repo, app });
    expect(getDeployedChartVersion).toHaveBeenCalledWith("stable/bar", "1.0.0");

    const app2 = {
      chart: {
        metadata: {
          name: "foobar",
          version: "1.0.0",
        },
      },
    } as IRelease;
    wrapper.setProps({ app: app2 });
    expect(getDeployedChartVersion).toHaveBeenCalledWith("stable/foobar", "1.0.0");
  });
});
