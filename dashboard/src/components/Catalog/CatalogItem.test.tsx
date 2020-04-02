import { shallow } from "enzyme";
import context from "jest-plugin-context";
import { cloneDeep } from "lodash";
import * as React from "react";

import { CardIcon } from "../Card";
import InfoCard from "../InfoCard";
import CatalogItem, { ICatalogItem } from "./CatalogItem";

jest.mock("../../placeholder.png", () => "placeholder.png");

const defaultItem = {
  id: "foo1",
  name: "foo",
  version: "1.0.0",
  description: "",
  type: "chart",
  namespace: "kubeapps",
  icon: "icon.png",
} as ICatalogItem;

it("should render an item", () => {
  const wrapper = shallow(<CatalogItem item={defaultItem} />);
  expect(wrapper).toMatchSnapshot();
});

it("should use the default placeholder for the icon if it doesn't exist", () => {
  const chartWithoutIcon = cloneDeep(defaultItem);
  chartWithoutIcon.icon = undefined;
  const wrapper = shallow(<CatalogItem item={chartWithoutIcon} />);
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
  const chartWithoutVersion = cloneDeep(defaultItem);
  chartWithoutVersion.version = "";
  const wrapper = shallow(<CatalogItem item={chartWithoutVersion} />);
  expect(
    wrapper
      .find(InfoCard)
      .shallow()
      .find(".type-color-light-blue")
      .text(),
  ).toBe("-");
});

it("show the chart description", () => {
  const chartWithDescription = cloneDeep(defaultItem);
  chartWithDescription.description = "This is a description";
  const wrapper = shallow(<CatalogItem item={chartWithDescription} />);
  expect(
    wrapper
      .find(InfoCard)
      .shallow()
      .find(".ListItem__content__description")
      .text(),
  ).toBe(chartWithDescription.description);
});

context("when the description is too long", () => {
  it("trims the description", () => {
    const chartWithDescription = cloneDeep(defaultItem);
    chartWithDescription.description =
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum ultrices velit leo, quis pharetra mi vestibulum quis.";
    const wrapper = shallow(<CatalogItem item={chartWithDescription} />);
    expect(
      wrapper
        .find(InfoCard)
        .shallow()
        .find(".ListItem__content__description")
        .text(),
    ).toMatch(/\.\.\.$/);
  });
});

context("when the item is a catalog", () => {
  const catalogItem = {
    ...defaultItem,
    csv: "foo-cluster",
    type: "operator",
  } as ICatalogItem;

  it("shows the proper tag", () => {
    const wrapper = shallow(<CatalogItem item={catalogItem} />);
    expect((wrapper.find(InfoCard).prop("tag1Content") as JSX.Element).props.children).toEqual(
      "foo-cluster",
    );
  });

  it("has the proper link", () => {
    const wrapper = shallow(<CatalogItem item={catalogItem} />);
    expect(wrapper.find(InfoCard).prop("link")).toEqual(
      "/ns/kubeapps/operators-instances/new/foo-cluster/foo1",
    );
  });
});
