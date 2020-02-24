import { shallow } from "enzyme";
import * as React from "react";
import { UnexpectedErrorAlert } from "../../ErrorAlert";
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

it("should render a validation error", () => {
  const wrapper = shallow(<AppRepoForm {...defaultProps} validateError={new Error("Boom!")} />);
  expect(wrapper.find(UnexpectedErrorAlert)).toExist();
  expect(wrapper.find(UnexpectedErrorAlert).html()).toMatch(/Validation Failed. Got:.*Boom!/);
});

it("should render validation confirmation", done => {
  const validate = jest.fn(() => true);
  const wrapper = shallow(<AppRepoForm {...defaultProps} validate={validate} />);
  wrapper.setState({
    url: "http://charts.com",
    authMethod: "custom",
    authHeader: "Bearer foo",
    customCA: "valid cert",
  });

  const button = wrapper.find("button").filterWhere(b => b.text() === "Validate");
  expect(button).toExist();
  expect(button.prop("disabled")).toBe(false);
  button.simulate("click");

  expect(validate).toHaveBeenCalledWith("http://charts.com", "Bearer foo", "valid cert");
  setTimeout(() => {
    expect(wrapper.text()).toMatch("Repository successfully validated");
    done();
  }, 1);
});

it("disables the validation button if there is no url", () => {
  const wrapper = shallow(<AppRepoForm {...defaultProps} />);
  const button = wrapper.find("button").filterWhere(b => b.text() === "Validate");
  expect(button).toExist();
  expect(button.prop("disabled")).toBe(true);
});

it("disables the validation button if the syncJobTemplate is not empty", () => {
  const wrapper = shallow(<AppRepoForm {...defaultProps} />);
  wrapper.setState({ url: "http://charts.com", syncJobPodTemplate: "not-empty" });
  const button = wrapper.find("button").filterWhere(b => b.text() === "Validate");
  expect(button).toExist();
  expect(button.prop("disabled")).toBe(true);
});

it("disables the submit and validation button while fetching", () => {
  const wrapper = shallow(<AppRepoForm {...defaultProps} isFetching={true} />);
  const buttons = wrapper.find("button");
  buttons.forEach(button => {
    expect(button.prop("disabled")).toBe(true);
  });
});
