import { mount } from "enzyme";
import * as React from "react";
import { Provider } from "react-redux";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";

import { CdsButton } from "components/Clarity/clarity";
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
  expect(wrapper.find(Link)).toHaveLength(2);
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
