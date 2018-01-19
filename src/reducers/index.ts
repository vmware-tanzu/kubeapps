import { routerReducer } from "react-router-redux";
import { combineReducers } from "redux";

import { IStoreState } from "../shared/types";
import chartsReducer from "./charts";

const rootReducer = combineReducers<IStoreState>({
  charts: chartsReducer,
  router: routerReducer,
});

export default rootReducer;
