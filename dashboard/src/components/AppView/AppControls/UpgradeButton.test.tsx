import { mount } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";
import { ArrowUpCircle } from "react-feather";
import { Provider } from "react-redux";
import { Redirect } from "react-router";
import configureMockStore, { MockStore } from "redux-mock-store";
import thunk from "redux-thunk";
import { IStoreState } from "shared/types";
import * as url from "shared/url";
import UpgradeButton, { IUpgradeButtonProps } from "./UpgradeButton";

const mockStore = configureMockStore([thunk]);

// TODO(absoludity): As we move to function components with (redux) hooks we'll need to
// be including state in tests, so we may want to put things like initialState
// and a generalized getWrapper in a test helpers or similar package?
const initialState = {
  apps: {},
  auth: {},
  catalog: {},
  charts: {},
  config: {},
  kube: {},
  clusters: {
    currentCluster: "default-cluster",
  },
  repos: {},
  operators: {},
} as IStoreState;

const getWrapper = (store: MockStore, props: IUpgradeButtonProps) =>
  mount(
    <Provider store={store}>
      <UpgradeButton {...props} />
    </Provider>,
  );

const defaultProps = {
  releaseName: "foo",
  releaseNamespace: "default",
  push: jest.fn(),
} as IUpgradeButtonProps;

it("renders a redirect when clicking upgrade", () => {
  const push = jest.fn();
  const store = mockStore(initialState);
  const wrapper = getWrapper(store, { ...defaultProps, push });
  const button = wrapper.find(".button").filterWhere(i => i.text() === "Upgrade");
  expect(button.exists()).toBe(true);
  expect(wrapper.find(Redirect).exists()).toBe(false);

  button.simulate("click");
  expect(push.mock.calls.length).toBe(1);
  expect(push.mock.calls[0]).toEqual([url.app.apps.upgrade("default-cluster", "default", "foo")]);
});

context("when a new version is available", () => {
  it("should show a modify the style", () => {
    const store = mockStore(initialState);
    const wrapper = getWrapper(store, { ...defaultProps, newVersion: true });
    const icon = wrapper.find(ArrowUpCircle);
    expect(icon).toExist();
  });
});
