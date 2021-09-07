import actions from "actions";
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
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import * as yaml from "js-yaml";
import * as ReactRedux from "react-redux";
import { MemoryRouter, Route } from "react-router";
import ResourceRef from "shared/ResourceRef";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { DeleteError, FetchError, IResource } from "shared/types";
import AccessURLTable from "./AccessURLTable/AccessURLTable";
import AppNotes from "./AppNotes/AppNotes";
import AppView from "./AppView";
import ChartInfo from "./ChartInfo/ChartInfo";
import ResourceTabs from "./ResourceTabs";

const routeParams = {
  cluster: "cluster-1",
  namespace: "default",
  releaseName: "mr-sunshine",
};
const routePathParam = `/foo/${routeParams.cluster}/${routeParams.namespace}/${routeParams.releaseName}`;
const routePath = "/foo/:cluster/:namespace/:releaseName";
let spyOnUseDispatch: jest.SpyInstance;
const appActions = { ...actions.apps };
const kubeaActions = { ...actions.kube };

beforeEach(() => {
  actions.apps = {
    ...actions.apps,
    getApp: jest.fn(),
  };
  actions.kube = {
    ...actions.kube,
    getAndWatchResource: jest.fn(),
    closeWatchResource: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.apps = { ...appActions };
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

describe("AppView", () => {
  // Generates a Yaml file separated by --- containing every object passed.
  const generateYamlManifest = (items: any[]): string => {
    let yamlManifest = "";
    items.forEach(i => {
      yamlManifest += "---\n" + yaml.dump(i);
    });
    return yamlManifest;
  };

  const installedPackage = {
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
  } as InstalledPackageDetail;

  const validState = { apps: { selected: installedPackage } };

  const resources = {
    configMap: { apiVersion: "v1", kind: "ConfigMap", metadata: { name: "cm-one" } },
    deployment: {
      apiVersion: "apps/v1beta1",
      kind: "Deployment",
      metadata: { name: "deployment-one" },
    } as IResource,
    service: { apiVersion: "v1", kind: "Service", metadata: { name: "svc-one" } } as IResource,
    ingress: {
      apiVersion: "extensions/v1beta1",
      kind: "Ingress",
      metadata: { name: "ingress-one" },
    } as IResource,
    secret: {
      apiVersion: "v1",
      kind: "Secret",
      metadata: { name: "secret-one" },
      type: "Opaque",
    } as IResource,
    daemonset: {
      apiVersion: "apps/v1beta1",
      kind: "DaemonSet",
      metadata: { name: "daemonset-one" },
    } as IResource,
    statefulset: {
      apiVersion: "apps/v1beta1",
      kind: "StatefulSet",
      metadata: { name: "statefulset-one" },
    } as IResource,
  };

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

  describe("State initialization", () => {
    /*
      The imported manifest contains one deployment, one service, one config map and some bogus manifests.
    */
    it("sets ResourceRefs for its deployments, services, ingresses and secrets", () => {
      const manifest = generateYamlManifest([
        resources.deployment,
        resources.service,
        resources.configMap,
        resources.ingress,
        resources.secret,
      ]);

      const wrapper = mountWrapper(
        getStore({ apps: { selected: { ...installedPackage, manifest } } }),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Route path={routePath}>
            <AppView />
          </Route>
        </MemoryRouter>,
      );

      const tabs = wrapper.find(ResourceTabs);
      expect(tabs.prop("deployments")).toEqual([
        new ResourceRef(
          resources.deployment,
          routeParams.cluster,
          "deployments",
          true,
          installedPackage.installedPackageRef?.context?.namespace,
        ),
      ]);
      expect(tabs.prop("services")).toEqual([
        new ResourceRef(
          resources.service,
          routeParams.cluster,
          "services",
          true,
          installedPackage.installedPackageRef?.context?.namespace,
        ),
      ]);
      expect(tabs.prop("secrets")).toEqual([
        new ResourceRef(
          resources.secret,
          routeParams.cluster,
          "secrets",
          true,
          installedPackage.installedPackageRef?.context?.namespace,
        ),
      ]);
    });

    it("watches the given resources and close watchers", async () => {
      const watchResource = jest.fn();
      const closeWatch = jest.fn();
      actions.kube = {
        ...actions.kube,
        getAndWatchResource: watchResource,
        closeWatchResource: closeWatch,
      };
      const manifest = generateYamlManifest([resources.deployment, resources.service]);
      const depResource = {
        cluster: routeParams.cluster,
        apiVersion: resources.deployment.apiVersion,
        kind: resources.deployment.kind,
        name: resources.deployment.metadata.name,
        namespace: installedPackage.installedPackageRef?.context?.namespace,
        namespaced: true,
        plural: "deployments",
      };
      const svcResource = {
        cluster: routeParams.cluster,
        apiVersion: resources.service.apiVersion,
        kind: resources.service.kind,
        name: resources.service.metadata.name,
        namespace: installedPackage.installedPackageRef?.context?.namespace,
        namespaced: true,
        plural: "services",
      };

      const wrapper = mountWrapper(
        getStore({ apps: { selected: { ...installedPackage, manifest } } }),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Route path={routePath}>
            <AppView />
          </Route>
        </MemoryRouter>,
      );
      expect(watchResource).toHaveBeenCalledWith(depResource);
      expect(watchResource).toHaveBeenCalledWith(svcResource);
      wrapper.unmount();
      expect(closeWatch).toHaveBeenCalledWith(depResource);
      expect(closeWatch).toHaveBeenCalledWith(svcResource);
    });

    it("stores other k8s resources", () => {
      const manifest = generateYamlManifest([
        resources.deployment,
        resources.service,
        resources.configMap,
        resources.secret,
      ]);

      const wrapper = mountWrapper(
        getStore({ apps: { selected: { ...installedPackage, manifest } } }),
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

    it("does not store empty resources, bogus or without kind attribute", () => {
      const manifest = generateYamlManifest([
        { apiVersion: "v1", metadata: { name: "cm-one" } },
        {},
        "# This is a comment",
        " ",
      ]);

      const wrapper = mountWrapper(
        getStore({ apps: { selected: { ...installedPackage, manifest } } }),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Route path={routePath}>
            <AppView />
          </Route>
        </MemoryRouter>,
      );

      const tabs = wrapper.find(ResourceTabs);
      expect(tabs.prop("otherResources")).toEqual([]);
      expect(tabs.prop("deployments")).toEqual([]);
      expect(tabs.prop("services")).toEqual([]);
      expect(tabs.prop("secrets")).toEqual([]);
    });

    // See https://github.com/kubeapps/kubeapps/issues/632
    it("supports manifests with duplicated keys", () => {
      const manifest = `
      apiVersion: v1
      metadata:
        name: cm-one
        labels:
          chart: cm-1.2.3
          chart: cm-1.2.3
`;

      expect(() => {
        mountWrapper(
          getStore({ apps: { selected: { ...installedPackage, manifest } } }),
          <MemoryRouter initialEntries={[routePathParam]}>
            <Route path={routePath}>
              <AppView />
            </Route>
          </MemoryRouter>,
        );
      }).not.toThrow();
    });

    it("supports manifests with YAML type casting", () => {
      const manifest = `
      apiVersion: v1
      kind: Deployment
      metadata:
        name: !!string foo
`;

      expect(() => {
        const wrapper = mountWrapper(
          getStore({ apps: { selected: { ...installedPackage, manifest } } }),
          <MemoryRouter initialEntries={[routePathParam]}>
            <Route path={routePath}>
              <AppView />
            </Route>
          </MemoryRouter>,
        );
        const tabs = wrapper.find(ResourceTabs);
        expect(tabs.prop("deployments")[0].name).toEqual("foo");
      }).not.toThrow();
    });
  });

  describe("renderization", () => {
    it("renders all the elements of an application", () => {
      const wrapper = mountWrapper(getStore(validState), <AppView />);
      expect(wrapper.find(ChartInfo)).toExist();
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

  it("renders a list of resources", () => {
    const obj = { kind: "ClusterRole", metadata: { name: "foo" } } as IResource;
    const list = {
      kind: "List",
      items: [obj, resources.deployment],
    };
    const manifest = generateYamlManifest([resources.service, list]);
    const wrapper = mountWrapper(
      getStore({ apps: { selected: { ...installedPackage, manifest } } }),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppView />
        </Route>
      </MemoryRouter>,
    );

    const tabs = wrapper.find(ResourceTabs);
    expect(tabs.props()).toMatchObject({
      deployments: [
        new ResourceRef(
          resources.deployment,
          routeParams.cluster,
          "deployments",
          true,
          installedPackage.installedPackageRef?.context?.namespace,
        ),
      ],
      services: [
        new ResourceRef(
          resources.service,
          routeParams.cluster,
          "services",
          true,
          installedPackage.installedPackageRef?.context?.namespace,
        ),
      ],
      otherResources: [
        new ResourceRef(
          obj,
          routeParams.cluster,
          "clusterroles",
          false,
          installedPackage.installedPackageRef?.context?.namespace,
        ),
      ],
    });
  });

  it("renders a list of roles", () => {
    const obj = { kind: "ClusterRole", metadata: { name: "foo" } } as IResource;
    const list = {
      kind: "RoleList",
      items: [obj, resources.deployment],
    };
    const manifest = generateYamlManifest([resources.service, list]);

    const wrapper = mountWrapper(
      getStore({ apps: { selected: { ...installedPackage, manifest } } }),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppView />
        </Route>
      </MemoryRouter>,
    );

    const tabs = wrapper.find(ResourceTabs);
    expect(tabs.props()).toMatchObject({
      deployments: [
        new ResourceRef(
          resources.deployment,
          routeParams.cluster,
          "deployments",
          true,
          installedPackage.installedPackageRef?.context?.namespace,
        ),
      ],
      services: [
        new ResourceRef(
          resources.service,
          routeParams.cluster,
          "services",
          true,
          installedPackage.installedPackageRef?.context?.namespace,
        ),
      ],
      otherResources: [
        new ResourceRef(
          obj,
          routeParams.cluster,
          "clusterroles",
          false,
          installedPackage.installedPackageRef?.context?.namespace,
        ),
      ],
    });
  });

  it("forwards statefulsets and daemonsets to the application status", () => {
    const r = [resources.statefulset, resources.daemonset];
    const manifest = generateYamlManifest(r);
    const wrapper = mountWrapper(
      getStore({ apps: { selected: { ...installedPackage, manifest } } }),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppView />
        </Route>
      </MemoryRouter>,
    );

    const applicationStatus = wrapper.find(ApplicationStatusContainer);
    expect(applicationStatus).toExist();

    expect(applicationStatus.prop("statefulsetRefs")).toEqual([
      new ResourceRef(
        resources.statefulset,
        routeParams.cluster,
        "statefulsets",
        true,
        installedPackage.installedPackageRef?.context?.namespace,
      ),
    ]);
    expect(applicationStatus.prop("daemonsetRefs")).toEqual([
      new ResourceRef(
        resources.daemonset,
        routeParams.cluster,
        "daemonsets",
        true,
        installedPackage.installedPackageRef?.context?.namespace,
      ),
    ]);
  });
});
