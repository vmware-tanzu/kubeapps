// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { RouterState } from "connected-react-router";
import { mount } from "enzyme";
import { cloneDeep, merge } from "lodash";
import { IntlProvider } from "react-intl";
import { DefaultRootState, Provider } from "react-redux";
import { MemoryRouter, BrowserRouter as Router } from "react-router-dom";
import { initialState as installedPackagesInitialState } from "reducers/installedpackages";
import { initialState as authInitialState } from "reducers/auth";
import { initialState as availablePackagesInitialState } from "reducers/availablepackages";
import { initialState as clustersInitialState } from "reducers/cluster";
import { initialState as configInitialState } from "reducers/config";
import { initialState as kubeInitialState } from "reducers/kube";
import { operatorsInitialState } from "reducers/operators";
import { initialState as reposInitialState } from "reducers/repos";
import configureMockStore, { MockStore } from "redux-mock-store";
import thunk from "redux-thunk";
import I18n from "shared/I18n";
import { IStoreState } from "shared/types";
import React, { PropsWithChildren } from "react";
import { render } from "@testing-library/react";
import type { RenderOptions } from "@testing-library/react";
import { configureStore } from "@reduxjs/toolkit";
import type { PreloadedState } from "@reduxjs/toolkit";
import { reducers } from "reducers";
import { AppStore } from "store";

const mockStore = configureMockStore([thunk]);

export const initialState = {
  apps: cloneDeep(installedPackagesInitialState),
  auth: cloneDeep(authInitialState),
  packages: cloneDeep(availablePackagesInitialState),
  config: {
    ...cloneDeep(configInitialState),
    kubeappsCluster: "default-cluster",
    kubeappsNamespace: "kubeapps",
    helmGlobalNamespace: "kubeapps-repos-global",
    carvelGlobalNamespace: "kapp-controller-packaging-global",
  } as IStoreState["config"],
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
  } as IStoreState["clusters"],
  repos: cloneDeep(reposInitialState),
  operators: cloneDeep(operatorsInitialState),
  router: {} as RouterState,
} as IStoreState;

export const defaultStore = mockStore(initialState);

// Default to english 1i8nconfiguration
const messages = I18n.getDefaultConfig().messages;
const locale = I18n.getDefaultConfig().locale;

// getStore returns a store initialised with a merge of
// the initial state with any passed extra state.
export const getStore = (extraState: object) => {
  const state = cloneDeep(initialState);
  return mockStore(merge(state, extraState));
};

export const mountWrapper = (store: MockStore, children: React.ReactElement) =>
  mount(
    <Provider store={store}>
      <IntlProvider locale={locale} key={locale} messages={messages} defaultLocale={locale}>
        <Router>{children}</Router>
      </IntlProvider>
      ,
    </Provider>,
  );

// Things have moved on for testing to utilise the React Testing Library (RTL)
// so that the redux documentation now recommends the following setup, which
// we should use for new code and gradually move old code over.
// https://redux.js.org/usage/writing-tests

// This type interface extends the default options for render from RTL, as well
// as allows the user to specify other things such as initialState, store.
interface ExtendedRenderOptions extends Omit<RenderOptions, "queries"> {
  preloadedState?: PreloadedState<DefaultRootState>;
  store?: AppStore;
}

export function renderWithProviders(
  ui: React.ReactElement,
  {
    preloadedState = {},
    // Automatically create a store instance if no store was passed in
    store = configureStore({ reducer: reducers, preloadedState }),
    ...renderOptions
  }: ExtendedRenderOptions = {},
) {
  function Wrapper({ children }: PropsWithChildren<{}>): JSX.Element {
    return (
      <MemoryRouter>
        <Provider store={store}>{children}</Provider>
      </MemoryRouter>
    );
  }

  // Return an object with the store and all of RTL's query functions
  return { store, ...render(ui, { wrapper: Wrapper, ...renderOptions }) };
}
