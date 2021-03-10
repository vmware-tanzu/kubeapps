import { CdsButton } from "@cds/react/button";
import actions from "actions";
import * as ReactRedux from "react-redux";
import { Link } from "react-router-dom";
import { IClustersState } from "reducers/cluster";
import { SupportedThemes } from "shared/Config";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
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

it("opens the dropdown menu", () => {
  const wrapper = mountWrapper(defaultStore, <Menu {...defaultProps} />);
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
      getStore({ config: { theme: SupportedThemes.dark } }),
      <Menu {...defaultProps} />,
    );
    const toggle = wrapper.find("cds-toggle input");
    expect(toggle.prop("checked")).toBe(true);
  });

  it("calls setTheme with the new theme", () => {
    actions.config.setTheme = jest.fn();
    const wrapper = mountWrapper(defaultStore, <Menu {...defaultProps} />);
    const toggle = wrapper.find("cds-toggle input");
    toggle.simulate("change");
    expect(actions.config.setTheme).toHaveBeenCalled();
  });
});
