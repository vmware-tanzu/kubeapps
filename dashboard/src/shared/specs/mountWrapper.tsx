import { mount } from "enzyme";
import * as React from "react";
import { Provider } from "react-redux";
import { BrowserRouter as Router } from "react-router-dom";
import configureMockStore, { MockStore } from "redux-mock-store";
import thunk from "redux-thunk";
import { merge } from "lodash";

import { IStoreState } from "../../shared/types";

const mockStore = configureMockStore([thunk]);

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

export const defaultStore = mockStore(initialState);

// getStore returns a store initialised with a merge of
// the initial state with any passed extra state.
export const getStore = (extraState: Object) => {
  const state = { ...initialState };
  return mockStore(merge(state, extraState));
};

export const mountWrapper = (store: MockStore, children: React.ReactElement) =>
  mount(
    <Provider store={store}>
      <Router>{children}</Router>
    </Provider>,
  );
