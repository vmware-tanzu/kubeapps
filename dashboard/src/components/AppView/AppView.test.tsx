// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { act } from "@testing-library/react";
import actions from "actions";
import AlertGroup from "components/AlertGroup";
import ApplicationStatus from "components/ApplicationStatus/ApplicationStatus";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import PageHeader from "components/PageHeader";
import {
  AvailablePackageDetail,
  AvailablePackageReference,
  Context,
  GetAvailablePackageDetailResponse,
  GetInstalledPackageDetailResponse,
  GetInstalledPackageResourceRefsResponse,
  InstalledPackageDetail,
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
  PackageAppVersion,
  ResourceRef,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins_pb";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { IConfigState } from "reducers/config";
import { InstalledPackage } from "shared/InstalledPackage";
import PackagesService from "shared/PackagesService";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import {
  CustomInstalledPackageDetail,
  DeleteError,
  FetchError,
  IInstalledPackageState,
  IStoreState,
  PluginNames,
} from "shared/types";
import { getType } from "typesafe-actions";
import AccessURLTable from "./AccessURLTable/AccessURLTable";
import DeleteButton from "./AppControls/DeleteButton/DeleteButton";
import RollbackButton from "./AppControls/RollbackButton";
import UpgradeButton from "./AppControls/UpgradeButton/UpgradeButton";
import AppNotes from "./AppNotes/AppNotes";
import AppView from "./AppView";
import CustomAppView from "./CustomAppView";
import PackageInfo from "./PackageInfo/PackageInfo";
import ResourceTabs from "./ResourceTabs";

const routeParams = {
  cluster: "cluster-1",
  namespace: "default",
  releaseName: "mr-sunshine",
  plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
};
const routePathParam = `/c/${routeParams.cluster}/ns/${routeParams.namespace}/apps/${routeParams.plugin.name}/${routeParams.plugin.version}/${routeParams.releaseName}`;
const routePath = "/c/:cluster/ns/:namespace/apps/:pluginName/:pluginVersion/:releaseName";

const installedPackage = new InstalledPackageDetail({
  name: "test",
  postInstallationNotes: "test",
  valuesApplied: "test",
  availablePackageRef: {
    identifier: "apache/1",
    plugin: { name: PluginNames.PACKAGES_HELM },
    context: { cluster: "", namespace: "chart-namespace" } as Context,
  } as AvailablePackageReference,
  currentVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
  installedPackageRef: new InstalledPackageReference({
    identifier: "apache/1",
    context: { cluster: "", namespace: "package-namespace" } as Context,
    plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
  }),
  latestMatchingVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
  latestVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
  pkgVersionReference: { version: "1" } as VersionReference,
  reconciliationOptions: {},
  status: {
    ready: true,
    reason: InstalledPackageStatus_StatusReason.INSTALLED,
    userReason: "deployed",
  } as InstalledPackageStatus,
});

const availablePackageDetail = new AvailablePackageDetail({
  displayName: "my-cool-package-1",
  availablePackageRef: {
    identifier: "apache/1",
    plugin: { name: PluginNames.PACKAGES_HELM },
    context: { cluster: "", namespace: "chart-namespace" } as Context,
  } as AvailablePackageReference,
  version: { appVersion: "4.5.6", pkgVersion: "1.2.3" },
});

const resourceRefs = {
  configMap: new ResourceRef({ apiVersion: "v1", kind: "ConfigMap", name: "cm-one" }),
  deployment: new ResourceRef({
    apiVersion: "apps/v1",
    kind: "Deployment",
    name: "deployment-one",
  }),
  service: new ResourceRef({ apiVersion: "v1", kind: "Service", name: "svc-one" }),
  ingress: new ResourceRef({
    apiVersion: "extensions/v1",
    kind: "Ingress",
    name: "ingress-one",
  }),
  secret: new ResourceRef({
    apiVersion: "v1",
    kind: "Secret",
    name: "secret-one",
  }),
  daemonset: new ResourceRef({
    apiVersion: "apps/v1",
    kind: "DaemonSet",
    name: "daemonset-one",
  }),
  statefulset: new ResourceRef({
    apiVersion: "apps/v1",
    kind: "StatefulSet",
    name: "statefulset-one",
  }),
};

const validState = {
  apps: {
    isFetching: false,
    items: [installedPackage],
    selected: new CustomInstalledPackageDetail(1, {
      ...installedPackage,
    }),
  } as IInstalledPackageState,
};

beforeEach(() => {
  InstalledPackage.GetInstalledPackageResourceRefs = jest
    .fn()
    .mockReturnValue(
      Promise.resolve(new GetInstalledPackageResourceRefsResponse({ resourceRefs: [] })),
    );
  InstalledPackage.GetInstalledPackageDetail = jest.fn().mockReturnValue(
    Promise.resolve({
      installedPackageDetail: {},
    } as GetInstalledPackageDetailResponse),
  );
  PackagesService.getAvailablePackageDetail = jest.fn().mockReturnValue(
    Promise.resolve({
      availablePackageDetail: {},
    } as GetAvailablePackageDetailResponse),
  );

  // mock the window.matchMedia for selecting the theme
  Object.defineProperty(window, "matchMedia", {
    writable: true,
    configurable: true,
    value: jest.fn().mockImplementation(query => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: jest.fn(),
      removeListener: jest.fn(),
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
      dispatchEvent: jest.fn(),
    })),
  });

  // mock the window.ResizeObserver, required by the MonacoEditor for the layout
  Object.defineProperty(window, "ResizeObserver", {
    writable: true,
    configurable: true,
    value: jest.fn().mockImplementation(() => ({
      observe: jest.fn(),
      unobserve: jest.fn(),
      disconnect: jest.fn(),
    })),
  });

  // mock the window.HTMLCanvasElement.getContext(), required by the MonacoEditor for the layout
  Object.defineProperty(HTMLCanvasElement.prototype, "getContext", {
    writable: true,
    configurable: true,
    value: jest.fn().mockImplementation(() => ({
      clearRect: jest.fn(),
      fillRect: jest.fn(),
    })),
  });
});
afterEach(() => {
  jest.resetAllMocks();
});

describe("AppView", () => {
  it("renders a loading wrapper", async () => {
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          apps: {
            selected: undefined,
            isFetching: true,
          } as IInstalledPackageState,
        } as Partial<IStoreState>),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Routes>
            <Route path={routePath} element={<AppView />} />
          </Routes>
        </MemoryRouter>,
        false,
      );
    });
    expect(wrapper.find(LoadingWrapper).prop("loaded")).toBe(false);
  });

  it("does not render an loading wrapper if it isn't fetching", async () => {
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          apps: {
            selected: undefined,
            isFetching: false,
            error: new Error("foo not found"),
          } as IInstalledPackageState,
        } as Partial<IStoreState>),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Routes>
            <Route path={routePath} element={<AppView />} />
          </Routes>
        </MemoryRouter>,
        false,
      );
    });
    expect(wrapper.find(LoadingWrapper).prop("loaded")).toBe(true);
    expect(wrapper.find(AlertGroup).html()).toContain("foo not found");
    expect(wrapper.find(PageHeader)).not.toExist();
  });

  it("renders a fetch error only", async () => {
    let wrapper: any;

    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          apps: { error: new FetchError("boom!") } as IInstalledPackageState,
        } as Partial<IStoreState>),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Routes>
            <Route path={routePath} element={<AppView />} />
          </Routes>
        </MemoryRouter>,
        false,
      );
    });
    expect(wrapper.find(AlertGroup)).toExist();
    expect(wrapper.find(PageHeader)).not.toExist();
  });

  it("renders a custom component when package is in customAppViews", async () => {
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          apps: { selected: { ...installedPackage } } as IInstalledPackageState,
          config: {
            customAppViews: [
              {
                name: "1",
                plugin: PluginNames.PACKAGES_HELM,
                repository: "apache",
              },
            ],
          } as IConfigState,
        } as Partial<IStoreState>),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Routes>
            <Route path={routePath} element={<AppView />} />
          </Routes>
        </MemoryRouter>,
        false,
      );
    });
    expect(wrapper.find(CustomAppView)).toExist();
  });

  it("does not render a custom component when package is not in customAppViews", async () => {
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          apps: { selected: { ...installedPackage } } as IInstalledPackageState,
          config: {
            customAppViews: [
              {
                name: "demo-chart",
                plugin: PluginNames.PACKAGES_HELM,
                repository: "demo-repo",
              },
            ],
          } as IConfigState,
        } as Partial<IStoreState>),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Routes>
            <Route path={routePath} element={<AppView />} />
          </Routes>
        </MemoryRouter>,
        false,
      );
    });
    expect(wrapper.find(CustomAppView)).not.toExist();
  });

  it("renders a PackageHeader with the information about the associated available package", async () => {
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          apps: {
            selected: { ...installedPackage },
            selectedDetails: { ...availablePackageDetail },
          } as IInstalledPackageState,
        } as Partial<IStoreState>),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Routes>
            <Route path={routePath} element={<AppView />} />
          </Routes>
        </MemoryRouter>,
        false,
      );
    });
    expect(wrapper.find(PageHeader).text()).toContain("from package my-cool-package-1");
  });

  it("renders a RollBack button if the installedPackage is from PACKAGES_HELM", async () => {
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          apps: {
            selected: {
              ...installedPackage,
              installedPackageRef: {
                ...installedPackage.installedPackageRef,
                plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" } as Plugin,
              } as InstalledPackageReference,
            },
          } as IInstalledPackageState,
        } as Partial<IStoreState>),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Routes>
            <Route path={routePath} element={<AppView />} />
          </Routes>
        </MemoryRouter>,
        false,
      );
    });

    expect(wrapper.find(UpgradeButton)).toExist();
    expect(wrapper.find(RollbackButton)).toExist();
    expect(wrapper.find(DeleteButton)).toExist();
  });

  it("does not render a RollBack button if the installedPackage is not from PACKAGES_HELM", async () => {
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          apps: { selected: { ...installedPackage } } as IInstalledPackageState,
        } as Partial<IStoreState>),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Routes>
            <Route path={routePath} element={<AppView />} />
          </Routes>
        </MemoryRouter>,
        false,
      );
    });
    expect(wrapper.find(UpgradeButton)).toExist();
    expect(wrapper.find(RollbackButton)).not.toExist();
    expect(wrapper.find(DeleteButton)).toExist();
  });

  describe("State initialization", () => {
    /*
      The imported resource refs contain one deployment, one service, one config map and some bogus manifests.
    */
    it("sets ResourceRefs for its deployments, services, ingresses and secrets", async () => {
      const apiResourceRefs = [
        resourceRefs.deployment,
        resourceRefs.service,
        resourceRefs.configMap,
        resourceRefs.ingress,
        resourceRefs.secret,
      ] as ResourceRef[];
      InstalledPackage.GetInstalledPackageResourceRefs = jest.fn().mockReturnValue(
        Promise.resolve({
          resourceRefs: apiResourceRefs,
        } as GetInstalledPackageResourceRefsResponse),
      );

      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({
            apps: { selected: installedPackage } as IInstalledPackageState,
          } as Partial<IStoreState>),
          <MemoryRouter initialEntries={[routePathParam]}>
            <Routes>
              <Route path={routePath} element={<AppView />} />
            </Routes>
          </MemoryRouter>,
          false,
        );
      });
      wrapper.update();

      const tabs = wrapper.find(ResourceTabs);
      expect(tabs.prop("deployments")).toEqual([resourceRefs.deployment]);
      expect(tabs.prop("services")).toEqual([resourceRefs.service]);
      expect(tabs.prop("secrets")).toEqual([resourceRefs.secret]);
    });

    it("stores other k8s resources", async () => {
      const apiResourceRefs = [
        resourceRefs.deployment,
        resourceRefs.service,
        resourceRefs.configMap,
        resourceRefs.secret,
      ] as ResourceRef[];
      InstalledPackage.GetInstalledPackageResourceRefs = jest.fn().mockReturnValue(
        Promise.resolve({
          resourceRefs: apiResourceRefs,
        } as GetInstalledPackageResourceRefsResponse),
      );

      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({
            apps: { selected: installedPackage } as IInstalledPackageState,
          } as Partial<IStoreState>),
          <MemoryRouter initialEntries={[routePathParam]}>
            <Routes>
              <Route path={routePath} element={<AppView />} />
            </Routes>
          </MemoryRouter>,
          false,
        );
      });
      wrapper.update();

      const tabs = wrapper.find(ResourceTabs);
      const otherResources: ResourceRef[] = tabs.prop("otherResources");
      const configMap = otherResources[0];
      // It should skip deployments, services and secrets from "other resources"
      expect(otherResources.length).toEqual(1);

      expect(configMap).toBeDefined();
      expect(configMap.name).toEqual("cm-one");
    });
  });

  describe("renderization", () => {
    it("renders all the elements of an application", async () => {
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(getStore(validState), <AppView />);
      });
      expect(wrapper.find(PackageInfo)).toExist();
      expect(wrapper.find(ApplicationStatus)).toExist();
      expect(wrapper.find(".control-buttons")).toExist();
      expect(wrapper.find(AppNotes)).toExist();
      expect(wrapper.find(ResourceTabs)).toExist();
      expect(wrapper.find(AccessURLTable)).toExist();
    });

    it("renders an error if error prop is set", async () => {
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({
            ...validState,
            apps: { ...validState.apps, error: new Error("Boom!") } as IInstalledPackageState,
          } as Partial<IStoreState>),
          <MemoryRouter initialEntries={[routePathParam]}>
            <Routes>
              <Route path={routePath} element={<AppView />} />
            </Routes>
          </MemoryRouter>,
          false,
        );
      });
      const err = wrapper.find(AlertGroup);
      expect(err).toExist();
      expect(err.html()).toContain("Boom!");
    });

    it("renders a delete-error", async () => {
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({
            ...validState,
            apps: { ...validState.apps, error: new DeleteError("Boom!") } as IInstalledPackageState,
          } as Partial<IStoreState>),
          <MemoryRouter initialEntries={[routePathParam]}>
            <Routes>
              <Route path={routePath} element={<AppView />} />
            </Routes>
          </MemoryRouter>,
          false,
        );
      });
      const err = wrapper.find(AlertGroup);
      expect(err).toExist();
      expect(err.html()).toContain("Unable to delete the application. Received: Boom!");
    });
  });

  it("forwards statefulsets and daemonsets to the application status", async () => {
    const apiResourceRefs = [resourceRefs.statefulset, resourceRefs.daemonset] as ResourceRef[];
    InstalledPackage.GetInstalledPackageResourceRefs = jest.fn().mockReturnValue(
      Promise.resolve({
        resourceRefs: apiResourceRefs,
      } as GetInstalledPackageResourceRefsResponse),
    );
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          apps: { selected: installedPackage } as IInstalledPackageState,
        } as Partial<IStoreState>),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Routes>
            <Route path={routePath} element={<AppView />} />
          </Routes>
        </MemoryRouter>,
        false,
      );
    });
    wrapper.update();

    const applicationStatus = wrapper.find(ApplicationStatus);
    expect(applicationStatus).toExist();

    expect(applicationStatus.prop("statefulsetRefs")).toEqual([resourceRefs.statefulset]);
    expect(applicationStatus.prop("daemonsetRefs")).toEqual([resourceRefs.daemonset]);
  });
});

describe("AppView actions", () => {
  it("watches certain resources and gets others when mounted", async () => {
    const apiResourceRefs = [
      resourceRefs.deployment,
      resourceRefs.service,
      resourceRefs.secret,
    ] as ResourceRef[];
    InstalledPackage.GetInstalledPackageResourceRefs = jest.fn().mockReturnValue(
      Promise.resolve({
        resourceRefs: apiResourceRefs,
      } as GetInstalledPackageResourceRefsResponse),
    );
    const store = getStore({
      apps: { selected: installedPackage } as IInstalledPackageState,
    } as Partial<IStoreState>);

    await act(async () => {
      mountWrapper(
        store,
        <MemoryRouter initialEntries={[routePathParam]}>
          <Routes>
            <Route path={routePath} element={<AppView />} />
          </Routes>
        </MemoryRouter>,
        false,
      );
    });

    expect(store.getActions()).toEqual([
      {
        type: getType(actions.installedpackages.requestInstalledPackage),
      },
      {
        type: getType(actions.installedpackages.selectInstalledPackage),
        payload: {
          pkg: {},
          details: {},
        },
      },
      {
        type: getType(actions.kube.requestResources),
        payload: {
          pkg: installedPackage.installedPackageRef,
          refs: [resourceRefs.secret],
          watch: false,
          handler: expect.any(Function),
          onError: expect.any(Function),
        },
      },
      {
        type: getType(actions.kube.requestResources),
        payload: {
          pkg: installedPackage.installedPackageRef,
          refs: [resourceRefs.deployment, resourceRefs.service],
          watch: true,
          handler: expect.any(Function),
          onError: expect.any(Function),
        },
      },
    ]);
  });
  it("closes the watches when unmounted", async () => {
    const apiResourceRefs = [resourceRefs.deployment, resourceRefs.service] as ResourceRef[];
    InstalledPackage.GetInstalledPackageResourceRefs = jest.fn().mockReturnValue(
      Promise.resolve({
        resourceRefs: apiResourceRefs,
      } as GetInstalledPackageResourceRefsResponse),
    );

    const store = getStore({
      apps: { selected: installedPackage } as IInstalledPackageState,
    } as Partial<IStoreState>);
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        store,
        <MemoryRouter initialEntries={[routePathParam]}>
          <Routes>
            <Route path={routePath} element={<AppView />} />
          </Routes>
        </MemoryRouter>,
        false,
      );
    });
    await act(async () => {
      wrapper.unmount();
    });

    const watch = true;
    expect(store.getActions()).toEqual([
      {
        type: getType(actions.installedpackages.requestInstalledPackage),
      },
      {
        type: getType(actions.installedpackages.selectInstalledPackage),
        payload: {
          pkg: {},
          details: {},
        },
      },
      {
        type: getType(actions.kube.requestResources),
        payload: {
          pkg: installedPackage.installedPackageRef,
          refs: apiResourceRefs,
          watch,
          handler: expect.any(Function),
          onError: expect.any(Function),
        },
      },
    ]);
  });
});
