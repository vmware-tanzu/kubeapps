// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { createHashHistory } from "history";
import { applyMiddleware, createStore } from "redux";
import { composeWithDevTools } from "redux-devtools-extension";
import thunkMiddleware from "redux-thunk";
import createRootReducer from "../reducers";

// Use Hash based routing to support deploying Kubeapps in arbitrary URL subpaths
export const history = createHashHistory();

export default createStore(
  createRootReducer(),
  composeWithDevTools(applyMiddleware(thunkMiddleware)),
);

export type AppStore = ReturnType<typeof createStore>;
