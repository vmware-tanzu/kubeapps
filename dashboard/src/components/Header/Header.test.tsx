import { shallow } from "enzyme";
import * as React from "react";

import { INamespaceState } from "../../reducers/namespace";
import Header from "./Header";

const defaultProps = {
  authenticated: true,
  fetchNamespaces: jest.fn(),
  logout: jest.fn(),
  namespace: {
    current: "",
    namespaces: [],
  } as INamespaceState,
  defaultNamespace: "kubeapps-user",
  pathname: "",
  push: jest.fn(),
  setNamespace: jest.fn(),
  logoutUrl: "",
};
it("renders the header links and titles", () => {
  const wrapper = shallow(<Header {...defaultProps} />);
  const menubar = wrapper.find(".header__nav__menu").first();
  const items = menubar.children().map(p => p.props().children.props);
  const expectedItems = [
    { children: "Applications", to: "/apps" },
    { children: "Catalog", to: "/catalog" },
    { children: "Service Instances (alpha)", to: "/services/instances" },
  ];
  items.forEach((item, index) => {
    expect(item.children).toBe(expectedItems[index].children);
    expect(item.to).toBe(expectedItems[index].to);
  });
});

it("renders the namespace switcher", () => {
  const wrapper = shallow(<Header {...defaultProps} />);

  const namespaceSelector = wrapper.find("NamespaceSelector");

  expect(namespaceSelector).toExist();
  expect(namespaceSelector.props()).toEqual(
    expect.objectContaining({
      defaultNamespace: defaultProps.defaultNamespace,
      namespace: defaultProps.namespace,
    }),
  );
});

it("renders an anchor logout link when logoutUrl is set", () => {
  const wrapper = shallow(<Header {...defaultProps} logoutUrl="http://localhost/log/out" />);
  const link = wrapper.find('a[href="http://localhost/log/out"]');
  expect(link.length).toBe(1);
  expect(link.text()).toContain("Logout");
});
