import { shallow } from "enzyme";
import * as React from "react";

import { IChart, IChartState } from "../../shared/types";
import { CardGrid } from "../Card";
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
    expect(wrapper.find(NotFoundErrorAlert)).toExist();
    expect(wrapper.find(".ChartList")).not.toExist();
    expect(
      wrapper
        .find(NotFoundErrorAlert)
        .children()
        .text(),
    ).toContain("Manage your Helm chart repositories");
    expect(wrapper).toMatchSnapshot();
  });

  it("should render a loading message if is still fetching charts", () => {
    const chartState = {
      isFetching: true,
      selected: {} as IChartState["selected"],
      items: [],
    } as IChartState;
    const wrapper = shallow(<ChartList {...defaultProps} charts={chartState} />);

    expect(wrapper.find(NotFoundErrorAlert)).not.toExist();
    expect(wrapper.text()).toContain("Loading");
  });

  it("should render the list of charts", () => {
    const chartState = {
      isFetching: false,
      selected: {} as IChartState["selected"],
      items: [{ id: "foo", attributes: {} } as IChart, { id: "bar", attributes: {} } as IChart],
    } as IChartState;
    const wrapper = shallow(<ChartList {...defaultProps} charts={chartState} />);

    expect(wrapper.find(NotFoundErrorAlert)).not.toExist();
    expect(wrapper.find(PageHeader)).toExist();
    expect(wrapper.find(SearchFilter)).toExist();
    expect(wrapper.text()).not.toContain("Loading");

    const cardGrid = wrapper.find(CardGrid);
    expect(cardGrid).toExist();
    expect(cardGrid.children().length).toBe(chartState.items.length);
    expect(
      cardGrid
        .children()
        .at(0)
        .props().chart,
    ).toEqual(chartState.items[0]);
    expect(
      cardGrid
        .children()
        .at(1)
        .props().chart,
    ).toEqual(chartState.items[1]);
    expect(wrapper).toMatchSnapshot();
  });

  it("should filter apps", () => {
    const chartState = {
      isFetching: false,
      selected: {} as IChartState["selected"],
      items: [{ id: "foo", attributes: {} } as IChart, { id: "bar", attributes: {} } as IChart],
    } as IChartState;
    // Filter "foo" app
    const wrapper = shallow(<ChartList {...defaultProps} charts={chartState} filter="foo" />);

    const cardGrid = wrapper.find(CardGrid);
    expect(cardGrid).toExist();
    expect(cardGrid.children().length).toBe(1);
    expect(
      cardGrid
        .children()
        .at(0)
        .props().chart,
    ).toEqual(chartState.items[0]);
  });
});
