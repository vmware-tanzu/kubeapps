import { shallow } from "enzyme";
import * as React from "react";
import { IClusterState } from "../../reducers/namespace";
import { app } from "../../shared/url";
import Header from "./Header";


const defaultProps = {
  authenticated: true,
  fetchNamespaces: jest.fn(),
  logout: jest.fn(),
  cluster: {
    currentNamespace: "default",
    namespaces: ["default", "other"],
  } as IClusterState,
  defaultNamespace: "kubeapps-user",
  pathname: "",
  push: jest.fn(),
  setNamespace: jest.fn(),
  createNamespace: jest.fn(),
  getNamespace: jest.fn(),
};
it("renders the header links and titles", () => {
  const wrapper = shallow(<Header {...defaultProps} />);
  const menubar = wrapper.find(".header__nav__menu").first();
  const items = menubar.children().map(p => p.props().children.props);
  const expectedItems = [
    { children: "Applications", to: app.apps.list("default") },
    { children: "Catalog", to: app.catalog("default") },
    { children: "Service Instances (alpha)", to: app.servicesInstances("default") },
  ];
  expect(items.length).toEqual(expectedItems.length);
  expectedItems.forEach((expectedItem, index) => {
    expect(expectedItem.children).toBe(items[index].children);
    expect(expectedItem.to).toBe(items[index].to);
  });
});

describe("settings", () => {
  it("renders settings", () => {
    const wrapper = shallow(
      <Header {...defaultProps} featureFlags={{ operators: false }} />,
    );
    const settingsbar = wrapper.find(".header__nav__submenu").first();
    const items = settingsbar.find("NavLink").map(p => p.props());
    const expectedItems = [
      { children: "App Repositories", to: "/config/ns/default/repos" },
      { children: "Service Brokers", to: "/config/brokers" },
    ];
    items.forEach((item, index) => {
      expect(item.children).toBe(expectedItems[index].children);
      expect(item.to).toBe(expectedItems[index].to);
    });
  });

  it("renders operators link", () => {
    const wrapper = shallow(
      <Header {...defaultProps} featureFlags={{ operators: true }} />,
    );
    const settingsbar = wrapper.find(".header__nav__submenu").first();
    const items = settingsbar.find("NavLink").map(p => p.props());
    const expectedItems = [
      { children: "App Repositories", to: "/config/ns/default/repos" },
      { children: "Service Brokers", to: "/config/brokers" },
      { children: "Operators", to: "/ns/default/operators" },
    ];
    items.forEach((item, index) => {
      expect(item.children).toBe(expectedItems[index].children);
      expect(item.to).toBe(expectedItems[index].to);
    });
  });
});

it("updates state when the path changes", () => {
  const wrapper = shallow(<Header {...defaultProps} />);
  wrapper.setState({ configOpen: true, mobileOpne: true });
  wrapper.setProps({ pathname: "foo" });
  expect(wrapper.state()).toMatchObject({ configOpen: false, mobileOpen: false });
});

it("renders the namespace switcher", () => {
  const wrapper = shallow(<Header {...defaultProps} />);

  const namespaceSelector = wrapper.find("NamespaceSelector");

  expect(namespaceSelector).toExist();
  expect(namespaceSelector.props()).toEqual(
    expect.objectContaining({
      defaultNamespace: defaultProps.defaultNamespace,
      cluster: defaultProps.cluster,
    }),
  );
});

it("call setNamespace and getNamespace when selecting a namespace", () => {
  const setNamespace = jest.fn();
  const createNamespace = jest.fn();
  const getNamespace = jest.fn();
  const cluster = {
    currentNamespace: "foo",
    namespaces: ["foo", "bar"],
  };
  const wrapper = shallow(
    <Header
      {...defaultProps}
      setNamespace={setNamespace}
      cluster={cluster}
      createNamespace={createNamespace}
      getNamespace={getNamespace}
    />,
  );

  const namespaceSelector = wrapper.find("NamespaceSelector");
  expect(namespaceSelector).toExist();
  const onChange = namespaceSelector.prop("onChange") as (ns: any) => void;
  onChange("bar");

  expect(setNamespace).toHaveBeenCalledWith("bar");
  expect(getNamespace).toHaveBeenCalledWith("bar");
  expect(createNamespace).not.toHaveBeenCalled();
});

it("doesn't call getNamespace when selecting all namespaces", () => {
  const setNamespace = jest.fn();
  const getNamespace = jest.fn();
  const cluster = {
    currentNamespace: "foo",
    namespaces: ["foo", "bar"],
  };
  const wrapper = shallow(
    <Header
      {...defaultProps}
      setNamespace={setNamespace}
      cluster={cluster}
      getNamespace={getNamespace}
    />,
  );

  const namespaceSelector = wrapper.find("NamespaceSelector");
  expect(namespaceSelector).toExist();
  const onChange = namespaceSelector.prop("onChange") as (ns: any) => void;
  onChange("_all");

  expect(setNamespace).toHaveBeenCalledWith("_all");
  expect(getNamespace).not.toHaveBeenCalled();
});
