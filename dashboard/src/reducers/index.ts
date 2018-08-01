import { routerReducer } from "react-router-redux";
import { combineReducers } from "redux";

import { IStoreState } from "../shared/types";
import appsReducer from "./apps";
import authReducer from "./auth";
import catalogReducer from "./catalog";
import chartsReducer from "./charts";
import configReducer from "./config";
import functionsReducer from "./functions";
import namespaceReducer from "./namespace";
import reposReducer from "./repos";

const rootReducer = combineReducers<IStoreState>({
  apps: appsReducer,
  auth: authReducer,
  catalog: catalogReducer,
  charts: chartsReducer,
  config: configReducer,
  functions: functionsReducer,
  namespace: namespaceReducer,
  repos: reposReducer,
  router: routerReducer,
});

export default rootReducer;
