import { getType } from "typesafe-actions";

import actions from "../actions";
import { AppsAction } from "../actions/apps";
import { IAppState } from "../shared/types";

const initialState: IAppState = {
  isFetching: false,
  items: [],
};

const appsReducer = (state: IAppState = initialState, action: AppsAction): IAppState => {
  switch (action.type) {
    case getType(actions.apps.requestApps):
      return { ...state, isFetching: true };
    case getType(actions.apps.receiveApps):
      return { ...state, isFetching: false, items: action.apps };
    case getType(actions.apps.selectApp):
      return { ...state, isFetching: false, selected: action.app };
    default:
  }
  return state;
};

export default appsReducer;
