import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper";
import {
  AvailablePackageDetail,
  AvailablePackageReference,
  Context,
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
  InstalledPackageSummary,
  PackageAppVersion,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { IAppRepositoryState } from "reducers/repos";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import {
  CustomInstalledPackageDetail,
  FetchError,
  IAppRepository,
  IAppState,
  IChartState,
  UpgradeError,
} from "shared/types";
import SelectRepoForm from "../SelectRepoForm/SelectRepoForm";
import UpgradeForm from "../UpgradeForm/UpgradeForm";
import AppUpgrade from "./AppUpgrade";

const installedPackageSummary1 = {} as InstalledPackageSummary;
const availablePackageDetail1 = {} as AvailablePackageDetail;

const installedPackage1 = {
  name: "test",
  postInstallationNotes: "test",
  valuesApplied: "test",
  availablePackageRef: {
    identifier: "apache/1",
    context: { cluster: "", namespace: "chart-namespace" } as Context,
  } as AvailablePackageReference,
  currentVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
  installedPackageRef: {
    identifier: "apache/1",
    pkgVersion: "1.0.0",
    context: { cluster: "", namespace: "chart-namespace" } as Context,
  } as InstalledPackageReference,
  latestMatchingVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
  latestVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
  pkgVersionReference: { version: "1" } as VersionReference,
  reconciliationOptions: {},
  status: {
    ready: true,
    reason: InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED,
    userReason: "deployed",
  } as InstalledPackageStatus,
} as CustomInstalledPackageDetail;

const installedPackage2 = {
  name: "test",
  postInstallationNotes: "test",
  valuesApplied: "test",
  availablePackageRef: {
    identifier: "apache/1",
    context: { cluster: "", namespace: "chart-namespace" } as Context,
  } as AvailablePackageReference,
  currentVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
  installedPackageRef: {
    identifier: "apache/1",
    pkgVersion: "1.0.0",
    context: { cluster: "", namespace: "chart-namespace" } as Context,
  } as InstalledPackageReference,
  latestMatchingVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
  latestVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
  pkgVersionReference: { version: "1" } as VersionReference,
  reconciliationOptions: {},
  status: {
    ready: true,
    reason: InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED,
    userReason: "deployed",
  } as InstalledPackageStatus,
} as CustomInstalledPackageDetail;

const repo1 = { metadata: { name: "stable", namespace: "default" } } as IAppRepository;

let spyOnUseDispatch: jest.SpyInstance;

beforeEach(() => {
  jest.resetAllMocks();
});

const FULL_STATE = {
  apps: {
    isFetching: false,
    error: undefined,
    items: [installedPackage1],
    listOverview: [installedPackageSummary1],
    selected: installedPackage1,
    selectedDetails: availablePackageDetail1,
  } as IAppState,
  repos: {
    repo: repo1,
    repos: [repo1],
    isFetching: false,
  } as IAppRepositoryState,
  charts: {
    isFetching: false,
    selected: {
      versions: [{ appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion],
      availablePackageDetail: { name: "test" } as AvailablePackageDetail,
      pkgVersion: "",
    },
  } as IChartState,
};

it("renders the repo selection form if not introduced", () => {
  const state = {
    apps: {
      isFetching: true,
    } as IAppState,
  };
  const wrapper = mountWrapper(
    getStore({ ...defaultStore, apps: { ...state.apps } }),
    <AppUpgrade />,
  );
  expect(wrapper.find(LoadingWrapper).prop("loaded")).toBe(false);
});

it("renders the repo selection form if not introduced when the app is loaded", () => {
  const state = {
    repos: {
      repos: [repo1],
    } as IAppRepositoryState,
  };
  const wrapper = mountWrapper(
    getStore({
      ...defaultStore,
      repos: { ...state.repos },
    }),
    <AppUpgrade />,
  );
  expect(wrapper.find(SelectRepoForm)).toExist();
  expect(wrapper.find(Alert)).not.toExist();
  expect(wrapper.find(UpgradeForm)).not.toExist();
});

describe("when an error exists", () => {
  it("renders a generic error message", () => {
    const state = {
      apps: {
        error: new FetchError("foo does not exist"),
      } as IAppState,
    };
    const wrapper = mountWrapper(
      getStore({
        ...defaultStore,
        apps: { ...state.apps },
      }),
      <AppUpgrade />,
    );

    expect(wrapper.find(Alert)).toExist();
    expect(wrapper.find(SelectRepoForm)).not.toExist();
    expect(wrapper.find(UpgradeForm)).not.toExist();

    expect(wrapper.html()).toContain("foo does not exist");
  });

  it("renders a warning message if there are no repositories", () => {
    const state = {
      repos: {
        repos: [] as IAppRepository[],
      } as IAppRepositoryState,
    };
    const wrapper = mountWrapper(
      getStore({
        ...defaultStore,
        repos: { ...state.repos },
      }),
      <AppUpgrade />,
    );
    expect(wrapper.find(SelectRepoForm).find(Alert)).toExist();
    expect(wrapper.find(UpgradeForm)).not.toExist();

    expect(wrapper.find(Alert).children().text()).toContain("Chart repositories not found");
  });

  it("still renders the upgrade form even if there is an upgrade error", () => {
    const upgradeError = new UpgradeError("foo upgrade failed");
    const state = {
      apps: {
        error: upgradeError,
        selected: installedPackage1,
      } as IAppState,
    };
    const wrapper = mountWrapper(
      getStore({
        ...defaultStore,
        apps: { ...state.apps },
      }),
      <AppUpgrade />,
    );
    expect(wrapper.find(UpgradeForm)).toExist();
    expect(wrapper.find(UpgradeForm).prop("error")).toEqual(upgradeError);
  });
});

it("renders the upgrade form when the repo is available", () => {
  const state = {
    apps: {
      selected: installedPackage1,
    } as IAppState,
    repos: {
      repo: repo1,
      repos: [repo1],
      isFetching: false,
    } as IAppRepositoryState,
  };
  const wrapper = mountWrapper(
    getStore({
      ...defaultStore,
      apps: { ...state.apps },
      repos: { ...state.repos },
    }),
    <AppUpgrade />,
  );
  expect(wrapper.find(UpgradeForm)).toExist();
  expect(wrapper.find(Alert)).not.toExist();
  expect(wrapper.find(SelectRepoForm)).not.toExist();
});

it("skips the repo selection form if the app contains upgrade info", () => {
  const state = {
    apps: {
      selected: installedPackage1,
    } as IAppState,
    repos: {
      repo: repo1,
      repos: [repo1],
      isFetching: false,
    } as IAppRepositoryState,
  };
  const wrapper = mountWrapper(
    getStore({
      ...defaultStore,
      apps: { ...state.apps },
      repos: { ...state.repos },
    }),
    <AppUpgrade />,
  );
  expect(wrapper.find(UpgradeForm)).toExist();
  expect(wrapper.find(Alert)).not.toExist();
  expect(wrapper.find(SelectRepoForm)).not.toExist();
});

// describe("when receiving new props", () => {

// TODO(agamez): Test temporarily commented out
//   it("should request the deployed chart when the app and repo are populated", () => {
//     const app = {
//       chart: {
//         metadata: {
//           name: "bar",
//           version: "1.0.0",
//         },
//       },
//     } as IRelease;
//     const getDeployedChartVersion = jest.fn();
//     mountWrapper(
//       defaultStore,
//       <AppUpgrade
//         {...defaultProps}
//         getDeployedChartVersion={getDeployedChartVersion}
//         repoName="stable"
//         app={app}
//       />,
//     );
//     expect(getDeployedChartVersion).toHaveBeenCalledWith(
//       defaultProps.cluster,
//       defaultProps.repoNamespace,
//       "stable/bar",
//       "1.0.0",
//     );
//   });

// TODO(agamez): Test temporarily commented out
//   it("should request the deployed chart when the repo is populated later", () => {
//     const app = {
//       chart: {
//         metadata: {
//           name: "bar",
//           version: "1.0.0",
//         },
//       },
//     } as IRelease;
//     const getDeployedChartVersion = jest.fn();
//     mountWrapper(
//       defaultStore,
//       <AppUpgrade
//         {...defaultProps}
//         app={app}
//         getDeployedChartVersion={getDeployedChartVersion}
//         repoName="stable"
//       />,
//     );
//     expect(getDeployedChartVersion).toHaveBeenCalledWith(
//       defaultProps.cluster,
//       defaultProps.repoNamespace,
//       "stable/bar",
//       "1.0.0",
//     );
//   });
// });
