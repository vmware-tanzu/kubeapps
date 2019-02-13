import { shallow } from "enzyme";
import { Location } from "history";
import context from "jest-plugin-context";
import * as React from "react";
import { Redirect } from "react-router-dom";
import itBehavesLike from "../../shared/specs";

import LoginForm from "./LoginForm";

const emptyLocation: Location = {
  hash: "",
  pathname: "",
  search: "",
  state: "",
};

const defaultProps = {
  authenticate: jest.fn(),
  authenticated: false,
  authenticating: false,
  authenticationError: undefined,
  location: emptyLocation,
  tryToAuthenticateWithOIDC: jest.fn(),
};

const authenticationError = "it's a trap";

describe("componentDidMount", () => {
  it("should call tryToAuthenticateWithOIDC", () => {
    const tryToAuthenticateWithOIDC = jest.fn();
    shallow(<LoginForm {...defaultProps} tryToAuthenticateWithOIDC={tryToAuthenticateWithOIDC} />);
    expect(tryToAuthenticateWithOIDC).toHaveBeenCalled();
  });
});

context("while authenticating", () => {
  itBehavesLike("aLoadingComponent", {
    component: LoginForm,
    props: { ...defaultProps, authenticating: true },
  });
});

it("renders a token login form", () => {
  const wrapper = shallow(<LoginForm {...defaultProps} />);
  expect(wrapper.find("input#token").exists()).toBe(true);
  expect(wrapper.find(Redirect).exists()).toBe(false);
  expect(wrapper).toMatchSnapshot();
});

it("renders a link to the access control documentation", () => {
  const wrapper = shallow(<LoginForm {...defaultProps} />);
  expect(wrapper.find("a").props()).toMatchObject({
    href: "https://github.com/kubeapps/kubeapps/blob/master/docs/user/access-control.md",
    target: "_blank",
  });
});

it("updates the token in the state when the input is changed", () => {
  const wrapper = shallow(<LoginForm {...defaultProps} />);
  wrapper.find("input#token").simulate("change", { currentTarget: { value: "f00b4r" } });
  expect(wrapper.state("token")).toBe("f00b4r");
});

describe("redirect if authenticated", () => {
  it("redirects to / if no current location", () => {
    const wrapper = shallow(<LoginForm {...defaultProps} authenticated={true} />);
    const redirect = wrapper.find(Redirect);
    expect(redirect.exists()).toBe(true);
    expect(redirect.props()).toEqual({ push: false, to: { pathname: "/" } });
  });

  it("redirects to previous location", () => {
    const location = Object.assign({}, emptyLocation);
    location.state = { from: "/test" };
    const wrapper = shallow(
      <LoginForm {...defaultProps} authenticated={true} location={location} />,
    );
    const redirect = wrapper.find(Redirect);
    expect(redirect.exists()).toBe(true);
    expect(redirect.props()).toEqual({ push: false, to: "/test" });
  });
});

it("calls the authenticate handler when the form is submitted", () => {
  const wrapper = shallow(<LoginForm {...defaultProps} />);
  wrapper.find("input#token").simulate("change", { currentTarget: { value: "f00b4r" } });
  wrapper.find("form").simulate("submit", { preventDefault: jest.fn() });
  expect(defaultProps.authenticate).toBeCalledWith("f00b4r");
});

it("displays an error if the authentication error is passed", () => {
  const wrapper = shallow(
    <LoginForm {...defaultProps} authenticationError={authenticationError} />,
  );

  expect(wrapper.find(".alert-error").exists()).toBe(true);
  expect(wrapper).toMatchSnapshot();
});
