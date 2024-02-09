// Copyright 2021-2024 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import "@testing-library/jest-dom";
import { screen } from "@testing-library/react";
import actions from "actions";
import {
  AvailablePackageDetail,
  AvailablePackageReference,
  Context,
  Maintainer,
  PackageAppVersion,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins_pb";
import * as ReactRedux from "react-redux";
import { Route, Routes } from "react-router-dom";
import { IConfigState } from "reducers/config";
import { getStore, renderWithProviders } from "shared/specs/mountWrapper";
import { IPackageState, IStoreState } from "shared/types";
import PackageView from "./PackageView";

const defaultProps = {
  packageID: "testrepo/test",
  packageNamespace: "kubeapps-namespace",
  isFetching: false,
  namespace: "test",
  cluster: "default",
  selected: { versions: [], metadatas: [] } as IPackageState["selected"],
  version: undefined,
  kubeappsNamespace: "kubeapps",
  id: "test",
  plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
};

const testVersion = new PackageAppVersion({
  pkgVersion: "1.2.3",
  appVersion: "4.5.6",
});

const defaultAvailablePkgDetail = new AvailablePackageDetail({
  name: "foo",
  categories: [""],
  displayName: "foo",
  iconUrl: "https://icon.com",
  repoUrl: "https://repo.com",
  homeUrl: "https://example.com",
  sourceUrls: ["test"],
  shortDescription: "test",
  longDescription: "test",
  availablePackageRef: new AvailablePackageReference({
    identifier: "foo/foo",
    context: { cluster: "", namespace: "package-namespace" } as Context,
    plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
  }),
  valuesSchema: "test",
  defaultValues: "test",
  additionalDefaultValues: {},
  maintainers: [{ name: "test", email: "test" }] as Maintainer[],
  readme: "test",
  version: {
    appVersion: testVersion.appVersion,
    pkgVersion: testVersion.pkgVersion,
  } as PackageAppVersion,
});

const defaultPackageState = {
  isFetching: false,
  hasFinishedFetching: false,
  selected: {
    error: undefined,
    availablePackageDetail: defaultAvailablePkgDetail,
    pkgVersion: testVersion.pkgVersion,
    appVersion: testVersion.appVersion,
    readme: "readme",
    readmeError: undefined,
    values: "values",
    versions: [testVersion],
  } as IPackageState["selected"],
  items: [],
  categories: [],
  nextPageToken: "",
  size: 20,
} as IPackageState;

const defaultState = {
  packages: defaultPackageState,
  config: {
    kubeappsCluster: "default",
    kubeappsNamespace: "kubeapps",
    skipAvailablePackageDetails: false,
  } as IConfigState,
} as IStoreState;

let spyOnUseDispatch: jest.SpyInstance;
const kubeActions = { ...actions.kube };

beforeEach(() => {
  actions.availablepackages = {
    ...actions.availablepackages,
    fetchAvailablePackageVersions: jest.fn(),
    resetSelectedAvailablePackageDetail: jest.fn(),
    receiveSelectedAvailablePackageDetail: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeActions };
  spyOnUseDispatch.mockRestore();
});

const routePathParam = `/c/${defaultProps.cluster}/ns/${defaultProps.namespace}/packages/${defaultProps.plugin.name}/${defaultProps.plugin.version}/${defaultProps.cluster}/${defaultProps.packageNamespace}/${defaultProps.id}`;
const routePath =
  "/c/:cluster/ns/:namespace/packages/:pluginName/:pluginVersion/:packageCluster/:packageNamespace/:packageId";

it("triggers the fetchAvailablePackageVersions when mounting", () => {
  const spy = jest.fn();
  actions.availablepackages.fetchAvailablePackageVersions = spy;
  renderWithProviders(
    <Routes>
      <Route path={routePath} element={<PackageView />} />
    </Routes>,
    {
      store: getStore(defaultState),
      initialEntries: [routePathParam],
    },
  );

  expect(spy).toHaveBeenCalledWith({
    context: { cluster: defaultProps.cluster, namespace: defaultProps.packageNamespace },
    identifier: defaultProps.id,
    plugin: defaultProps.plugin,
  } as AvailablePackageReference);
});

describe("when receiving new props", () => {
  it("finds and selects the package version when version changes", () => {
    const spy = jest.fn();
    actions.availablepackages.fetchAndSelectAvailablePackageDetail = spy;
    renderWithProviders(
      <Routes>
        <Route path={routePath} element={<PackageView />} />
      </Routes>,
      {
        store: getStore(defaultState),
        initialEntries: [routePathParam],
      },
    );

    expect(spy).toHaveBeenCalledWith(
      {
        context: { cluster: defaultProps.cluster, namespace: defaultProps.packageNamespace },
        identifier: defaultProps.id,
        plugin: defaultProps.plugin,
      } as AvailablePackageReference,
      undefined,
    );
  });
});

it("behaves as a loading component when fetching is false but no package is available", () => {
  const store = getStore({
    ...defaultState,
    packages: { ...defaultPackageState, selected: {}, isFetching: false },
  } as IStoreState);
  renderWithProviders(
    <Routes>
      <Route path={routePath} element={<PackageView />} />
    </Routes>,
    {
      store,
      initialEntries: [routePathParam],
    },
  );

  expect(screen.getByLabelText("Loading")).toBeInTheDocument();
});

it("behaves as a loading component when fetching is true and the package is available", () => {
  const store = getStore({
    ...defaultState,
    packages: { ...defaultPackageState, isFetching: true },
  } as IStoreState);
  renderWithProviders(
    <Routes>
      <Route path={routePath} element={<PackageView />} />
    </Routes>,
    {
      store,
      initialEntries: [routePathParam],
    },
  );

  expect(screen.getByLabelText("Loading")).toBeInTheDocument();
});

it("does not render the app version, home and sources sections if not set", () => {
  const store = getStore({
    ...defaultState,
    packages: {
      ...defaultPackageState,
      selected: { availablePackageDetail: undefined },
    },
  } as IStoreState);
  renderWithProviders(
    <Routes>
      <Route path={routePath} element={<PackageView />} />
    </Routes>,
    {
      store,
      initialEntries: [routePathParam],
    },
  );

  expect(screen.queryByText("App Version")).not.toBeInTheDocument();
  expect(screen.queryByText("Home")).not.toBeInTheDocument();
  expect(screen.queryByText("Related")).not.toBeInTheDocument();
  expect(screen.queryByText("Maintainers")).not.toBeInTheDocument();
});

it("renders the app version when set", () => {
  renderWithProviders(
    <Routes>
      <Route path={routePath} element={<PackageView />} />
    </Routes>,
    {
      store: getStore(defaultState),
      initialEntries: [routePathParam],
    },
  );

  expect(screen.getByText("App Version")).toBeInTheDocument();
  expect(screen.getByText(testVersion.appVersion)).toBeInTheDocument();
});

it("renders the home link when set", () => {
  const store = getStore(defaultState);
  renderWithProviders(
    <Routes>
      <Route path={routePath} element={<PackageView />} />
    </Routes>,
    {
      store,
      initialEntries: [routePathParam],
    },
  );

  expect(screen.getByText("Home")).toBeInTheDocument();
  expect(screen.getByRole("link", { name: "https://example.com" })).toHaveAttribute(
    "href",
    "https://example.com",
  );
});

describe("when setting the skipAvailablePackageDetails option", () => {
  it("does not redirect when skipAvailablePackageDetails is set to false", () => {
    const store = getStore({
      ...defaultState,
      config: { skipAvailablePackageDetails: false },
    } as IStoreState);

    renderWithProviders(
      <Routes>
        <Route path={routePath} element={<PackageView />} />
      </Routes>,
      {
        store,
        initialEntries: [routePathParam],
      },
    );

    expect(screen.getByText("readme")).toBeInTheDocument();
  });

  it("redirects when skipAvailablePackageDetails is set to true", () => {
    const store = getStore({
      ...defaultState,
      config: { skipAvailablePackageDetails: true },
    } as IStoreState);

    renderWithProviders(
      <Routes>
        <Route path={routePath} element={<PackageView />} />
        <Route
          path={
            "/c/:cluster/ns/:namespace/apps/new/:pluginName/:pluginVersion/:packageCluster/:packageNamespace/:something/versions/:versionId"
          }
          element={<h1>NewApp</h1>}
        />
      </Routes>,
      {
        store,
        initialEntries: [routePathParam],
      },
    );

    expect(screen.queryByText("readme")).not.toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "NewApp" })).toBeInTheDocument();
  });
});

describe("AvailablePackageMaintainers githubIDAsNames prop value", () => {
  const tests: Array<{
    expectGHLink: boolean;
    name: string;
    repoURL: string;
    maintainers: Array<{ name: string; email?: string }>;
  }> = [
    {
      expectGHLink: true,
      name: "the stable Helm repo uses github IDs",
      maintainers: [{ name: "Bitnami" }],
      repoURL: "https://kubernetes-charts.storage.googleapis.com",
    },
    {
      expectGHLink: true,
      name: "the incubator Helm repo uses github IDs",
      maintainers: [{ name: "Bitnami", email: "email: containers@bitnami.com" }],
      repoURL: "https://kubernetes-charts-incubator.storage.googleapis.com",
    },
    {
      expectGHLink: false,
      name: "a random Helm repo does not use github IDs as names",
      maintainers: [{ name: "Bitnami" }],
      repoURL: "https://examplerepo.com",
    },
  ];

  for (const t of tests) {
    it(`for ${t.name}`, () => {
      const myAvailablePkgDetail = defaultAvailablePkgDetail;
      myAvailablePkgDetail.maintainers = [
        new Maintainer({ name: "John Smith", email: "john@example.com" }),
      ];
      myAvailablePkgDetail.repoUrl = t.repoURL;

      const store = getStore({
        ...defaultState,
        packages: {
          selected: { availablePackageDetail: myAvailablePkgDetail, pkgVersion: "0.0.1" },
        },
      } as IStoreState);
      renderWithProviders(
        <Routes>
          <Route path={routePath} element={<PackageView />} />
        </Routes>,
        {
          store,
          initialEntries: [routePathParam],
        },
      );

      const maintainerLink = screen.getByRole("link", { name: "John Smith" });
      expect(maintainerLink).toBeInTheDocument();
      if (t.expectGHLink) {
        expect(maintainerLink).toHaveAttribute("href", "https://github.com/John Smith");
      } else {
        expect(maintainerLink).toHaveAttribute("href", "mailto:john@example.com");
      }
    });
  }
});

it("renders the sources links when set", () => {
  const myAvailablePkgDetail = defaultAvailablePkgDetail;
  myAvailablePkgDetail.sourceUrls = ["https://example.com", "https://example2.com"];
  const store = getStore({
    ...defaultState,
    packages: { selected: { availablePackageDetail: myAvailablePkgDetail, pkgVersion: "0.0.1" } },
  } as IStoreState);
  renderWithProviders(
    <Routes>
      <Route path={routePath} element={<PackageView />} />
    </Routes>,
    {
      store,
      initialEntries: [routePathParam],
    },
  );

  expect(screen.getByText("Related")).toBeInTheDocument();
  expect(screen.getAllByRole("link", { name: "https://example.com" })[0]).toBeInTheDocument();
  expect(screen.getByRole("link", { name: "https://example2.com" })).toBeInTheDocument();
});

describe("renders errors", () => {
  it("renders a not found error if it exists", () => {
    const store = getStore({
      ...defaultState,
      packages: { ...defaultPackageState, selected: { error: new Error("Boom!") } },
    } as IStoreState);
    renderWithProviders(
      <Routes>
        <Route path={routePath} element={<PackageView />} />
      </Routes>,
      {
        store,
        initialEntries: [routePathParam],
      },
    );

    expect(screen.getAllByRole("region")[1]).toBeInTheDocument();
    expect(screen.getAllByRole("region")[1]).toHaveTextContent(
      "Unable to fetch the package: Boom!.",
    );
  });
});
