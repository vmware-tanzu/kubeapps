// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import * as ReactRedux from "react-redux";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import Layout from "./Layout";

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.namespace };
beforeEach(() => {
  actions.kube = {
    ...actions.kube,
    getResourceKinds: jest.fn(),
  };
  const mockDispatch = jest.fn(res => res);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.namespace = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
  jest.restoreAllMocks();
});

const defaultState = {
  clusters: {
    currentCluster: "default",
    clusters: {
      default: {
        currentNamespace: "default",
        namespaces: ["default", "other"],
      },
    },
  },
  auth: { authenticated: true },
};

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

it("does not fetch resource kinds when operators disaabled", () => {
  mountWrapper(getStore(defaultState), <Layout />);

  expect(actions.kube.getResourceKinds).not.toHaveBeenCalled();
});
