import { mount } from "enzyme";
import { merge } from "lodash";
import { cloneDeep } from "lodash";
import * as React from "react";
import { Provider } from "react-redux";
import { BrowserRouter as Router } from "react-router-dom";
import configureMockStore, { MockStore } from "redux-mock-store";
import thunk from "redux-thunk";

import { operatorsInitialState } from "reducers/operators";
import { IAppRepository, ISecret, IStoreState } from "../../shared/types";

const mockStore = configureMockStore([thunk]);

export const initialState = {
  apps: {},
  auth: {},
  catalog: {},
  charts: {},
  config: { featureFlags: {}, kubeappsCluster: "default-cluster" },
  kube: {
    items: {},
  },
  clusters: {
    currentCluster: "default-cluster",
  },
  repos: {
    errors: {},
    repos: [] as IAppRepository[],
    repoSecrets: [] as ISecret[],
    imagePullSecrets: [] as ISecret[],
  },
  operators: cloneDeep(operatorsInitialState),
} as IStoreState;

export const defaultStore = mockStore(initialState);

// getStore returns a store initialised with a merge of
// the initial state with any passed extra state.
export const getStore = (extraState: object) => {
  const state = cloneDeep(initialState);
  return mockStore(merge(state, extraState));
};

export const mountWrapper = (store: MockStore, children: React.ReactElement) =>
  mount(
    <Provider store={store}>
      <Router>{children}</Router>
    </Provider>,
  );
