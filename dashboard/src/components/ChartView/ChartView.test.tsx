import * as ReactRedux from "react-redux";

import actions from "actions";
import Alert from "components/js/Alert";
import { set } from "lodash";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { IChartState, IChartVersion } from "../../shared/types";
import ChartMaintainers from "./ChartMaintainers";
import ChartView, { IChartViewProps } from "./ChartView";

const props: IChartViewProps = {
  chartID: "testrepo/test",
  chartNamespace: "kubeapps-namespace",
  isFetching: false,
  namespace: "test",
  cluster: "default",
  selected: { versions: [] } as IChartState["selected"],
  version: undefined,
  kubeappsNamespace: "kubeapps",
};

const testChart: IChartVersion["relationships"]["chart"] = {
  data: {
    repo: {
      name: "testrepo",
    },
  },
} as IChartVersion["relationships"]["chart"];

const testVersion: IChartVersion = {
  attributes: {
    version: "1.2.3",
    app_version: "4.5.6",
    created: "",
  },
  id: "1",
  relationships: { chart: testChart },
};

const defaultSelected: IChartState["selected"] = {
  versions: [testVersion],
};

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.charts = {
    ...actions.charts,
    fetchChartVersionsAndSelectVersion: jest.fn(),
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

it("triggers the fetchChartVersionsAndSelectVersion when mounting", () => {
  const spy = jest.fn();
  actions.charts.fetchChartVersionsAndSelectVersion = spy;
  mountWrapper(defaultStore, <ChartView {...props} />);
  expect(spy).toHaveBeenCalledWith(props.cluster, props.chartNamespace, "testrepo/test", undefined);
});

describe("when receiving new props", () => {
  it("finds and selects the chart version when version changes", () => {
    const versions = [{ attributes: { version: "1.2.3" } }] as IChartVersion[];
    const spy = jest.fn();
    actions.charts = {
      ...actions.charts,
      selectChartVersion: spy,
    };
    mountWrapper(defaultStore, <ChartView {...props} selected={{ versions }} version={"1.2.3"} />);
    expect(spy).toHaveBeenCalledWith(versions[0]);
  });
});

it("triggers resetChartVersion when unmounting", () => {
  const spy = jest.fn();
  actions.charts = {
    ...actions.charts,
    resetChartVersion: spy,
  };
  const wrapper = mountWrapper(defaultStore, <ChartView {...props} />);
  wrapper.unmount();
  expect(spy).toHaveBeenCalled();
});

it("behaves as a loading component when fetching is false but no chart is available", () => {
  const wrapper = mountWrapper(defaultStore, <ChartView {...props} isFetching={false} />);
  expect(wrapper.find("LoadingWrapper")).toExist();
});

it("behaves as a loading component when fetching is true and chart is available", () => {
  const versions = [{ attributes: { version: "1.2.3" } }] as IChartVersion[];

  const wrapper = mountWrapper(
    defaultStore,
    <ChartView {...props} isFetching={true} selected={{ versions, version: versions[0] }} />,
  );
  expect(wrapper.find("LoadingWrapper")).toExist();
});

it("does not render the app version, home and sources sections if not set", () => {
  const version = { ...testVersion, attributes: { ...testVersion.attributes } };
  set(version, "attributes.app_version", undefined);
  const wrapper = mountWrapper(
    defaultStore,
    <ChartView {...props} selected={{ versions: [], version }} />,
  );
  expect(wrapper.contains("App Version")).toBe(false);
  expect(wrapper.contains("Home")).toBe(false);
  expect(wrapper.contains("Related")).toBe(false);
  expect(wrapper.contains("Maintainers")).toBe(false);
});

it("renders the app version when set", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <ChartView {...props} selected={{ ...defaultSelected, version: testVersion }} />,
  );
  expect(wrapper.contains("App Version")).toBe(true);
  expect(wrapper.contains(<div>{testVersion.attributes.app_version}</div>)).toBe(true);
});

it("renders the home link when set", () => {
  const v = testVersion as IChartVersion;
  v.relationships.chart.data.home = "https://example.com";
  const wrapper = mountWrapper(
    defaultStore,
    <ChartView {...props} selected={{ ...defaultSelected, version: v }} />,
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

describe("ChartMaintainers githubIDAsNames prop value", () => {
  const v = testVersion as IChartVersion;
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
      v.relationships.chart.data.maintainers = [{ name: "John Smith" }];
      v.relationships.chart.data.repo.url = t.repoURL;
      const wrapper = mountWrapper(
        defaultStore,
        <ChartView {...props} selected={{ ...defaultSelected, version: v }} />,
      );
      const chartMaintainers = wrapper.find(ChartMaintainers);
      expect(chartMaintainers.props().githubIDAsNames).toBe(t.expected);
    });
  }
});

it("renders the sources links when set", () => {
  const v = testVersion as IChartVersion;
  v.relationships.chart.data.sources = ["https://example.com", "https://example2.com"];
  const wrapper = mountWrapper(
    defaultStore,
    <ChartView {...props} selected={{ ...defaultSelected, version: v }} />,
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
      defaultStore,
      <ChartView {...props} selected={{ ...defaultSelected, error: new Error("Boom!") }} />,
    );
    expect(wrapper.find(Alert)).toExist();
    expect(wrapper.find(Alert).text()).toContain("Unable to fetch chart: Boom!");
  });
});
