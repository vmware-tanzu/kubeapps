import { shallow } from "enzyme";
import { Location } from "history";
import context from "jest-plugin-context";
import * as React from "react";
import { Redirect } from "react-router";
import itBehavesLike from "../../shared/specs";

import LoginForm from "./LoginForm";

const emptyLocation: Location = {
  hash: "",
  pathname: "",
  search: "",
  state: "",
};

const defaultCluster = "default";

const defaultProps = {
  cluster: defaultCluster,
  authenticate: jest.fn(),
  authenticated: false,
  authenticating: false,
  authenticationError: undefined,
  location: emptyLocation,
  checkCookieAuthentication: jest.fn(),
  oauthLoginURI: "",
};

const authenticationError = "it's a trap";

describe("componentDidMount", () => {
  it("calls checkCookieAuthentication when oauthLoginURI provided", () => {
    const props = {
      ...defaultProps,
      oauthLoginURI: "/sign/in",
    };
    const checkCookieAuthentication = jest.fn();
    shallow(<LoginForm {...props} checkCookieAuthentication={checkCookieAuthentication} />);
    expect(checkCookieAuthentication).toHaveBeenCalled();
  });

  it("does not call checkCookieAuthentication when oauthLoginURI not provided", () => {
    const checkCookieAuthentication = jest.fn();
    shallow(<LoginForm {...defaultProps} checkCookieAuthentication={checkCookieAuthentication} />);
    expect(checkCookieAuthentication).not.toHaveBeenCalled();
  });
});

context("while authenticating", () => {
  itBehavesLike("aLoadingComponent", {
    component: LoginForm,
    props: { ...defaultProps, authenticating: true },
  });
});

describe("token login form", () => {
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
      expect(redirect.props()).toEqual({ to: { pathname: "/" } });
    });

    it("redirects to previous location", () => {
      const location = Object.assign({}, emptyLocation);
      location.state = { from: "/test" };
      const wrapper = shallow(
        <LoginForm {...defaultProps} authenticated={true} location={location} />,
      );
      const redirect = wrapper.find(Redirect);
      expect(redirect.props()).toEqual({ to: "/test" });
    });
  });

  it("calls the authenticate handler when the form is submitted", () => {
    const wrapper = shallow(<LoginForm {...defaultProps} />);
    wrapper.find("input#token").simulate("change", { currentTarget: { value: "f00b4r" } });
    wrapper.find("form").simulate("submit", { preventDefault: jest.fn() });
    expect(defaultProps.authenticate).toBeCalledWith(defaultCluster, "f00b4r");
  });

  it("displays an error if the authentication error is passed", () => {
    const wrapper = shallow(
      <LoginForm {...defaultProps} authenticationError={authenticationError} />,
    );

    expect(wrapper.find(".alert-error").exists()).toBe(true);
    expect(wrapper).toMatchSnapshot();
  });

  it("does not display the oauth login if oauthLoginURI provided", () => {
    const props = {
      ...defaultProps,
      oauthLoginURI: "",
    };

    const wrapper = shallow(<LoginForm {...props} />);

    expect(wrapper.find("a.button").exists()).toBe(false);
  });
});

describe("oauth login form", () => {
  const props = {
    ...defaultProps,
    oauthLoginURI: "/sign/in",
  };
  it("does not display the token login if oauthLoginURI provided", () => {
    const wrapper = shallow(<LoginForm {...props} />);

    expect(wrapper.find("input#token").exists()).toBe(false);
  });

  it("displays the oauth login if oauthLoginURI provided", () => {
    const wrapper = shallow(<LoginForm {...props} />);

    expect(wrapper.find("a.button").exists()).toBe(true);
  });

  it("renders a login button link", () => {
    const wrapper = shallow(<LoginForm {...props} />);
    expect(wrapper).toMatchSnapshot();
  });
});
