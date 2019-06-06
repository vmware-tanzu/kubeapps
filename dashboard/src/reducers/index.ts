import { combineReducers } from "redux";

import { IStoreState } from "../shared/types";
import appsReducer from "./apps";
import authReducer from "./auth";
import catalogReducer from "./catalog";
import chartsReducer from "./charts";
import configReducer from "./config";
import kubeReducer from "./kube";
import namespaceReducer from "./namespace";
import reposReducer from "./repos";

const rootReducer = combineReducers<IStoreState>({
  apps: appsReducer,
  auth: authReducer,
  catalog: catalogReducer,
  charts: chartsReducer,
  config: configReducer,
  kube: kubeReducer,
  namespace: namespaceReducer,
  repos: reposReducer,
});

export default rootReducer;
