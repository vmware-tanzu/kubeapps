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
import CatalogItem from "./CatalogItem";

const defaultChartState = {
  isFetching: false,
  selected: {} as IChartState["selected"],
  deployed: {} as IChartState["deployed"],
  items: [],
  updatesInfo: {},
} as IChartState;
const defaultProps = {
  charts: defaultChartState,
  repo: "stable",
  filter: "",
  fetchCharts: jest.fn(),
  pushSearchFilter: jest.fn(),
  namespace: "kubeapps",
  csvs: [],
  getCSVs: jest.fn(),
  operators: [],
  getOperators: jest.fn(),
  featureFlags: { operators: false },
};

it("propagates the filter from the props", () => {
  const wrapper = shallow(<Catalog {...defaultProps} filter="foo" />);
  expect(wrapper.state("filter")).toBe("foo");
});

it("reloads charts when the repo changes", () => {
  const fetchCharts = jest.fn();
  const wrapper = shallow(<Catalog {...defaultProps} fetchCharts={fetchCharts} />);
  wrapper.setProps({ ...defaultProps, fetchCharts, repo: "bitnami" });
  expect(fetchCharts.mock.calls.length).toBe(2);
  expect(fetchCharts.mock.calls[1]).toEqual(["bitnami"]);
});

it("updates the filter from props", () => {
  const wrapper = shallow(<Catalog {...defaultProps} />);
  wrapper.setProps({ filter: "foo" });
  expect(wrapper.state("filter")).toBe("foo");
});

it("keeps the filter from the state", () => {
  const wrapper = shallow(<Catalog {...defaultProps} />);
  expect(wrapper.state("filter")).toBe("");
  wrapper.setState({ filter: "foo" });
  expect(wrapper.state("filter")).toBe("foo");
});

describe("componentDidMount", () => {
  it("retrieves operators and csvs in the namespace", () => {
    const getOperators = jest.fn();
    const getCSVs = jest.fn();
    const namespace = "foo";
    shallow(
      <Catalog
        {...defaultProps}
        getOperators={getOperators}
        getCSVs={getCSVs}
        namespace={namespace}
        featureFlags={{ operators: true }}
      />,
    );
    expect(getCSVs).toHaveBeenCalledWith(namespace);
    expect(getOperators).toHaveBeenCalledWith(namespace);
  });
});

describe("componentDidUpdate", () => {
  it("re-fetches operators and csvs if the namespace changes", () => {
    const getOperators = jest.fn();
    const getCSVs = jest.fn();
    const wrapper = shallow(
      <Catalog
        {...defaultProps}
        getOperators={getOperators}
        getCSVs={getCSVs}
        featureFlags={{ operators: true }}
      />,
    );
    wrapper.setProps({ namespace: "a-different-one" });
    expect(getCSVs).toHaveBeenCalledWith("a-different-one");
    expect(getOperators).toHaveBeenCalledWith("a-different-one");
  });
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
        {
          id: "foo",
          attributes: { description: "", repo: { name: "foo" } },
          relationships: { latestChartVersion: { data: { app_version: "v1.0.0" } } },
        } as IChart,
        {
          id: "bar",
          attributes: { description: "", repo: { name: "bar" } },
          relationships: { latestChartVersion: { data: { app_version: "v2.0.0" } } },
        } as IChart,
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
      const expectedItem1 = {
        description: "",
        id: "foo",
        namespace: "kubeapps",
        repoName: "foo",
        type: "chart",
        version: "v1.0.0",
      };
      const expectedItem2 = {
        description: "",
        id: "bar",
        namespace: "kubeapps",
        repoName: "bar",
        type: "chart",
        version: "v2.0.0",
      };
      expect(
        cardGrid
          .children()
          .at(0)
          .props().item,
      ).toEqual(expectedItem1);
      expect(
        cardGrid
          .children()
          .at(1)
          .props().item,
      ).toEqual(expectedItem2);
      expect(wrapper).toMatchSnapshot();
    });

    it("should filter apps", () => {
      // Filter "foo" app
      const wrapper = shallow(<Catalog {...defaultProps} charts={chartState} filter="foo" />);

      const cardGrid = wrapper.find(CardGrid);
      expect(cardGrid).toExist();
      expect(cardGrid.children().length).toBe(1);
      const expectedItem = {
        description: "",
        id: "foo",
        namespace: "kubeapps",
        repoName: "foo",
        type: "chart",
        version: "v1.0.0",
      };
      expect(
        cardGrid
          .children()
          .at(0)
          .props().item,
      ).toEqual(expectedItem);
    });

    describe("when operators available", () => {
      const operators = [
        {
          metadata: {
            name: "foo-op",
          },
          status: {
            defaultChannel: "foo",
            channels: [
              {
                name: "foo",
                currentCSV: "test-csv",
              },
            ],
          },
        } as any,
      ];
      const csvs = [
        {
          metadata: {
            name: "test-csv",
          },
          spec: {
            customresourcedefinitions: {
              owned: [
                {
                  name: "foo-cluster",
                  displayName: "Foo Cluster",
                  version: "v1.0.0",
                  description: "a meaningful description",
                },
              ],
            },
          },
        } as any,
      ];

      it("show render the list of charts and operators", () => {
        const wrapper = shallow(
          <Catalog {...defaultProps} charts={chartState} operators={operators} csvs={csvs} />,
        );
        const cardGrid = wrapper.find(CardGrid);
        expect(cardGrid).toExist();
        expect(cardGrid.children().length).toBe(chartState.items.length + csvs.length);

        const expectedItem = {
          csv: "test-csv",
          description: "a meaningful description",
          icon: "api/v1/namespaces/kubeapps/operator/foo-op/logo",
          id: "foo-cluster",
          name: "Foo Cluster",
          namespace: "kubeapps",
          operator: "foo-op",
          type: "operator",
          version: "v1.0.0",
        };
        const csvCard = cardGrid
          .find(CatalogItem)
          .findWhere(c => c.prop("item").id === "foo-cluster");
        expect(csvCard).toExist();
        expect(csvCard.prop("item")).toMatchObject(expectedItem);
      });

      it("should filter out charts or operators when requested", () => {
        const wrapper = shallow(
          <Catalog {...defaultProps} charts={chartState} operators={operators} csvs={csvs} />,
        );

        wrapper.setState({ listCharts: false });
        let cardGrid = wrapper.find(CardGrid);
        expect(cardGrid).toExist();
        expect(cardGrid.children().length).toBe(csvs.length);

        wrapper.setState({ listOperators: false, listCharts: true });
        cardGrid = wrapper.find(CardGrid);
        expect(cardGrid).toExist();
        expect(cardGrid.children().length).toBe(chartState.items.length);
      });
    });
  });
});
