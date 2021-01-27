import { RouterState } from "connected-react-router";
import { mount } from "enzyme";
import { merge } from "lodash";
import { cloneDeep } from "lodash";
import * as React from "react";
import { Provider } from "react-redux";
import { BrowserRouter as Router } from "react-router-dom";
import { initialState as appsInitialState } from "reducers/apps";
import { initialState as authInitialState } from "reducers/auth";
import { initialState as chartsInitialState } from "reducers/charts";
import { initialState as clustersInitialState } from "reducers/cluster";
import { initialState as configInitialState } from "reducers/config";
import { initialState as kubeInitialState } from "reducers/kube";
import { operatorsInitialState } from "reducers/operators";
import { initialState as reposInitialState } from "reducers/repos";
import configureMockStore, { MockStore } from "redux-mock-store";
import thunk from "redux-thunk";
import { IStoreState } from "../../shared/types";

const mockStore = configureMockStore([thunk]);

export const initialState = {
  apps: cloneDeep(appsInitialState),
  auth: cloneDeep(authInitialState),
  charts: cloneDeep(chartsInitialState),
  config: {
    ...cloneDeep(configInitialState),
    kubeappsCluster: "default-cluster",
    kubeappsNamespace: "kubeapps",
  },
  kube: cloneDeep(kubeInitialState),
  clusters: {
    ...cloneDeep(clustersInitialState),
    currentCluster: "default-cluster",
    clusters: {
      "default-cluster": {
        currentNamespace: "default",
        namespaces: ["default", "other"],
        canCreateNS: true,
      },
      "second-cluster": {
        currentNamespace: "default",
        namespaces: ["default", "other"],
        canCreateNS: true,
      },
    },
  },
  repos: cloneDeep(reposInitialState),
  operators: cloneDeep(operatorsInitialState),
  router: {} as RouterState,
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
