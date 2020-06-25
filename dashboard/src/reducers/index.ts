import { connectRouter } from "connected-react-router"
import { History, LocationState } from "history"
import { combineReducers } from "redux";

import { IStoreState } from "../shared/types";
import appsReducer from "./apps";
import authReducer from "./auth";
import catalogReducer from "./catalog";
import chartsReducer from "./charts";
import clusterReducer from "./cluster";
import configReducer from "./config";
import kubeReducer from "./kube";
import operatorReducer from "./operators";
import reposReducer from "./repos";

export default (history: History<LocationState>) => combineReducers<IStoreState>({
  router: connectRouter(history),
  apps: appsReducer,
  auth: authReducer,
  catalog: catalogReducer,
  charts: chartsReducer,
  config: configReducer,
  kube: kubeReducer,
  clusters: clusterReducer,
  repos: reposReducer,
  operators: operatorReducer,
});
