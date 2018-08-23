import { shallow } from "enzyme";
import * as React from "react";

import ErrorPageHeader from "./ErrorAlertHeader";
import MessageAlert from "./MessageAlert";

it("renders the heading passed to it", () => {
  const wrapper = shallow(
    <MessageAlert>
      <div>test</div>
    </MessageAlert>,
  );
  expect(wrapper.text()).toContain("test");
  expect(wrapper.find(".message__content").text()).toContain("test");
});

it("should skip the header and children content if not set", () => {
  const wrapper = shallow(<MessageAlert />);
  expect(wrapper.text()).toEqual("");
  expect(wrapper.find(".message__content .margin-l-enormous").exists()).toBe(false);
});

it("should include a header if set", () => {
  const wrapper = shallow(<MessageAlert header="foo" />);
  expect(
    wrapper
      .find(ErrorPageHeader)
      .children()
      .text(),
  ).toBe("foo");
});

it("should include a type if set", () => {
  const wrapper = shallow(<MessageAlert level="foo" />);
  expect(wrapper.find(".alert-foo").exists()).toBe(true);
});

it("should not set margin if header doesn't exists", () => {
  const wrapper = shallow(
    <MessageAlert>
      <div>foo</div>
    </MessageAlert>,
  );
  expect(wrapper.find(".message__content .margin-l-enormous").exists()).toBe(false);
});
