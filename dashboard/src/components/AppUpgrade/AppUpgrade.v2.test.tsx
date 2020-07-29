import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import Alert from "components/js/Alert";
import { hapi } from "shared/hapi/release";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { IAppRepository, IChartState, IChartVersion, IRelease } from "shared/types";
import itBehavesLike from "../../shared/specs";
import SelectRepoForm from "../SelectRepoForm/SelectRepoForm.v2";
import UpgradeForm from "../UpgradeForm/UpgradeForm.v2";
import AppUpgrade, { IAppUpgradeProps } from "./AppUpgrade.v2";

const versions = [
  {
    id: "foo",
    attributes: { version: "1.2.3" },
    relationships: { chart: { data: { repo: { name: "bitnami" } } } },
  },
  {
    id: "foo",
    attributes: { version: "1.2.4" },
    relationships: { chart: { data: { repo: { name: "bitnami" } } } },
  },
] as IChartVersion[];
const schema = { properties: { foo: { type: "string" } } };

const defaultProps = {
  app: {} as hapi.release.Release,
  appsIsFetching: false,
  chartsIsFetching: false,
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
  cluster: "default",
  push: jest.fn(),
  goBack: jest.fn(),
  releaseName: "foo",
  repo: {} as IAppRepository,
  repoError: undefined,
  repos: [],
  selected: { versions, version: versions[0], schema },
  upgradeApp: jest.fn(),
  version: "1.0.0",
} as IAppUpgradeProps;

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
  expect(wrapper.find(Alert)).not.toExist();
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
        appsError={new Error("foo does not exist")}
        repos={[repo]}
        repo={repo}
      />,
    );

    expect(wrapper.find(Alert)).toExist();
    expect(wrapper.find(SelectRepoForm)).not.toExist();
    expect(wrapper.find(UpgradeForm)).not.toExist();

    expect(wrapper.html()).toContain("foo does not exist");
  });

  it("renders a warning message if there are no repositories", () => {
    const wrapper = mountWrapper(
      defaultStore,
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
      />,
    );

    expect(wrapper.find(SelectRepoForm).find(Alert)).toExist();
    expect(wrapper.find(UpgradeForm)).not.toExist();

    expect(
      wrapper
        .find(Alert)
        .children()
        .text(),
    ).toContain("Chart repositories not found");
  });
});

it("renders the upgrade form when the repo is available", () => {
  const wrapper = mountWrapper(
    defaultStore,
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
      repoName="foobar"
    />,
  );
  expect(wrapper.find(UpgradeForm)).toExist();
  expect(wrapper.find(Alert)).not.toExist();
  expect(wrapper.find(SelectRepoForm)).not.toExist();
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
  const wrapper = mountWrapper(
    defaultStore,
    <AppUpgrade {...defaultProps} repos={[repo]} app={app} />,
  );
  expect(wrapper.find(UpgradeForm)).toExist();
  expect(wrapper.find(Alert)).not.toExist();
  expect(wrapper.find(SelectRepoForm)).not.toExist();
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
    mountWrapper(
      defaultStore,
      <AppUpgrade
        {...defaultProps}
        getDeployedChartVersion={getDeployedChartVersion}
        repoName="stable"
        app={app}
      />,
    );
    expect(getDeployedChartVersion).toHaveBeenCalledWith(
      defaultProps.repoNamespace,
      "stable/bar",
      "1.0.0",
    );
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
    mountWrapper(
      defaultStore,
      <AppUpgrade
        {...defaultProps}
        app={app}
        getDeployedChartVersion={getDeployedChartVersion}
        repoName="stable"
      />,
    );
    expect(getDeployedChartVersion).toHaveBeenCalledWith(
      defaultProps.repoNamespace,
      "stable/bar",
      "1.0.0",
    );
  });
});
