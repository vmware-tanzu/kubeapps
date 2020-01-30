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
  createNamespace: jest.fn(),
  getNamespace: jest.fn(),
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
      namespace: defaultProps.namespace,
    }),
  );
});

it("call setNamespace and getNamespace when selecting a namespace", () => {
  const setNamespace = jest.fn();
  const createNamespace = jest.fn();
  const getNamespace = jest.fn();
  const namespace = {
    current: "foo",
    namespaces: ["foo", "bar"],
  };
  const wrapper = shallow(
    <Header
      {...defaultProps}
      setNamespace={setNamespace}
      namespace={namespace}
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
