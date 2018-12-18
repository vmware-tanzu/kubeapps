import { connectRouter, routerMiddleware } from "connected-react-router";
import { createHashHistory } from "history";
import { applyMiddleware, createStore } from "redux";
import { composeWithDevTools } from "redux-devtools-extension";
import thunkMiddleware from "redux-thunk";

import rootReducer from "../reducers";

export const history = createHashHistory();
export default createStore(
  connectRouter(history)(rootReducer), // add router state to reducer
  composeWithDevTools(
    applyMiddleware(
      thunkMiddleware,
      routerMiddleware(history), // // for dispatching history actions
    ),
  ),
);
