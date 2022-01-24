import actions from "actions";
import { getType } from "typesafe-actions";
import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import PageHeader from "components/PageHeader";
import ApplicationStatusContainer from "containers/ApplicationStatusContainer";
import {
  AvailablePackageReference,
  Context,
  InstalledPackageDetail,
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
  PackageAppVersion,
  ResourceRef,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { MemoryRouter, Route } from "react-router-dom";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { DeleteError, FetchError } from "shared/types";
import { PluginNames } from "shared/utils";
import AccessURLTable from "./AccessURLTable/AccessURLTable";
import DeleteButton from "./AppControls/DeleteButton/DeleteButton";
import RollbackButton from "./AppControls/RollbackButton";
import UpgradeButton from "./AppControls/UpgradeButton/UpgradeButton";
import AppNotes from "./AppNotes/AppNotes";
import AppView from "./AppView";
import PackageInfo from "./PackageInfo/PackageInfo";
import CustomAppView from "./CustomAppView";
import ResourceTabs from "./ResourceTabs";

const routeParams = {
  cluster: "cluster-1",
  namespace: "default",
  releaseName: "mr-sunshine",
  plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
};
const routePathParam = `/c/${routeParams.cluster}/ns/${routeParams.namespace}/apps/${routeParams.plugin.name}/${routeParams.plugin.version}/${routeParams.releaseName}`;
const routePath = "/c/:cluster/ns/:namespace/apps/:pluginName/:pluginVersion/:releaseName";

const installedPackage = {
  name: "test",
  postInstallationNotes: "test",
  valuesApplied: "test",
  availablePackageRef: {
    identifier: "apache/1",
    plugin: { name: PluginNames.PACKAGES_HELM },
    context: { cluster: "", namespace: "chart-namespace" } as Context,
  } as AvailablePackageReference,
  currentVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
  installedPackageRef: {
    identifier: "apache/1",
    pkgVersion: "1.0.0",
    context: { cluster: "", namespace: "package-namespace" } as Context,
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
} as InstalledPackageDetail;

const resourceRefs = {
  configMap: { apiVersion: "v1", kind: "ConfigMap", name: "cm-one" } as ResourceRef,
  deployment: {
    apiVersion: "apps/v1beta1",
    kind: "Deployment",
    name: "deployment-one",
  } as ResourceRef,
  service: { apiVersion: "v1", kind: "Service", name: "svc-one" } as ResourceRef,
  ingress: {
    apiVersion: "extensions/v1beta1",
    kind: "Ingress",
    name: "ingress-one",
  } as ResourceRef,
  secret: {
    apiVersion: "v1",
    kind: "Secret",
    name: "secret-one",
  } as ResourceRef,
  daemonset: {
    apiVersion: "apps/v1beta1",
    kind: "DaemonSet",
    name: "daemonset-one",
  } as ResourceRef,
  statefulset: {
    apiVersion: "apps/v1beta1",
    kind: "StatefulSet",
    name: "statefulset-one",
  } as ResourceRef,
};

const validState = {
  apps: {
    selected: {
      ...installedPackage,
      resourceRefs: [resourceRefs.configMap] as ResourceRef[],
    },
  },
};

describe("AppView", () => {
  it("renders a loading wrapper", () => {
    const wrapper = mountWrapper(defaultStore, <AppView />);
    expect(wrapper.find(LoadingWrapper)).toExist();
  });

  it("renders a fetch error only", () => {
    const wrapper = mountWrapper(
      getStore({ apps: { error: new FetchError("boom!") } }),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppView />
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(Alert)).toExist();
    expect(wrapper.find(PageHeader)).not.toExist();
  });

  it("renders a custom component when package is in customAppViews", () => {
    const wrapper = mountWrapper(
      getStore({
        apps: { selected: { ...installedPackage } },
        config: {
          customAppViews: [
            {
              name: "1",
              plugin: PluginNames.PACKAGES_HELM,
              repository: "apache",
            },
          ],
        },
      }),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppView />
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(CustomAppView)).toExist();
  });

  it("does not render a custom component when package is not in customAppViews", () => {
    const wrapper = mountWrapper(
      getStore({
        apps: { selected: { ...installedPackage } },
        config: {
          customAppViews: [
            {
              name: "demo-chart",
              plugin: PluginNames.PACKAGES_HELM,
              repository: "demo-repo",
            },
          ],
        },
      }),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppView />
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(CustomAppView)).not.toExist();
  });

  it("renders a RollBack button if the installedPackage is from PACKAGES_HELM", () => {
    const wrapper = mountWrapper(
      getStore({
        apps: {
          selected: {
            ...installedPackage,
            installedPackageRef: {
              ...installedPackage.installedPackageRef,
              plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" } as Plugin,
            } as InstalledPackageReference,
          },
        },
        config: {},
      }),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppView />
        </Route>
      </MemoryRouter>,
    );

    expect(wrapper.find(UpgradeButton)).toExist();
    expect(wrapper.find(RollbackButton)).toExist();
    expect(wrapper.find(DeleteButton)).toExist();
  });

  it("does not render a RollBack button if the installedPackage is not from PACKAGES_HELM", () => {
    const wrapper = mountWrapper(
      getStore({
        apps: { selected: { ...installedPackage } },
        config: {},
      }),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppView />
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(UpgradeButton)).toExist();
    expect(wrapper.find(RollbackButton)).not.toExist();
    expect(wrapper.find(DeleteButton)).toExist();
  });

  describe("State initialization", () => {
    /*
      The imported resource refs contain one deployment, one service, one config map and some bogus manifests.
    */
    it("sets ResourceRefs for its deployments, services, ingresses and secrets", () => {
      const apiResourceRefs = [
        resourceRefs.deployment,
        resourceRefs.service,
        resourceRefs.configMap,
        resourceRefs.ingress,
        resourceRefs.secret,
      ] as ResourceRef[];

      const wrapper = mountWrapper(
        getStore({ apps: { selected: { ...installedPackage, apiResourceRefs } } }),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Route path={routePath}>
            <AppView />
          </Route>
        </MemoryRouter>,
      );

      const tabs = wrapper.find(ResourceTabs);
      expect(tabs.prop("deployments")).toEqual([resourceRefs.deployment]);
      expect(tabs.prop("services")).toEqual([resourceRefs.service]);
      expect(tabs.prop("secrets")).toEqual([resourceRefs.secret]);
    });

    it("stores other k8s resources", () => {
      const apiResourceRefs = [
        resourceRefs.deployment,
        resourceRefs.service,
        resourceRefs.configMap,
        resourceRefs.secret,
      ] as ResourceRef[];

      const wrapper = mountWrapper(
        getStore({ apps: { selected: { ...installedPackage, apiResourceRefs } } }),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Route path={routePath}>
            <AppView />
          </Route>
        </MemoryRouter>,
      );

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
    it("renders all the elements of an application", () => {
      const wrapper = mountWrapper(getStore(validState), <AppView />);
      expect(wrapper.find(PackageInfo)).toExist();
      expect(wrapper.find(ApplicationStatusContainer)).toExist();
      expect(wrapper.find(".control-buttons")).toExist();
      expect(wrapper.find(AppNotes)).toExist();
      expect(wrapper.find(ResourceTabs)).toExist();
      expect(wrapper.find(AccessURLTable)).toExist();
    });

    it("renders an error if error prop is set", () => {
      const wrapper = mountWrapper(
        getStore({ ...validState, apps: { ...validState.apps, error: new Error("Boom!") } }),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Route path={routePath}>
            <AppView />
          </Route>
        </MemoryRouter>,
      );
      const err = wrapper.find(Alert);
      expect(err).toExist();
      expect(err.html()).toContain("Boom!");
    });

    it("renders a delete-error", () => {
      const wrapper = mountWrapper(
        getStore({ ...validState, apps: { ...validState.apps, error: new DeleteError("Boom!") } }),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Route path={routePath}>
            <AppView />
          </Route>
        </MemoryRouter>,
      );
      const err = wrapper.find(Alert);
      expect(err).toExist();
      expect(err.html()).toContain("Unable to delete the application. Received: Boom!");
    });
  });

  it("forwards statefulsets and daemonsets to the application status", () => {
    const apiResourceRefs = [resourceRefs.statefulset, resourceRefs.daemonset] as ResourceRef[];
    const wrapper = mountWrapper(
      getStore({ apps: { selected: { ...installedPackage, apiResourceRefs } } }),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppView />
        </Route>
      </MemoryRouter>,
    );

    const applicationStatus = wrapper.find(ApplicationStatusContainer);
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
    const store = getStore({ apps: { selected: { ...installedPackage, apiResourceRefs } } });

    mountWrapper(
      store,
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppView />
        </Route>
      </MemoryRouter>,
    );

    expect(store.getActions()).toEqual([
      {
        type: getType(actions.apps.requestApps),
      },
      {
        type: getType(actions.kube.requestResources),
        payload: {
          pkg: installedPackage.installedPackageRef,
          refs: [resourceRefs.secret],
          watch: false,
          handler: expect.any(Function),
          onError: expect.any(Function),
          onComplete: expect.any(Function),
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
          onComplete: expect.any(Function),
        },
      },
    ]);
  });
  it("closes the watches when unmounted", async () => {
    const apiResourceRefs = [resourceRefs.deployment, resourceRefs.service] as ResourceRef[];

    const store = getStore({ apps: { selected: { ...installedPackage, apiResourceRefs } } });
    const wrapper = mountWrapper(
      store,
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppView />
        </Route>
      </MemoryRouter>,
    );
    wrapper.unmount();

    const watch = true;
    expect(store.getActions()).toEqual([
      {
        type: getType(actions.apps.requestApps),
      },
      {
        type: getType(actions.kube.requestResources),
        payload: {
          pkg: installedPackage.installedPackageRef,
          refs: apiResourceRefs,
          watch,
          handler: expect.any(Function),
          onError: expect.any(Function),
          onComplete: expect.any(Function),
        },
      },
      {
        type: getType(actions.kube.closeRequestResources),
        payload: installedPackage.installedPackageRef,
      },
    ]);
  });
});
