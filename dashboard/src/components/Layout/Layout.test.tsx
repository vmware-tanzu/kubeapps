// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import * as ReactRedux from "react-redux";
import { getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import Layout from "./Layout";

let spyOnUseDispatch: jest.SpyInstance;
const kubeActions = { ...actions.namespace };
beforeEach(() => {
  actions.kube = {
    ...actions.kube,
    getResourceKinds: jest.fn(),
  };
  const mockDispatch = jest.fn(res => res);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.namespace = { ...kubeActions };
  spyOnUseDispatch.mockRestore();
  jest.restoreAllMocks();
});

const defaultState = {
  ...initialState,
  clusters: {
    ...initialState.clusters,
    currentCluster: "default",
    clusters: {
      ...initialState.clusters.clusters,
      default: {
        ...initialState.clusters.clusters[initialState.clusters.currentCluster],
        currentNamespace: "default",
        namespaces: ["default", "other"],
      },
    },
  },
  auth: { ...initialState.auth, authenticated: true },
} as IStoreState;

it("fetches resource kinds when operators enabled", () => {
  const state = {
    ...defaultState,
    config: {
      featureFlags: {
        operators: true,
      },
    },
  };

  mountWrapper(getStore(state), <Layout />);

  expect(actions.kube.getResourceKinds).toHaveBeenCalled();
});

it("does not fetch resource kinds when operators disabled", () => {
  mountWrapper(getStore(defaultState), <Layout />);

  expect(actions.kube.getResourceKinds).not.toHaveBeenCalled();
});
