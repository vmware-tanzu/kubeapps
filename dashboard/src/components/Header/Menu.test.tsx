// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { deepClone } from "@cds/core/internal/utils/identity";
import { CdsButton } from "@cds/react/button";
import actions from "actions";
import * as ReactRedux from "react-redux";
import { app } from "shared/url";
import { Link } from "react-router-dom";
import { IClustersState } from "reducers/cluster";
import { SupportedThemes } from "shared/Config";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import Menu from "./Menu";

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

let spyOnUseDispatch: jest.SpyInstance;
beforeEach(() => {
  const mockDispatch = jest.fn(res => res);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

const defaultActions = { ...actions.config };
afterEach(() => {
  spyOnUseDispatch.mockRestore();
  actions.config = defaultActions;
});

it("opens the dropdown full menu", () => {
  const state = deepClone(initialState) as IStoreState;
  state.config.featureFlags = { operators: true };
  const store = getStore(state);
  const wrapper = mountWrapper(store, <Menu {...defaultProps} />);
  expect(wrapper.find(".dropdown")).not.toHaveClassName("open");
  const menu = wrapper.find("button");
  menu.simulate("click");
  wrapper.update();
  expect(wrapper.find(".dropdown")).toHaveClassName("open");
  // It render links for PackageRepositories and operators
  const links = wrapper.find(Link);
  expect(links).toHaveLength(3);
  expect(links.get(0).props.to).toEqual(app.config.pkgrepositories("default", "default"));
  expect(links.get(1).props.to).toEqual(app.config.operators("default", "default"));
  expect(links.get(2).props.to).toEqual("/docs");
});

it("opens the dropdown menu without operators item", () => {
  const state = deepClone(initialState) as IStoreState;
  state.config.featureFlags = { operators: false };
  const store = getStore(state);
  const wrapper = mountWrapper(store, <Menu {...defaultProps} />);
  expect(wrapper.find(".dropdown")).not.toHaveClassName("open");
  const menu = wrapper.find("button");
  menu.simulate("click");
  wrapper.update();
  expect(wrapper.find(".dropdown")).toHaveClassName("open");
  // It render links for PackageRepositories and operators
  const links = wrapper.find(Link);
  expect(links).toHaveLength(2);
  expect(links.get(0).props.to).toEqual(app.config.pkgrepositories("default", "default"));
  expect(links.get(1).props.to).toEqual("/docs");
});

it("logs out", () => {
  const logout = jest.fn();
  const wrapper = mountWrapper(defaultStore, <Menu {...defaultProps} logout={logout} />);
  const logoutButton = wrapper.find(CdsButton);
  // Simulate doesn't work with CdsButtons
  (logoutButton.prop("onClick") as any)();
  expect(logout).toHaveBeenCalled();
});

describe("theme switcher toggle", () => {
  it("toggle not checked by default", () => {
    const wrapper = mountWrapper(defaultStore, <Menu {...defaultProps} />);
    const toggle = wrapper.find("cds-toggle input");
    expect(toggle.prop("checked")).toBe(false);
  });

  it("toggle checked if dark theme is configured", () => {
    const wrapper = mountWrapper(
      getStore({ config: { theme: SupportedThemes.dark } } as Partial<IStoreState>),
      <Menu {...defaultProps} />,
    );
    const toggle = wrapper.find("cds-toggle input");
    expect(toggle.prop("checked")).toBe(true);
  });

  it("calls setTheme with the new theme", () => {
    actions.config.setUserTheme = jest.fn();
    const wrapper = mountWrapper(defaultStore, <Menu {...defaultProps} />);
    const toggle = wrapper.find("cds-toggle input");
    toggle.simulate("change");
    expect(actions.config.setUserTheme).toHaveBeenCalled();
  });
});
