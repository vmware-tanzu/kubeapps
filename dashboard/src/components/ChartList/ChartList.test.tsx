import { shallow } from "enzyme";
import * as React from "react";

import { IChart, IChartState } from "../../shared/types";
import { NotFoundErrorAlert } from "../ErrorAlert";
import PageHeader from "../PageHeader";
import SearchFilter from "../SearchFilter";
import ChartList from "./ChartList";

const defaultChartState = {
  isFetching: false,
  selected: {} as IChartState["selected"],
  items: [],
} as IChartState;
const defaultProps = {
  charts: defaultChartState,
  repo: "stable",
  filter: "",
  fetchCharts: jest.fn(),
  pushSearchFilter: jest.fn(),
};

it("propagates the filter from the props", () => {
  const wrapper = shallow(<ChartList {...defaultProps} filter="foo" />);
  expect(wrapper.state().filter).toBe("foo");
});

describe("renderization", () => {
  it("should render an error if no charts are found", () => {
    const wrapper = shallow(<ChartList {...defaultProps} />);
    expect(wrapper.find(NotFoundErrorAlert).exists()).toBe(true);
    expect(wrapper.find(".ChartList").exists()).toBe(false);
    expect(
      wrapper
        .find(NotFoundErrorAlert)
        .children()
        .text(),
    ).toContain("Manage your Helm chart repositories");
  });

  it("should render a loading message if is still fetching charts", () => {
    const chartState = {
      isFetching: true,
      selected: {} as IChartState["selected"],
      items: [],
    } as IChartState;
    const wrapper = shallow(<ChartList {...defaultProps} charts={chartState} />);
    expect(wrapper.find(NotFoundErrorAlert).exists()).toBe(false);
    expect(wrapper.text()).toContain("Loading");
  });

  it("should render the list of charts charts", () => {
    const chartState = {
      isFetching: false,
      selected: {} as IChartState["selected"],
      items: [{ id: "foo" } as IChart, { id: "bar" } as IChart],
    } as IChartState;
    const wrapper = shallow(<ChartList {...defaultProps} charts={chartState} />);
    expect(wrapper.find(NotFoundErrorAlert).exists()).toBe(false);
    expect(wrapper.find(PageHeader).exists()).toBe(true);
    expect(wrapper.find(SearchFilter).exists()).toBe(true);
    expect(wrapper.text()).not.toContain("Loading");
    // Check list
  });
});
