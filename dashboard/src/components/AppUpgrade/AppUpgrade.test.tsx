import actions from "actions";
import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper";
import {
  AvailablePackageDetail,
  AvailablePackageReference,
  Context,
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
  PackageAppVersion,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import * as ReactRedux from "react-redux";
import * as ReactRouter from "react-router";
import { MemoryRouter, Route } from "react-router-dom";
import { IAppRepositoryState } from "reducers/repos";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import {
  CustomInstalledPackageDetail,
  FetchError,
  IAppRepository,
  IAppState,
  IPackageState,
  UpgradeError,
} from "shared/types";
import { PluginNames } from "shared/utils";
import SelectRepoForm from "../SelectRepoForm/SelectRepoForm";
import UpgradeForm from "../UpgradeForm/UpgradeForm";
import AppUpgrade from "./AppUpgrade";

const defaultProps = {
  pkgName: "foo",
  cluster: "default",
  namespace: "default",
  repoNamespace: "stable",
  repo: "repo",
  releaseName: "my-release",
  plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
};

const installedPackage1 = {
  name: "test",
  postInstallationNotes: "test",
  valuesApplied: "test",
  availablePackageRef: {
    identifier: "stable/bar",
    context: { cluster: defaultProps.cluster, namespace: defaultProps.repoNamespace } as Context,
    plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
  } as AvailablePackageReference,
  currentVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
  installedPackageRef: {
    identifier: "stable/bar",
    pkgVersion: "1.0.0",
    context: { cluster: defaultProps.cluster, namespace: defaultProps.repoNamespace } as Context,
    plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
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

const availablePackageDetail = {
  availablePackageRef: {
    context: { cluster: "default", namespace: "my-ns" },
    identifier: "test",
    plugin: { name: PluginNames.PACKAGES_HELM, version: "0.0.1" } as Plugin,
  },
  version: { appVersion: "4.5.6", pkgVersion: "1.2.3" },
} as AvailablePackageDetail;

const selectedPackage = {
  versions: [{ appVersion: "10.0.0", pkgVersion: "1.2.3" }],
  availablePackageDetail: { name: "test" } as AvailablePackageDetail,
} as IPackageState["selected"];

const repo1 = {
  metadata: {
    name: defaultProps.repo,
    namespace: defaultProps.repoNamespace,
  },
} as IAppRepository;

let spyOnUseDispatch: jest.SpyInstance;
let spyOnUseHistory: jest.SpyInstance;

beforeEach(() => {
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  spyOnUseHistory = jest
    .spyOn(ReactRouter, "useHistory")
    .mockReturnValue({ push: jest.fn() } as any);
});

afterEach(() => {
  jest.restoreAllMocks();
  spyOnUseDispatch.mockRestore();
  spyOnUseHistory.mockRestore();
});

const routePathParam = `/c/${defaultProps.cluster}/ns/${defaultProps.namespace}/apps/${defaultProps.plugin.name}/${defaultProps.plugin.version}/${defaultProps.releaseName}/upgrade`;
const routePath = "/c/:cluster/ns/:namespace/apps/:pluginName/:pluginVersion/:releaseName/upgrade";

it("renders the repo selection form if not introduced", () => {
  const state = {
    apps: {
      isFetching: true,
    } as IAppState,
  };
  const wrapper = mountWrapper(
    getStore({ ...defaultStore, ...state }),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <AppUpgrade />,
      </Route>
    </MemoryRouter>,
  );
  expect(wrapper.find(LoadingWrapper).prop("loaded")).toBe(false);
});

it("renders the repo selection form if not introduced when the app is loaded", () => {
  const state = {
    repos: {
      repos: [repo1],
    } as IAppRepositoryState,
    apps: { selected: { name: "foo" }, isFetching: false, error: undefined } as IAppState,
  };
  const wrapper = mountWrapper(
    getStore({
      ...defaultStore,
      ...state,
    }),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <AppUpgrade />,
      </Route>
    </MemoryRouter>,
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
        ...state,
      }),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppUpgrade />,
        </Route>
      </MemoryRouter>,
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
      apps: { selected: { name: "foo" }, isFetching: false, error: undefined } as IAppState,
    };
    const wrapper = mountWrapper(
      getStore({
        ...defaultStore,
        ...state,
      }),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppUpgrade />,
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(SelectRepoForm).find(Alert)).toExist();
    expect(wrapper.find(UpgradeForm)).not.toExist();

    expect(wrapper.find(Alert).children().text()).toContain("Repositories not found");
  });

  it("still renders the upgrade form even if there is an upgrade error", () => {
    const upgradeError = new UpgradeError("foo upgrade failed");
    const state = {
      apps: {
        error: upgradeError,
        selected: installedPackage1,
        selectedDetails: availablePackageDetail,
      } as IAppState,
      packages: { selected: selectedPackage } as IPackageState,
    };

    const wrapper = mountWrapper(
      getStore({
        ...defaultStore,
        ...state,
      }),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppUpgrade />,
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(UpgradeForm)).toExist();
    expect(wrapper.find(UpgradeForm).find(Alert)).toIncludeText(upgradeError.message);
  });
});

it("renders the upgrade form when the repo is available, clears state and fetches app", () => {
  const getApp = jest.fn();
  actions.apps.getApp = getApp;
  const resetSelectedAvailablePackageDetail = jest
    .spyOn(actions.packages, "resetSelectedAvailablePackageDetail")
    .mockImplementation(jest.fn());

  const state = {
    apps: {
      selected: installedPackage1,
      selectedDetails: availablePackageDetail,
    } as IAppState,
    repos: {
      repo: repo1,
      repos: [repo1],
      isFetching: false,
    } as IAppRepositoryState,
    packages: { selected: selectedPackage } as IPackageState,
  };
  const wrapper = mountWrapper(
    getStore({
      ...defaultStore,
      ...state,
    }),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <AppUpgrade />,
      </Route>
    </MemoryRouter>,
  );
  expect(wrapper.find(UpgradeForm)).toExist();
  expect(wrapper.find(Alert)).not.toExist();
  expect(wrapper.find(SelectRepoForm)).not.toExist();

  expect(resetSelectedAvailablePackageDetail).toHaveBeenCalled();
  expect(getApp).toHaveBeenCalledWith({
    context: { cluster: defaultProps.cluster, namespace: defaultProps.namespace },
    identifier: defaultProps.releaseName,
    plugin: defaultProps.plugin,
  });
});

it("renders the upgrade form with the version property", () => {
  const state = {
    apps: {
      selected: installedPackage1,
      selectedDetails: availablePackageDetail,
    } as IAppState,
    repos: {
      repo: repo1,
      repos: [repo1],
      isFetching: false,
    } as IAppRepositoryState,
    packages: { selected: selectedPackage } as IPackageState,
  };
  const wrapper = mountWrapper(
    getStore({
      ...defaultStore,
      ...state,
    }),
    <MemoryRouter initialEntries={[routePathParam + "/0.0.1"]}>
      <Route path={routePath + "/:version"}>
        <AppUpgrade />,
      </Route>
    </MemoryRouter>,
  );
  expect(wrapper.find(UpgradeForm)).toExist();
  expect(wrapper.find(UpgradeForm)).toHaveProp("version", "0.0.1");
});

it("skips the repo selection form if the app contains upgrade info", () => {
  const state = {
    apps: {
      selected: installedPackage1,
      selectedDetails: availablePackageDetail,
    } as IAppState,
    repos: {
      repo: repo1,
      repos: [repo1],
      isFetching: false,
    } as IAppRepositoryState,
    packages: { selected: selectedPackage } as IPackageState,
  };
  const wrapper = mountWrapper(
    getStore({
      ...defaultStore,
      ...state,
    }),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <AppUpgrade />,
      </Route>
    </MemoryRouter>,
  );
  expect(wrapper.find(UpgradeForm)).toExist();
  expect(wrapper.find(Alert)).not.toExist();
  expect(wrapper.find(SelectRepoForm)).not.toExist();
});
