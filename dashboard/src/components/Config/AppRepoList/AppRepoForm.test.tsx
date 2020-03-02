import { shallow } from "enzyme";
import * as React from "react";
import UnexpectedErrorPage from "../../ErrorAlert/UnexpectedErrorAlert";
import { AppRepoForm } from "./AppRepoForm";

const defaultProps = {
  install: jest.fn(),
  validate: jest.fn(),
  isFetching: false,
};

it("should render the repo form", () => {
  const wrapper = shallow(<AppRepoForm {...defaultProps} />);
  expect(wrapper).toMatchSnapshot();
});

it("disables the submit button while fetching", () => {
  const wrapper = shallow(<AppRepoForm {...defaultProps} isFetching={true} />);
  expect(wrapper.find("button").prop("disabled")).toBe(true);
});

it("should show a validation error", () => {
  const wrapper = shallow(<AppRepoForm {...defaultProps} validationError={new Error("Boom!")} />);
  expect(
    wrapper
      .find(UnexpectedErrorPage)
      .dive()
      .text(),
  ).toContain("Boom!");
});

it("should call the install method when the validation success", async () => {
  const validate = jest.fn(() => true);
  const install = jest.fn(() => true);
  const wrapper = shallow(<AppRepoForm {...defaultProps} validate={validate} install={install} />);
  const button = wrapper.find("form");
  button.simulate("submit", { preventDefault: jest.fn() });
  // wait for async functions
  await new Promise(s => s());
  expect(install).toHaveBeenCalled();
});

it("should not call the install method when the validation fails unless forced", async () => {
  const validate = jest.fn(() => false);
  const install = jest.fn(() => true);
  const wrapper = shallow(<AppRepoForm {...defaultProps} validate={validate} install={install} />);
  let button = wrapper.find("form");

  button.simulate("submit", { preventDefault: jest.fn() });
  // wait for async functions
  await new Promise(s => s());
  expect(install).not.toHaveBeenCalled();
  wrapper.update();
  button = wrapper.find("button");
  expect(button.text()).toContain("Install Repo (force)");

  wrapper.find("form").simulate("submit", { preventDefault: jest.fn() });
  // wait for async functions
  await new Promise(s => s());
  expect(install).toHaveBeenCalled();
});
