import { History } from "history";
import { routerMiddleware } from "react-router-redux";
import { applyMiddleware, createStore } from "redux";
import thunkMiddleware from "redux-thunk";

import rootReducer from "../reducers";
import { IStoreState } from "../shared/types";

const configureStore = (history: History) =>
  createStore<IStoreState>(
    rootReducer,
    applyMiddleware(thunkMiddleware, routerMiddleware(history)),
  );

export default configureStore;
