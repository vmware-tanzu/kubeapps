import { shallow } from "enzyme";
import * as React from "react";
import NewNamespace from "./NewNamespace";

const defaultProps = {
  modalIsOpen: true,
  namespace: "foo",
  onConfirm: jest.fn(),
  closeModal: jest.fn(),
  onChange: jest.fn(),
  error: new Error("boom!"),
};

it("renders the snapshot", () => {
  const wrapper = shallow(<NewNamespace {...defaultProps} />);
  expect(wrapper).toMatchSnapshot();
});
