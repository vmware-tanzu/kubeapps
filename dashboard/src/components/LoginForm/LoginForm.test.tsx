import { shallow } from "enzyme";
import { Location } from "history";
import * as React from "react";
import { Redirect } from "react-router-dom";

import LoginForm from "./LoginForm";

const emptyLocation: Location = {
  hash: "",
  pathname: "",
  search: "",
  state: "",
};

it("renders a token login form", () => {
  const wrapper = shallow(
    <LoginForm authenticated={false} authenticate={jest.fn()} location={emptyLocation} />,
  );
  expect(wrapper.find("input#token")).toHaveLength(1);
  expect(wrapper.find(Redirect)).toHaveLength(0);
  expect(wrapper).toMatchSnapshot();
});

it("renders a link to the access control documentation", () => {
  const wrapper = shallow(
    <LoginForm authenticated={false} authenticate={jest.fn()} location={emptyLocation} />,
  );
  expect(wrapper.find("a").props()).toMatchObject({
    href: "https://github.com/kubeapps/kubeapps/blob/master/docs/access-control.md",
    target: "_blank",
  });
});

it("updates the token in the state when the input is changed", () => {
  const wrapper = shallow(
    <LoginForm authenticated={false} authenticate={jest.fn()} location={emptyLocation} />,
  );
  wrapper.find("input#token").simulate("change", { currentTarget: { value: "f00b4r" } });
  expect(wrapper.state("token")).toBe("f00b4r");
});

describe("redirect if authenticated", () => {
  it("redirects to / if no current location", () => {
    const wrapper = shallow(
      <LoginForm authenticated={true} authenticate={jest.fn()} location={emptyLocation} />,
    );
    const redirect = wrapper.find(Redirect);
    expect(redirect).toHaveLength(1);
    expect(redirect.props()).toEqual({ push: false, to: { pathname: "/" } });
  });

  it("redirects to previous location", () => {
    const location = emptyLocation;
    location.state = { from: "/test" };
    const wrapper = shallow(
      <LoginForm authenticated={true} authenticate={jest.fn()} location={location} />,
    );
    const redirect = wrapper.find(Redirect);
    expect(redirect).toHaveLength(1);
    expect(redirect.props()).toEqual({ push: false, to: "/test" });
  });
});

it("calls the authenticate handler when the form is submitted", () => {
  const authenticate = jest.fn();
  const wrapper = shallow(
    <LoginForm authenticated={false} authenticate={authenticate} location={emptyLocation} />,
  );
  wrapper.find("input#token").simulate("change", { currentTarget: { value: "f00b4r" } });
  wrapper.find("form").simulate("submit", { preventDefault: jest.fn() });
  expect(authenticate).toBeCalledWith("f00b4r");
  expect(wrapper.state("authenticating")).toBe(true);
});

it("displays an error if the authenticate handler throws an error", async () => {
  const authenticate = async () => {
    throw new Error("it's a trap");
  };
  const wrapper = shallow(
    <LoginForm authenticated={false} authenticate={authenticate} location={emptyLocation} />,
  );
  wrapper.find("input#token").simulate("change", { currentTarget: { value: "f00b4r" } });
  wrapper.find("form").simulate("submit", { preventDefault: jest.fn() });

  // wait for promise to resolve
  try {
    await authenticate();
    fail(new Error("expected authenticate handler to throw error"));
  } catch (e) {
    expect(wrapper.state()).toMatchObject({
      authenticating: false,
      error: e.toString(),
    });

    wrapper.update();
    expect(wrapper.find(".alert-error").exists()).toBe(true);
    expect(wrapper).toMatchSnapshot();
  }
});

it("allows you to dismiss the error alert", () => {
  const wrapper = shallow(
    <LoginForm authenticated={false} authenticate={jest.fn()} location={emptyLocation} />,
  );
  wrapper.setState({ error: "it's a trap" });
  expect(wrapper.find(".alert-error").exists()).toBe(true);

  wrapper.find("button.alert__close").simulate("click");
  expect(wrapper.find(".alert-error").exists()).toBe(false);
  expect(wrapper.state("error")).toBeUndefined();
});
