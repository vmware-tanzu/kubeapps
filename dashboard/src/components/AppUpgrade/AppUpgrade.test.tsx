import { shallow } from "enzyme";
import context from "jest-plugin-context";

import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper";
import { hapi } from "shared/hapi/release";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import {
  FetchError,
  IAppRepository,
  IChartState,
  IChartVersion,
  IRelease,
  UpgradeError,
} from "shared/types";
import SelectRepoForm from "../SelectRepoForm/SelectRepoForm";
import UpgradeForm from "../UpgradeForm/UpgradeForm";
import AppUpgrade, { IAppUpgradeProps } from "./AppUpgrade";

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

it("renders the repo selection form if not introduced", () => {
  const wrapper = shallow(<AppUpgrade {...defaultProps} appsIsFetching={true} />);
  expect(wrapper.find(LoadingWrapper).prop("loaded")).toBe(false);
});

it("renders the repo selection form if not introduced when the app is loaded", () => {
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
});

context("when an error exists", () => {
  it("renders a generic error message", () => {
    const repo = {
      metadata: { name: "stable" },
    } as IAppRepository;
    const wrapper = shallow(
      <AppUpgrade
        {...defaultProps}
        error={new FetchError("foo does not exist")}
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

    expect(wrapper.find(Alert).children().text()).toContain("Chart repositories not found");
  });

  it("still renders the upgrade form even if there is an upgrade error", () => {
    const repo = {
      metadata: { name: "stable" },
    } as IAppRepository;
    const upgradeError = new UpgradeError("foo upgrade failed");
    const wrapper = shallow(
      <AppUpgrade
        {...defaultProps}
        error={upgradeError}
        repos={[repo]}
        repo={repo}
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
    expect(wrapper.find(UpgradeForm).prop("error")).toEqual(upgradeError);
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
      defaultProps.cluster,
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
      defaultProps.cluster,
      defaultProps.repoNamespace,
      "stable/bar",
      "1.0.0",
    );
  });
});
