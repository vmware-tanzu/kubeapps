import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import itBehavesLike from "../../shared/specs";
import { IChart, IChartState } from "../../shared/types";
import { CardGrid } from "../Card";
import { MessageAlert } from "../ErrorAlert";
import PageHeader from "../PageHeader";
import SearchFilter from "../SearchFilter";
import Catalog from "./Catalog";

const defaultChartState = {
  isFetching: false,
  selected: {} as IChartState["selected"],
  items: [],
  updatesInfo: {},
} as IChartState;
const defaultProps = {
  charts: defaultChartState,
  repo: "stable",
  filter: "",
  fetchCharts: jest.fn(),
  pushSearchFilter: jest.fn(),
};

it("propagates the filter from the props", () => {
  const wrapper = shallow(<Catalog {...defaultProps} filter="foo" />);
  expect(wrapper.state("filter")).toBe("foo");
});

it("reloads charts when the repo changes", () => {
  const fetchCharts = jest.fn();
  const wrapper = shallow(<Catalog {...defaultProps} fetchCharts={fetchCharts} />);
  wrapper.setProps({ ...defaultProps, repo: "bitnami" });
  expect(fetchCharts.mock.calls.length).toBe(2);
  expect(fetchCharts.mock.calls[1]).toEqual(["bitnami"]);
});

describe("renderization", () => {
  context("when no charts", () => {
    it("should render an error", () => {
      const wrapper = shallow(<Catalog {...defaultProps} />);
      expect(wrapper.find(MessageAlert)).toExist();
      expect(wrapper.find(".Catalog")).not.toExist();
      expect(
        wrapper
          .find(MessageAlert)
          .children()
          .text(),
      ).toContain("Manage your Helm chart repositories");
      expect(wrapper).toMatchSnapshot();
    });
  });

  context("when fetching apps", () => {
    itBehavesLike("aLoadingComponent", {
      component: Catalog,
      props: { ...defaultProps, charts: { isFetching: true, items: [] } },
    });
  });

  context("when charts available", () => {
    const chartState = {
      isFetching: false,
      selected: {} as IChartState["selected"],
      items: [
        { id: "foo", attributes: { description: "" } } as IChart,
        { id: "bar", attributes: { description: "" } } as IChart,
      ],
    } as IChartState;

    it("should render the list of charts", () => {
      const wrapper = shallow(<Catalog {...defaultProps} charts={chartState} />);

      expect(wrapper.find(MessageAlert)).not.toExist();
      expect(wrapper.find(PageHeader)).toExist();
      expect(wrapper.find(SearchFilter)).toExist();

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
      // Filter "foo" app
      const wrapper = shallow(<Catalog {...defaultProps} charts={chartState} filter="foo" />);

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
});
