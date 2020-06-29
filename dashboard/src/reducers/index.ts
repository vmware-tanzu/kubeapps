import { History } from "history";
import { combineReducers } from "redux";

import { connectRouter } from "connected-react-router";
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

const rootReducer = (history: History) =>
  combineReducers<IStoreState>({
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

export default rootReducer;
