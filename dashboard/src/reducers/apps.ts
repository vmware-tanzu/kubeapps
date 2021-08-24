import { LocationChangeAction, LOCATION_CHANGE } from "connected-react-router";
import { IAppState } from "shared/types";
import { getType } from "typesafe-actions";
import actions from "../actions";
import { AppsAction } from "../actions/apps";

export const initialState: IAppState = {
  isFetching: false,
  items: [],
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
    case getType(actions.apps.errorApp):
      return { ...state, isFetching: false, error: action.payload };
    case getType(actions.apps.selectApp):
      return {
        ...state,
        isFetching: false,
        selected: action.payload.app,
        selectedDetails: action.payload.details,
      };
    case getType(actions.apps.listApps):
      return { ...state, isFetching: true };
    case getType(actions.apps.receiveAppList):
      return { ...state, isFetching: false, listOverview: action.payload };
    case getType(actions.apps.requestDeleteApp):
      return { ...state, isFetching: true };
    case getType(actions.apps.receiveDeleteApp):
      return { ...state, isFetching: false };
    case getType(actions.apps.requestDeployApp):
      return { ...state, isFetching: true };
    case getType(actions.apps.receiveDeployApp):
      return { ...state, isFetching: false };
    case getType(actions.apps.requestRollbackApp):
      return { ...state, isFetching: true };
    case getType(actions.apps.receiveRollbackApp):
      return { ...state, isFetching: false };
    case LOCATION_CHANGE:
      return {
        ...state,
        error: undefined,
        isFetching: false,
        selected: undefined,
      };
    default:
  }
  return state;
};

export default appsReducer;
