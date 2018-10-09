import { shallow } from "enzyme";
import * as React from "react";

import AccessURLItem from "./AccessURLItem";
import { IURLItem } from "./IURLItem";

it("should show only a message (without a link) if the item is not a link", () => {
  const item = { name: "foo", isLink: false, URLs: ["Pending"] } as IURLItem;
  const wrapper = shallow(<AccessURLItem URLItem={item} />);
  expect(wrapper.text()).toContain("Pending");
  expect(wrapper).toMatchSnapshot();
  const link = wrapper.find(".ServiceItem").find("a");
  expect(link).not.toExist();
});

it("should show only an URL with a link", () => {
  const item = {
    name: "foo",
    isLink: true,
    URLs: ["http://1.2.3.4:8080", "https://foo.bar"],
  } as IURLItem;
  const wrapper = shallow(<AccessURLItem URLItem={item} />);
  expect(wrapper.text()).toContain("http://1.2.3.4:8080");
  expect(wrapper.text()).toContain("https://foo.bar");
  expect(wrapper).toMatchSnapshot();
  const link = wrapper.find(".ServiceItem").find("a");
  expect(link).toExist();
});
