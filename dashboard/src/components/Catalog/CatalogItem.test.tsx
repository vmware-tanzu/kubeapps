import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";

import { IRepo } from "../../shared/types";
import { CardIcon } from "../Card";
import InfoCard from "../InfoCard";
import CatalogItem, {
  ICatalogItemProps,
  IChartCatalogItem,
  IOperatorCatalogItem,
} from "./CatalogItem";

jest.mock("../../placeholder.png", () => "placeholder.png");

const defaultItem = {
  id: "foo1",
  name: "foo",
  version: "1.0.0",
  description: "",
  type: "chart",
  repo: {
    name: "repo-name",
    namespace: "repo-namespace",
  } as IRepo,
  cluster: "default",
  namespace: "repo-namespace",
  icon: "icon.png",
} as IChartCatalogItem;

const defaultProps: ICatalogItemProps = {
  item: defaultItem,
  type: "chart",
};

const defaultStore = getStore({});

it("should render a chart item in a namespace", () => {
  const wrapper = mountWrapper(defaultStore, <CatalogItem {...defaultProps} />);
  // Can't shallow render connected components for easy snapshotting :/
  // https://github.com/enzymejs/enzyme/issues/2202
  expect(wrapper.find(InfoCard)).toMatchSnapshot();
});

it("should render a global chart item in a namespace", () => {
  const props = {
    ...defaultProps,
    item: {
      ...defaultItem,
      repo: {
        name: "repo-name",
        namespace: "kubeapps",
      } as IRepo,
    },
  };
  const wrapper = mountWrapper(defaultStore, <CatalogItem {...props} />);
  expect(wrapper.find(InfoCard)).toMatchSnapshot();
});

it("should use the default placeholder for the icon if it doesn't exist", () => {
  const props = {
    ...defaultProps,
    item: {
      ...defaultItem,
      icon: undefined,
    },
  };
  const wrapper = mountWrapper(defaultStore, <CatalogItem {...props} />);
  // Importing an image returns "undefined"
  expect(wrapper.find(CardIcon).prop("src")).toBe(undefined);
});

it("should place a dash if the version is not avaliable", () => {
  const props = {
    ...defaultProps,
    item: {
      ...defaultItem,
      version: "",
    },
  };
  const wrapper = mountWrapper(defaultStore, <CatalogItem {...props} />);
  expect(wrapper.find(".type-color-light-blue").text()).toBe("-");
});

it("show the chart description", () => {
  const props = {
    ...defaultProps,
    item: {
      ...defaultItem,
      description: "This is a description",
    },
  };
  const wrapper = mountWrapper(defaultStore, <CatalogItem {...props} />);
  expect(wrapper.find(".ListItem__content__description").text()).toBe(props.item.description);
});

context("when the description is too long", () => {
  it("trims the description", () => {
    const props = {
      ...defaultProps,
      item: {
        ...defaultItem,
        description:
          "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum ultrices velit leo, quis pharetra mi vestibulum quis.",
      },
    };
    const wrapper = mountWrapper(defaultStore, <CatalogItem {...props} />);
    expect(wrapper.find(".ListItem__content__description").text()).toMatch(/\.\.\.$/);
  });
});

context("when the item is a catalog", () => {
  const catalogItem = {
    ...defaultItem,
    csv: "foo-cluster",
  } as IOperatorCatalogItem;

  it("shows the proper tag", () => {
    const wrapper = shallow(<CatalogItem item={catalogItem} type={"operator"} />);
    expect((wrapper.find(InfoCard).prop("tag1Content") as JSX.Element).props.children).toEqual(
      "foo-cluster",
    );
  });

  it("has the proper link", () => {
    const wrapper = shallow(<CatalogItem item={catalogItem} type={"operator"} />);
    expect(wrapper.find(InfoCard).prop("link")).toEqual(
      `/c/default/ns/${defaultItem.namespace}/operators-instances/new/foo-cluster/foo1`,
    );
  });
});
