import { mount } from "enzyme";
import * as React from "react";
import { Provider } from "react-redux";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";

import { CdsButton } from "@cds/react/button";
import { SupportedThemes } from "components/HeadManager/HeadManager";
import { act } from "react-dom/test-utils";
import { BrowserRouter, Link } from "react-router-dom";
import { IClustersState } from "reducers/cluster";
import Menu from "./Menu";

const mockStore = configureMockStore([thunk]);
const defaultStore = mockStore({});

const defaultProps = {
  clusters: {
    currentCluster: "default",
    clusters: {
      default: {
        currentNamespace: "default",
        namespaces: ["default", "other"],
        canCreateNS: true,
      },
    },
  } as IClustersState,
  defaultNamespace: "kubeapps-user",
  appVersion: "v2.0.0",
  logout: jest.fn(),
};

it("opens the dropdown menu", () => {
  const wrapper = mount(
    <Provider store={defaultStore}>
      <BrowserRouter>
        <Menu {...defaultProps} />
      </BrowserRouter>
    </Provider>,
  );
  expect(wrapper.find(".dropdown")).not.toHaveClassName("open");
  const menu = wrapper.find("button");
  menu.simulate("click");
  wrapper.update();
  expect(wrapper.find(".dropdown")).toHaveClassName("open");
  // It render links for AppRepositories and operators
  expect(wrapper.find(Link)).toHaveLength(3);
});

it("logs out", () => {
  const logout = jest.fn();
  const wrapper = mount(
    <Provider store={defaultStore}>
      <BrowserRouter>
        <Menu {...defaultProps} logout={logout} />
      </BrowserRouter>
    </Provider>,
  );
  const logoutButton = wrapper.find(CdsButton);
  // Simulate doesn't work with CdsButtons
  (logoutButton.prop("onClick") as any)();
  expect(logout).toHaveBeenCalled();
});

describe("theme switcher toggle", () => {
  it("toggle not checked by default", () => {
    const wrapper = mount(
      <Provider store={defaultStore}>
        <BrowserRouter>
          <Menu {...defaultProps} />
        </BrowserRouter>
      </Provider>,
    );
    const toggle = wrapper.find("cds-toggle input");
    expect(toggle.prop("checked")).toBe(false);
  });

  it("toggle checked if dark theme is configured", () => {
    localStorage.setItem("theme", SupportedThemes.dark);
    const wrapper = mount(
      <Provider store={defaultStore}>
        <BrowserRouter>
          <Menu {...defaultProps} />
        </BrowserRouter>
      </Provider>,
    );
    const toggle = wrapper.find("cds-toggle input");
    expect(toggle.prop("checked")).toBe(true);
  });

  it("toggle reloads page after changing theme", () => {
    localStorage.setItem("theme", SupportedThemes.dark);
    const wrapper = mount(
      <Provider store={defaultStore}>
        <BrowserRouter>
          <Menu {...defaultProps} />
        </BrowserRouter>
      </Provider>,
    );
    // After the JSDOM upgrade, window.xxx are read-only properties
    // https://github.com/facebook/jest/issues/9471
    Object.defineProperty(window, "location", {
      configurable: true,
      writable: true,
      value: { reload: jest.fn() },
    });
    const toggle = wrapper.find("cds-toggle input");
    act(() => {
      (toggle.prop("onChange") as any)();
    });
    wrapper.update();
    expect(window.location.reload).toHaveBeenCalled();
  });
});
