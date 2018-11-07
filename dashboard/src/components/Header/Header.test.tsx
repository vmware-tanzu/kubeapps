import { shallow } from "enzyme";
import * as React from "react";

import { INamespaceState } from "../../reducers/namespace";
import Header from "./Header";

it("renders the header links and titles", () => {
  const wrapper = shallow(
    <Header
      authenticated={true}
      fetchNamespaces={jest.fn()}
      logout={jest.fn()}
      namespace={
        {
          current: "default",
          namespaces: ["default"],
        } as INamespaceState
      }
      pathname=""
      push={jest.fn()}
      setNamespace={jest.fn()}
    />,
  );
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
