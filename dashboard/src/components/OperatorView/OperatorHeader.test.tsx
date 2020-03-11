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
};

it("renders the header", () => {
  const wrapper = shallow(<OperatorHeader {...defaultProps} />);
  expect(wrapper).toMatchSnapshot();
});
