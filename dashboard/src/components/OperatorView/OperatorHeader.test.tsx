import { shallow } from "enzyme";
import * as React from "react";
import OperatorHeader from "./OperatorHeader";

const defaultProps = {
  id: "foo",
  icon: "/path/to/icon.png",
  description: "this is a description",
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
