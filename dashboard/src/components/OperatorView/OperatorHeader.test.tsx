import { shallow } from "enzyme";
import * as React from "react";
import OperatorHeader from "./OperatorHeader";

const defaultProps = {
  id: "foo",
  icon: "/path/to/icon.png",
  description: "this is a description",
  cluster: "default",
  namespace: "kubeapps",
  version: "1.0.0",
  provider: "Kubeapps",
  namespaced: false,
  push: jest.fn(),
};

it("renders the header", () => {
  const wrapper = shallow(<OperatorHeader {...defaultProps} />);
  expect(wrapper).toMatchSnapshot();
});

it("omits the button", () => {
  const wrapper = shallow(<OperatorHeader {...defaultProps} hideButton={true} />);
  expect(wrapper.find("button")).not.toExist();
});

it("disables the button", () => {
  const wrapper = shallow(<OperatorHeader {...defaultProps} disableButton={true} />);
  const button = wrapper.find("button");
  expect(button).toExist();
  expect(button.prop("disabled")).toBe(true);
  expect(button.text()).toBe("Deployed");
});
