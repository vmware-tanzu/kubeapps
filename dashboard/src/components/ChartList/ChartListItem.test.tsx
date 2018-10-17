import { shallow } from "enzyme";
import * as React from "react";

import { IChart, IRepo } from "../../shared/types";
import { CardIcon } from "../Card";
import InfoCard from "../InfoCard";
import ChartListItem from "./ChartListItem";

jest.mock("../../placeholder.png", () => "placeholder.png");

const defaultChart = {
  id: "foo",
  attributes: {
    description: "",
    keywords: [""],
    maintainers: [{ name: "" }],
    sources: [""],
    icon: "icon.png",
    name: "foo",
    repo: {} as IRepo,
  },
  relationships: {
    latestChartVersion: {
      data: {
        app_version: "1.0.0",
      },
    },
  },
} as IChart;

it("should render an item", () => {
  const wrapper = shallow(<ChartListItem chart={defaultChart} />);
  expect(wrapper).toMatchSnapshot();
});

it("should use the default placeholder for the icon if it doesn't exist", () => {
  const chartWithoutIcon = { ...defaultChart };
  chartWithoutIcon.attributes.icon = undefined;
  const wrapper = shallow(<ChartListItem chart={chartWithoutIcon} />);
  // Importing an image returns "undefined"
  expect(
    wrapper
      .find(InfoCard)
      .shallow()
      .find(CardIcon)
      .prop("src"),
  ).toBe(undefined);
});

it("should place a dash if the version is not avaliable", () => {
  const chartWithoutVersion = { ...defaultChart };
  chartWithoutVersion.relationships.latestChartVersion.data.app_version = "";
  const wrapper = shallow(<ChartListItem chart={chartWithoutVersion} />);
  expect(
    wrapper
      .find(InfoCard)
      .shallow()
      .find(".type-color-light-blue")
      .text(),
  ).toBe("-");
});
