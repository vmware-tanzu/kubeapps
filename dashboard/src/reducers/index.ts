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

const rootReducer = combineReducers<IStoreState>({
  apps: appsReducer,
  auth: authReducer,
  catalog: catalogReducer,
  charts: chartsReducer,
  config: configReducer,
  kube: kubeReducer,
  namespace: clusterReducer,
  repos: reposReducer,
  operators: operatorReducer,
});

export default rootReducer;
