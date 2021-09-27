import actions from "actions";
import Alert from "components/js/Alert";
import {
  AvailablePackageDetail,
  AvailablePackageReference,
  Context,
  Maintainer,
  PackageAppVersion,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { createMemoryHistory } from "history";
import * as ReactRedux from "react-redux";
import { Route, Router } from "react-router";
import { IConfigState } from "reducers/config";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IChartState } from "../../shared/types";
import AvailablePackageMaintainers from "./AvailablePackageMaintainers";
import ChartReadme from "./ChartReadme";
import ChartView from "./ChartView";

const defaultProps = {
  chartID: "testrepo/test",
  chartNamespace: "kubeapps-namespace",
  isFetching: false,
  namespace: "test",
  cluster: "default",
  selected: { versions: [] } as IChartState["selected"],
  version: undefined,
  kubeappsNamespace: "kubeapps",
  repo: "testrepo",
  id: "test",
  plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
};

const testVersion: PackageAppVersion = {
  pkgVersion: "1.2.3",
  appVersion: "4.5.6",
};

const defaultAvailablePkgDetail: AvailablePackageDetail = {
  name: "foo",
  categories: [""],
  displayName: "foo",
  iconUrl: "https://icon.com",
  repoUrl: "https://repo.com",
  homeUrl: "https://example.com",
  sourceUrls: ["test"],
  shortDescription: "test",
  longDescription: "test",
  availablePackageRef: {
    identifier: "foo/foo",
    context: { cluster: "", namespace: "chart-namespace" } as Context,
    plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
  },
  valuesSchema: "test",
  defaultValues: "test",
  maintainers: [{ name: "test", email: "test" }] as Maintainer[],
  readme: "test",
  version: {
    appVersion: testVersion.appVersion,
    pkgVersion: testVersion.pkgVersion,
  } as PackageAppVersion,
};

const emptyAvailablePkg: AvailablePackageDetail = {
  name: "foo",
  categories: [""],
  displayName: "foo",
  iconUrl: "",
  repoUrl: "",
  homeUrl: "",
  sourceUrls: [],
  shortDescription: "",
  longDescription: "",
  availablePackageRef: {
    identifier: "foo/foo",
    context: { cluster: "", namespace: "chart-namespace" } as Context,
    plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
  },
  valuesSchema: "",
  defaultValues: "",
  maintainers: [],
  readme: "",
  version: {
    appVersion: "",
    pkgVersion: testVersion.pkgVersion,
  },
};

const defaultChartsState = {
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
  } as IChartState["selected"],
  deployed: {} as IChartState["deployed"],
  items: [],
  categories: [],
  size: 20,
} as IChartState;

const defaultState = {
  charts: defaultChartsState,
  config: { kubeappsCluster: "default", kubeappsNamespace: "kubeapps" } as IConfigState,
};

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.charts = {
    ...actions.charts,
    fetchChartVersions: jest.fn(),
    resetChartVersion: jest.fn(),
    selectChartVersion: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

const routePathParam = `/c/${defaultProps.cluster}/ns/${defaultProps.chartNamespace}/charts/${defaultProps.repo}/${defaultProps.plugin.name}/${defaultProps.plugin.version}/${defaultProps.id}`;
const routePath = "/c/:cluster/ns/:namespace/charts/:repo/:pluginName/:pluginVersion/:id";
const history = createMemoryHistory({ initialEntries: [routePathParam] });

it("triggers the fetchChartVersions when mounting", () => {
  const spy = jest.fn();
  actions.charts.fetchChartVersions = spy;
  mountWrapper(
    getStore(defaultState),
    <Router history={history}>
      <Route path={routePath}>
        <ChartView />
      </Route>
    </Router>,
  );
  expect(spy).toHaveBeenCalledWith({
    context: { cluster: defaultProps.cluster, namespace: defaultProps.chartNamespace },
    identifier: `${defaultProps.repo}/${defaultProps.id}`,
    plugin: defaultProps.plugin,
  } as AvailablePackageReference);
});

describe("when receiving new props", () => {
  it("finds and selects the chart version when version changes", () => {
    const spy = jest.fn();
    actions.charts.fetchChartVersion = spy;
    mountWrapper(
      getStore(defaultState),
      <Router history={history}>
        <Route path={routePath}>
          <ChartView />
        </Route>
      </Router>,
    );
    expect(spy).toHaveBeenCalledWith(
      {
        context: { cluster: defaultProps.cluster, namespace: defaultProps.chartNamespace },
        identifier: "testrepo/test",
        plugin: defaultProps.plugin,
      } as AvailablePackageReference,
      undefined,
    );
  });
});

it("behaves as a loading component when fetching is false but no chart is available", () => {
  const wrapper = mountWrapper(
    getStore({
      ...defaultState,
      charts: { ...defaultChartsState, selected: {}, isFetching: false },
    }),
    <Router history={history}>
      <Route path={routePath}>
        <ChartView />
      </Route>
    </Router>,
  );
  expect(wrapper.find("LoadingWrapper")).toExist();
});

it("behaves as a loading component when fetching is true and chart is available", () => {
  const wrapper = mountWrapper(
    getStore({ ...defaultState, charts: { ...defaultChartsState, isFetching: false } }),
    <Router history={history}>
      <Route path={routePath}>
        <ChartView />
      </Route>
    </Router>,
  );
  expect(wrapper.find("LoadingWrapper")).toExist();
});

it("does not render the app version, home and sources sections if not set", () => {
  const wrapper = mountWrapper(
    getStore({
      ...defaultChartsState,
      charts: { selected: { availablePackageDetail: emptyAvailablePkg } },
    }),
    <Router history={history}>
      <Route path={routePath}>
        <ChartView />
      </Route>
    </Router>,
  );
  expect(wrapper.contains("App Version")).toBe(false);
  expect(wrapper.contains("Home")).toBe(false);
  expect(wrapper.contains("Related")).toBe(false);
  expect(wrapper.contains("Maintainers")).toBe(false);
});

it("renders the app version when set", () => {
  const wrapper = mountWrapper(
    getStore(defaultState),
    <Router history={history}>
      <Route path={routePath}>
        <ChartView />
      </Route>
    </Router>,
  );
  expect(wrapper.contains("App Version")).toBe(true);
  expect(wrapper.contains(<div>{testVersion.appVersion}</div>)).toBe(true);
});

it("renders the home link when set", () => {
  const wrapper = mountWrapper(
    getStore(defaultState),
    <Router history={history}>
      <Route path={routePath}>
        <ChartView />
      </Route>
    </Router>,
  );
  expect(wrapper.contains("Home")).toBe(true);
  expect(
    wrapper.contains(
      <a href="https://example.com" target="_blank" rel="noopener noreferrer">
        {"https://example.com"}
      </a>,
    ),
  ).toBe(true);
});

describe("when package details are not available", () => {
  it("redirects when skipAvailablePackageDetails is set to true", () => {
    const wrapper = mountWrapper(
      getStore({
        ...defaultState,
        charts: { selected: { readme: "" } },
        config: { skipAvailablePackageDetails: true },
      }),
      <Router history={history}>
        <Route path={routePath}>
          <ChartView />
        </Route>
      </Router>,
    );
    expect(wrapper.text()).not.toContain("Fetching application README...");
  });

  it("does not redirect when skipAvailablePackageDetails is set to false", () => {
    const wrapper = mountWrapper(
      getStore({
        ...defaultState,
        config: { skipAvailablePackageDetails: false },
      }),
      <Router history={history}>
        <Route path={routePath}>
          <ChartView />
        </Route>
      </Router>,
    );
    expect(wrapper.containsMatchingElement(<ChartReadme />)).toBe(true);
  });
});

describe("ChartMaintainers githubIDAsNames prop value", () => {
  const tests: Array<{
    expected: boolean;
    name: string;
    repoURL: string;
    maintainers: Array<{ name: string; email?: string }>;
  }> = [
    {
      expected: true,
      name: "the stable Helm repo uses github IDs",
      maintainers: [{ name: "Bitnami" }],
      repoURL: "https://kubernetes-charts.storage.googleapis.com",
    },
    {
      expected: true,
      name: "the incubator Helm repo uses github IDs",
      maintainers: [{ name: "Bitnami", email: "email: containers@bitnami.com" }],
      repoURL: "https://kubernetes-charts-incubator.storage.googleapis.com",
    },
    {
      expected: false,
      name: "a random Helm repo does not use github IDs as names",
      maintainers: [{ name: "Bitnami" }],
      repoURL: "https://examplerepo.com",
    },
  ];

  for (const t of tests) {
    it(`for ${t.name}`, () => {
      const myAvailablePkgDetail = defaultAvailablePkgDetail;
      myAvailablePkgDetail.maintainers = [{ name: "John Smith", email: "john@example.com" }];
      myAvailablePkgDetail.repoUrl = t.repoURL;

      const wrapper = mountWrapper(
        getStore({
          ...defaultState,
          charts: { selected: { availablePackageDetail: myAvailablePkgDetail } },
        }),
        <Router history={history}>
          <Route path={routePath}>
            <ChartView />
          </Route>
        </Router>,
      );

      const chartMaintainers = wrapper.find(AvailablePackageMaintainers);
      expect(chartMaintainers.props().githubIDAsNames).toBe(t.expected);
    });
  }
});

it("renders the sources links when set", () => {
  const myAvailablePkgDetail = defaultAvailablePkgDetail;
  myAvailablePkgDetail.sourceUrls = ["https://example.com", "https://example2.com"];
  const wrapper = mountWrapper(
    getStore({
      ...defaultState,
      charts: { selected: { availablePackageDetail: myAvailablePkgDetail } },
    }),
    <Router history={history}>
      <Route path={routePath}>
        <ChartView />
      </Route>
    </Router>,
  );
  expect(wrapper.contains("Related")).toBe(true);
  expect(
    wrapper.contains(
      <a href="https://example.com" target="_blank" rel="noopener noreferrer">
        {"https://example.com"}
      </a>,
    ),
  ).toBe(true);
  expect(
    wrapper.contains(
      <a href="https://example2.com" target="_blank" rel="noopener noreferrer">
        {"https://example2.com"}
      </a>,
    ),
  ).toBe(true);
});

describe("renders errors", () => {
  it("renders a not found error if it exists", () => {
    const wrapper = mountWrapper(
      getStore({
        ...defaultState,
        charts: { ...defaultChartsState, selected: { error: new Error("Boom!") } },
      }),
      <Router history={history}>
        <Route path={routePath}>
          <ChartView />
        </Route>
      </Router>,
    );
    expect(wrapper.find(Alert)).toExist();
    expect(wrapper.find(Alert).text()).toContain("Unable to fetch package: Boom!");
  });
});
