import { routerReducer } from "react-router-redux";
import { combineReducers } from "redux";

import { IStoreState } from "../shared/types";
import appsReducer from "./apps";
import catalogReducer from "./catalog";
import chartsReducer from "./charts";
import functionsReducer from "./functions";
import reposReducer from "./repos";
import runtimesReducer from "./runtimes";

const rootReducer = combineReducers<IStoreState>({
  apps: appsReducer,
  catalog: catalogReducer,
  charts: chartsReducer,
  functions: functionsReducer,
  repos: reposReducer,
  router: routerReducer,
  runtimes: runtimesReducer,
});

export default rootReducer;
