import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import itBehavesLike from "../../shared/specs";
import { IChartState, IChartVersion, NotFoundError } from "../../shared/types";
import { ErrorSelector } from "../ErrorAlert";
import ChartHeader from "./ChartHeader";
import ChartMaintainers from "./ChartMaintainers";
import ChartReadme from "./ChartReadme";
import ChartVersionsList from "./ChartVersionsList";
import ChartView, { IChartViewProps } from "./ChartView";

const props: IChartViewProps = {
  chartID: "testrepo/test",
  chartNamespace: "kubeapps-namespace",
  fetchChartVersionsAndSelectVersion: jest.fn(),
  getChartReadme: jest.fn(),
  isFetching: false,
  namespace: "test",
  cluster: "default",
  resetChartVersion: jest.fn(),
  selectChartVersion: jest.fn(),
  selected: {} as IChartState["selected"],
  version: undefined,
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

it("triggers the fetchChartVersionsAndSelectVersion when mounting", () => {
  const spy = jest.fn();
  shallow(<ChartView {...props} fetchChartVersionsAndSelectVersion={spy} />);
  expect(spy).toHaveBeenCalledWith("kubeapps-namespace", "testrepo/test", undefined);
});

describe("when receiving new props", () => {
  it("finds and selects the chart version when version changes", () => {
    const versions = [{ attributes: { version: "1.2.3" } }] as IChartVersion[];
    const spy = jest.fn();
    const wrapper = shallow(
      <ChartView {...props} selectChartVersion={spy} selected={{ versions }} />,
    );
    wrapper.setProps({ version: "1.2.3" });
    expect(spy).toHaveBeenCalledWith(versions[0]);
  });

  it("does not trigger selectChartVersion if version does not change", () => {
    const spy = jest.fn();
    const wrapper = shallow(<ChartView {...props} selectChartVersion={spy} version="1.2.3" />);
    wrapper.setProps({ isFetching: true });
    expect(spy).toHaveBeenCalledTimes(0);
  });

  it("throws an error if the chart version doesn't exist", () => {
    const versions = [{ attributes: { version: "1.2.3" } }] as IChartVersion[];
    const spy = jest.fn();
    const wrapper = shallow(
      <ChartView {...props} selectChartVersion={spy} selected={{ versions }} />,
    );
    expect(() => {
      wrapper.setProps({ version: "1.0.0" });
    }).toThrow("could not find chart");
  });
});

it("triggers resetChartVersion when unmounting", () => {
  const spy = jest.fn();
  const wrapper = shallow(<ChartView {...props} resetChartVersion={spy} />);
  wrapper.unmount();
  expect(spy).toHaveBeenCalled();
});

context("when fetching is false but no chart is available", () => {
  itBehavesLike("aLoadingComponent", {
    component: ChartView,
    props: {
      ...props,
      isFetching: false,
    },
  });
});

context("when fetching is true and chart is available", () => {
  itBehavesLike("aLoadingComponent", {
    component: ChartView,
    props: {
      ...props,
      isFetching: true,
      selected: { version: {} as IChartVersion },
    },
  });
});

describe("subcomponents", () => {
  const wrapper = shallow(
    <ChartView {...props} selected={{ ...defaultSelected, version: testVersion }} />,
  );

  for (const component of [ChartHeader, ChartReadme, ChartVersionsList]) {
    it(`renders ${component.name}`, () => {
      expect(wrapper.find(component).exists()).toBe(true);
    });
  }
});

it("does not render the app version, home and sources sections if not set", () => {
  const version = { ...testVersion, attributes: { ...testVersion.attributes } };
  delete version.attributes.app_version;
  const wrapper = shallow(<ChartView {...props} selected={{ versions: [], version }} />);
  expect(wrapper.contains(<h2>App Version</h2>)).toBe(false);
  expect(wrapper.contains(<h2>Home</h2>)).toBe(false);
  expect(wrapper.contains(<h2>Related</h2>)).toBe(false);
  expect(wrapper.contains(<h2>Maintainers</h2>)).toBe(false);
});

it("renders the app version when set", () => {
  const wrapper = shallow(
    <ChartView {...props} selected={{ ...defaultSelected, version: testVersion }} />,
  );
  expect(wrapper.contains(<h2>App Version</h2>)).toBe(true);
  expect(wrapper.contains(<div>{testVersion.attributes.app_version}</div>)).toBe(true);
});

it("renders the home link when set", () => {
  const v = testVersion as IChartVersion;
  v.relationships.chart.data.home = "https://example.com";
  const wrapper = shallow(<ChartView {...props} selected={{ ...defaultSelected, version: v }} />);
  expect(wrapper.contains(<h2>Home</h2>)).toBe(true);
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
      name: "stable Helm repo",
      maintainers: [{ name: "Bitnami" }],
      repoURL: "https://kubernetes-charts.storage.googleapis.com",
    },
    {
      expected: true,
      name: "incubator Helm repo",
      maintainers: [{ name: "Bitnami", email: "email: containers@bitnami.com" }],
      repoURL: "https://kubernetes-charts-incubator.storage.googleapis.com",
    },
    {
      expected: false,
      name: "random Helm repo",
      maintainers: [{ name: "Bitnami" }],
      repoURL: "https://examplerepo.com",
    },
  ];

  for (const t of tests) {
    it(`for ${t.name}`, () => {
      v.relationships.chart.data.maintainers = [{ name: "John Smith" }];
      v.relationships.chart.data.repo.url = t.repoURL;
      const wrapper = shallow(
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
  const wrapper = shallow(<ChartView {...props} selected={{ ...defaultSelected, version: v }} />);
  expect(wrapper.contains(<h2>Related</h2>)).toBe(true);
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
    const wrapper = shallow(
      <ChartView {...props} selected={{ ...defaultSelected, error: new NotFoundError() }} />,
    );
    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper.find(ErrorSelector).html()).toContain(`Chart ${props.chartID} not found`);
  });
  it("renders a generic error if it exists", () => {
    const wrapper = shallow(
      <ChartView {...props} selected={{ ...defaultSelected, error: new Error() }} />,
    );
    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper.find(ErrorSelector).html()).toContain("Sorry! Something went wrong");
  });
});
