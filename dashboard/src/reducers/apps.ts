import { LOCATION_CHANGE, LocationChangeAction } from "connected-react-router";
import { cloneDeep } from "lodash";

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
    case getType(actions.apps.receiveAppUpdateInfo):
      const stateCopy = cloneDeep(state);
      if (stateCopy.listOverview) {
        // TODO: Review structure to use byID and update items directly
        const appOverviewIndex = stateCopy.listOverview.findIndex(
          a => a.releaseName === action.payload.releaseName,
        );
        stateCopy.listOverview[appOverviewIndex] = {
          ...stateCopy.listOverview[appOverviewIndex],
          updateInfo: action.payload.updateInfo,
        };
      }
      if (stateCopy.selected && stateCopy.selected.name === action.payload.releaseName) {
        stateCopy.selected.updateInfo = action.payload.updateInfo;
      }
      return { ...stateCopy };
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
