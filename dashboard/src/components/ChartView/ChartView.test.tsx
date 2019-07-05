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
import ChartView from "./ChartView";

const props: any = {
  chartID: "testrepo/test",
  fetchChartVersionsAndSelectVersion: jest.fn(),
  getChartReadme: jest.fn(),
  isFetching: false,
  namespace: "test",
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
    sources: [] as string[],
  },
} as IChartVersion["relationships"]["chart"];

const testVersion = {
  attributes: {
    version: "1.2.3",
  },
  id: "1",
  relationships: { chart: testChart },
};

it("triggers the fetchChartVersionsAndSelectVersion when mounting", () => {
  const spy = jest.fn();
  shallow(<ChartView {...props} fetchChartVersionsAndSelectVersion={spy} />);
  expect(spy).toHaveBeenCalledWith("testrepo/test", undefined);
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
  const wrapper = shallow(<ChartView {...props} selected={{ version: testVersion }} />);

  for (const component of [ChartHeader, ChartReadme, ChartVersionsList, ChartMaintainers]) {
    it(`renders ${component.name}`, () => {
      expect(wrapper.find(component).exists()).toBe(true);
    });
  }
});

it("does not render the app version, home and sources sections if not set", () => {
  const wrapper = shallow(<ChartView {...props} selected={{ version: testVersion }} />);
  expect(wrapper.contains(<h2>App Version</h2>)).toBe(false);
  expect(wrapper.contains(<h2>Home</h2>)).toBe(false);
  expect(wrapper.contains(<h2>Related</h2>)).toBe(false);
});

it("renders the app version when set", () => {
  const v = testVersion as IChartVersion;
  v.attributes.app_version = "1.2.3-appversion";
  const wrapper = shallow(<ChartView {...props} selected={{ version: v }} />);
  expect(wrapper.contains(<h2>App Version</h2>)).toBe(true);
  expect(wrapper.contains(<div>1.2.3-appversion</div>)).toBe(true);
});

it("renders the home link when set", () => {
  const v = testVersion as IChartVersion;
  v.relationships.chart.data.home = "https://example.com";
  const wrapper = shallow(<ChartView {...props} selected={{ version: v }} />);
  expect(wrapper.contains(<h2>Home</h2>)).toBe(true);
  expect(
    wrapper.contains(
      <a href="https://example.com" target="_blank">
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
  }> = [
    {
      expected: true,
      name: "stable Helm repo",
      repoURL: "https://kubernetes-charts.storage.googleapis.com",
    },
    {
      expected: true,
      name: "incubator Helm repo",
      repoURL: "https://kubernetes-charts-incubator.storage.googleapis.com",
    },
    { name: "random Helm repo", repoURL: "https://examplerepo.com", expected: false },
  ];

  for (const t of tests) {
    it(`for ${t.name}`, () => {
      v.relationships.chart.data.repo.url = t.repoURL;
      const wrapper = shallow(<ChartView {...props} selected={{ version: v }} />);
      const chartMaintainers = wrapper.find(ChartMaintainers);
      expect(chartMaintainers.props().githubIDAsNames).toBe(t.expected);
    });
  }
});

it("renders the sources links when set", () => {
  const v = testVersion as IChartVersion;
  v.relationships.chart.data.sources = ["https://example.com", "https://example2.com"];
  const wrapper = shallow(<ChartView {...props} selected={{ version: v }} />);
  expect(wrapper.contains(<h2>Related</h2>)).toBe(true);
  expect(
    wrapper.contains(
      <a href="https://example.com" target="_blank">
        {"https://example.com"}
      </a>,
    ),
  ).toBe(true);
  expect(
    wrapper.contains(
      <a href="https://example2.com" target="_blank">
        {"https://example2.com"}
      </a>,
    ),
  ).toBe(true);
});

describe("renders errors", () => {
  it("renders a not found error if it exists", () => {
    const wrapper = shallow(<ChartView {...props} selected={{ error: new NotFoundError() }} />);
    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper.find(ErrorSelector).html()).toContain(`Chart ${props.chartID} not found`);
  });
  it("renders a generic error if it exists", () => {
    const wrapper = shallow(<ChartView {...props} selected={{ error: new Error() }} />);
    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper.find(ErrorSelector).html()).toContain("Sorry! Something went wrong");
  });
});
