// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import "@testing-library/jest-dom";
import { act, screen } from "@testing-library/react";
import actions from "actions";
import LoadingWrapper from "components/LoadingWrapper";
import * as ReactRedux from "react-redux";
import { Route, Routes } from "react-router-dom";
import { IConfigState } from "reducers/config";
import {
  defaultStore,
  getStore,
  mountWrapper,
  renderWithProviders,
} from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import LoginForm from "./LoginForm";
import OAuthLogin from "./OauthLogin";

const defaultCluster = "default-cluster";

let spyOnUseDispatch: jest.SpyInstance;
beforeEach(() => {
  const mockDispatch = jest.fn(res => res);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});
afterEach(() => {
  spyOnUseDispatch.mockRestore();
  jest.restoreAllMocks();
});

const authenticationError = "it's a trap";

describe("while authenticating", () => {
  it("behaves like a loading component", () => {
    const state = {
      ...defaultStore,
      auth: {
        authenticating: true,
      },
    };
    renderWithProviders(
      <Routes>
        <Route path="/login" element={<LoginForm />} />
      </Routes>,
      {
        store: getStore(state),
        initialEntries: ["/login"],
      },
    );

    expect(screen.getByRole("img")).toHaveAttribute("aria-label", "Loading");
    expect(screen.queryByLabelText("Token")).not.toBeInTheDocument();
  });
});

describe("token login form", () => {
  it("renders a token login form", () => {
    renderWithProviders(
      <Routes>
        <Route path="/login" element={<LoginForm />} />
      </Routes>,
      {
        initialEntries: ["/login"],
      },
    );

    expect(screen.queryByRole("img")).not.toBeInTheDocument();
    expect(screen.getByLabelText("Token")).toBeInTheDocument();
  });

  it("renders a link to the access control documentation", () => {
    const state = {
      ...defaultStore,
      config: {
        appVersion: "devel",
      },
    };
    renderWithProviders(
      <Routes>
        <Route path="/login" element={<LoginForm />} />
      </Routes>,
      {
        initialEntries: ["/login"],
        store: getStore(state),
      },
    );

    expect(screen.queryByRole("link", { name: "More Info" })).toHaveAttribute(
      "href",
      "https://github.com/vmware-tanzu/kubeapps/blob/devel/site/content/docs/latest/howto/access-control.md",
    );
  });

  it("updates the token in the state when the input is changed", () => {
    const wrapper = mountWrapper(defaultStore, <LoginForm />);
    let input = wrapper.find("input#token");
    act(() => {
      input.simulate("change", {
        target: { value: "f00b4r" },
        current: { value: "ff00b4r" },
      });
    });
    wrapper.update();
    input = wrapper.find("input#token");
    expect(input.prop("value")).toBe("f00b4r");
  });

  describe("redirect if authenticated", () => {
    it("redirects to / if no current location", () => {
      const state = {
        ...defaultStore,
        auth: {
          authenticated: true,
        },
      };
      renderWithProviders(
        <Routes>
          <Route path="/login" element={<LoginForm />} />
          <Route path="/" element={<h1>home</h1>} />
        </Routes>,
        {
          store: getStore(state),
          initialEntries: ["/login"],
        },
      );

      expect(screen.getByRole("heading", { name: "home" })).toBeInTheDocument();
    });
  });

  it("calls the authenticate handler when the form is submitted", () => {
    const authenticate = jest.fn().mockReturnValue({
      then: jest.fn(f => f()),
      catch: jest.fn(f => f()),
    });
    actions.auth.authenticate = authenticate;
    const wrapper = mountWrapper(defaultStore, <LoginForm />);
    act(() => {
      wrapper.find("input#token").simulate("change", { target: { value: "f00b4r" } });
    });
    act(() => {
      wrapper.find("form").simulate("submit", { preventDefault: jest.fn() });
    });
    expect(authenticate).toBeCalledWith(defaultCluster, "f00b4r", false);
  });

  it("calls the authenticate handler if a token is passed as query param", () => {
    const authenticate = jest.fn().mockReturnValue({
      then: jest.fn(f => f()),
      catch: jest.fn(f => f()),
    });
    actions.auth.authenticate = authenticate;
    renderWithProviders(
      <Routes>
        <Route path="/login" element={<LoginForm />} />
      </Routes>,
      {
        store: defaultStore,
        initialEntries: ["/login?token=f00b4r"],
      },
    );

    expect(authenticate).toBeCalledWith(defaultCluster, "f00b4r", false);
  });

  it("calls the authenticate handler just once if a failed token is passed as query param", () => {
    const authenticate = jest.fn().mockReturnValue({
      then: jest.fn(f => f()),
      catch: jest.fn(f => f()),
    });
    actions.auth.authenticate = authenticate;
    renderWithProviders(
      <Routes>
        <Route path="/login" element={<LoginForm />} />
      </Routes>,
      {
        store: defaultStore,
        initialEntries: ["/login?token=bad-token"],
      },
    );

    expect(authenticate).toBeCalledWith(defaultCluster, "bad-token", false);
    expect(authenticate).toBeCalledTimes(1);
  });

  it("does not call the authenticate handler in oauth login if token is passed as query param", () => {
    const authenticate = jest.fn();
    renderWithProviders(
      <Routes>
        <Route path="/login" element={<LoginForm />} />
      </Routes>,
      {
        store: defaultStore,
        initialEntries: ["/login?token=f00b4r"],
      },
    );

    expect(authenticate).not.toBeCalled();
  });

  it("displays an error if the authentication error is passed", () => {
    const state = {
      ...defaultStore,
      auth: {
        authenticationError,
      },
    };
    const wrapper = mountWrapper(getStore(state), <LoginForm />);

    expect(wrapper.find(".error").exists()).toBe(true);
  });

  it("does not display the oauth login if oauthLoginURI provided", () => {
    const wrapper = mountWrapper(defaultStore, <LoginForm />);

    expect(wrapper.find("a.button").exists()).toBe(false);
  });
});

describe("oauth login form", () => {
  it("does not display the token login if oauthLoginURI provided", () => {
    const state = {
      ...defaultStore,
      config: {
        oauthLoginURI: "/sign/in",
        authProxyEnabled: true,
      } as IConfigState,
    };
    const checkCookieAuthentication = jest.fn().mockReturnValue({
      then: jest.fn(() => false),
      catch: jest.fn(() => false),
    });
    actions.auth.checkCookieAuthentication = checkCookieAuthentication;
    const wrapper = mountWrapper(getStore(state), <LoginForm />);

    expect(wrapper.find("input#token").exists()).toBe(false);
  });

  it("displays the oauth login if authProxyEnabled", () => {
    const checkCookieAuthentication = jest.fn().mockReturnValue({
      then: jest.fn(f => f()),
      catch: jest.fn(f => f()),
    });
    actions.auth.checkCookieAuthentication = checkCookieAuthentication;

    const state = {
      ...defaultStore,
      config: {
        authProxyEnabled: true,
        oauthLoginURI: "/sign/in",
      } as IConfigState,
    };
    const wrapper = mountWrapper(getStore({ ...state } as Partial<IStoreState>), <LoginForm />);
    expect(checkCookieAuthentication).toHaveBeenCalled();
    expect(wrapper.find(OAuthLogin)).toExist();
    expect(wrapper.find("a").findWhere(a => a.prop("href") === "/sign/in")).toExist();
  });

  it("doesn't render the login form if the cookie has not been checked yet", () => {
    const state = {
      ...defaultStore,
      config: {
        authProxyEnabled: true,
        oauthLoginURI: "/sign/in",
      } as IConfigState,
    };
    const checkCookieAuthentication = jest.fn().mockReturnValue({
      then: jest.fn(() => false),
      catch: jest.fn(() => false),
    });
    actions.auth.checkCookieAuthentication = checkCookieAuthentication;

    const wrapper = mountWrapper(getStore({ ...state } as Partial<IStoreState>), <LoginForm />);
    expect(wrapper.find(LoadingWrapper)).toExist();
    expect(wrapper.find(OAuthLogin)).not.toExist();
  });

  it("changes window location when skipping oauth login page", () => {
    // After the JSDOM upgrade, window.xxx are read-only properties
    // https://github.com/facebook/jest/issues/9471
    Object.defineProperty(window, "location", {
      configurable: true,
      writable: true,
      value: { replace: jest.fn() },
    });
    const state = {
      ...defaultStore,
      config: {
        authProxySkipLoginPage: true,
        oauthLoginURI: "/sign/in",
        authProxyEnabled: true,
      },
    };
    const checkCookieAuthentication = jest.fn().mockReturnValue({
      then: jest.fn(() => true),
      catch: jest.fn(() => true),
    });
    actions.auth.checkCookieAuthentication = checkCookieAuthentication;

    act(() => {
      mountWrapper(getStore(state), <LoginForm />);
    });
    expect(window.location.replace).toHaveBeenCalledWith("/sign/in");
  });
});
