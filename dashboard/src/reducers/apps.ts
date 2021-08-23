import { LocationChangeAction, LOCATION_CHANGE } from "connected-react-router";
import { IAppState, IRelease } from "shared/types";
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
      return { ...state, isFetching: false, selected: action.payload };
    case getType(actions.apps.listApps):
      return { ...state, isFetching: true };
    case getType(actions.apps.receiveAppList):
      return { ...state, isFetching: false, listOverview: action.payload };
    case getType(actions.apps.requestAppUpdateInfo):
      return { ...state, isFetching: true };
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
    case getType(actions.apps.receiveAppUpdateInfo): {
      let listOverview;
      if (state.listOverview) {
        // TODO: Review structure to use byID and update items directly
        const appOverviewIndex = state.listOverview.findIndex(
          a => a.releaseName === action.payload.releaseName,
        );
        // If the release exists in the listOverview, update the updateInfo
        // field. If it doesn't exist, we don't do anything, it will get fetched
        // when loading the app list
        if (appOverviewIndex >= 0) {
          listOverview = [
            ...state.listOverview.slice(0, appOverviewIndex),
            { ...state.listOverview[appOverviewIndex], updateInfo: action.payload.updateInfo },
            ...state.listOverview.slice(appOverviewIndex + 1),
          ];
        }
      }
      let selected;
      if (state.selected && state.selected.name === action.payload.releaseName) {
        // TODO(andres) It's required to convert as IRelease to avoid missing toJSON property
        selected = { ...state.selected, updateInfo: action.payload.updateInfo } as IRelease;
      }
      return {
        ...state,
        isFetching: false,
        listOverview: listOverview || state.listOverview,
        selected: selected || state.selected,
      };
    }
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
