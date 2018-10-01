import { LOCATION_CHANGE, LocationChangeAction } from "connected-react-router";

import { getType } from "typesafe-actions";
import actions from "../actions";
import { AppsAction } from "../actions/apps";
import { IAppState } from "../shared/types";

const initialState: IAppState = {
  isFetching: false,
  items: [],
  listingAll: false,
};

const appsReducer = (
  state: IAppState = initialState,
  action: AppsAction | LocationChangeAction,
): IAppState => {
  switch (action.type) {
    case getType(actions.apps.requestApps):
      return { ...state, isFetching: true };
    case getType(actions.apps.receiveApps):
      return { ...state, isFetching: false, items: action.payload };
    case getType(actions.apps.errorApps):
      return { ...state, isFetching: false, error: action.payload };
    case getType(actions.apps.errorDeleteApp):
      return { ...state, isFetching: false, deleteError: action.payload };
    case getType(actions.apps.selectApp):
      return { ...state, isFetching: false, selected: action.payload };
    case getType(actions.apps.listApps):
      return { ...state, isFetching: true, listingAll: action.payload };
    case getType(actions.apps.receiveAppList):
      return { ...state, isFetching: false, listOverview: action.payload };
    case LOCATION_CHANGE:
      return {
        ...state,
        deleteError: undefined,
        error: undefined,
        isFetching: false,
        selected: undefined,
      };
    default:
  }
  return state;
};

export default appsReducer;
