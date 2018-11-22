import { shallow } from "enzyme";
import * as React from "react";

import { IServiceInstance } from "shared/ServiceInstance";
import DeprovisionButton from "./DeprovisionButton";

const defaultProps = {
  disabled: false,
  instance: {} as IServiceInstance,
  deprovision: jest.fn(),
};

it("shows a button", () => {
  const wrapper = shallow(<DeprovisionButton {...defaultProps} />);
  expect(wrapper.find(".button")).toExist();
  expect(wrapper.find(".button").text()).toBe("Deprovision");
  expect(wrapper).toMatchSnapshot();
});

it("disables the button with a prop", () => {
  const wrapper = shallow(<DeprovisionButton {...defaultProps} disabled={true} />);
  expect(wrapper.find(".button").prop("disabled")).toBe(true);
});

it("disables the button while deprovisioning", () => {
  const wrapper = shallow(<DeprovisionButton {...defaultProps} />);
  wrapper.setState({ isDeprovisioning: true });
  expect(wrapper.find(".button").prop("disabled")).toBe(true);
  expect(wrapper.text()).toContain("Deprovisioning...");
});
