// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { connectRouter } from "connected-react-router";
import { History } from "history";
import { combineReducers } from "redux";
import { IStoreState } from "shared/types";
import installedPackagesReducer from "./installedpackages";
import authReducer from "./auth";
import packageReducer from "./availablepackages";
import clusterReducer from "./cluster";
import configReducer from "./config";
import kubeReducer from "./kube";
import operatorReducer from "./operators";
import reposReducer from "./repos";

const rootReducer = (history: History) =>
  combineReducers<IStoreState>({
    router: connectRouter(history),
    apps: installedPackagesReducer,
    auth: authReducer,
    packages: packageReducer,
    config: configReducer,
    kube: kubeReducer,
    clusters: clusterReducer,
    repos: reposReducer,
    operators: operatorReducer,
  });

export default rootReducer;
