import {
  AvailablePackageDetail,
  AvailablePackageReference,
  GetAvailablePackageVersionsResponse,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { ThunkAction } from "redux-thunk";
import { ActionType, deprecated } from "typesafe-actions";
import PackagesService from "../shared/PackagesService";
import {
  FetchError,
  IReceivePackagesActionPayload as IReceiveAvailablePackageSummariesActionPayload,
  IStoreState,
} from "../shared/types";

const { createAction } = deprecated;

// ** AvailablePackageSummaries actions **
// related to the list of available packages (aka Catalog)

// Request action
export const requestAvailablePackageSummaries = createAction(
  "REQUEST_AVAILABLE_PACKAGE_SUMMARIES",
  resolve => {
    return (page?: number) => resolve(page);
  },
);

// Receive action
export const receiveAvailablePackageSummaries = createAction(
  "RECEIVE_AVAILABLE_PACKAGE_SUMMARIES",
  resolve => {
    return (payload: IReceiveAvailablePackageSummariesActionPayload) => resolve(payload);
  },
);

// Reset action
export const resetAvailablePackageSummaries = createAction("RESET_AVAILABLE_PACKAGE_SUMMARIES");

// ** SelectedAvailablePackage actions **
// related to the selected package in the state (package detail and list of versions)

// Request action
export const requestSelectedAvailablePackageDetail = createAction(
  "REQUEST_SELECTED_AVAILABLE_PACKAGE_DETAIL",
);

// Receive action
export const receiveSelectedAvailablePackageDetail = createAction(
  "RECEIVE_SELECTED_AVAILABLE_PACKAGE_DETAIL",
  resolve => {
    return (selectedPackage: AvailablePackageDetail) => resolve({ selectedPackage });
  },
);

// Reset action
export const resetSelectedAvailablePackageDetail = createAction("RESET_PACKAGE_VERSION");

// Request action
export const requestSelectedAvailablePackageVersions = createAction(
  "REQUEST_SELECTED_AVAILABLE_PACKAGE_VERSIONS",
);

// Receive action
export const receiveSelectedAvailablePackageVersions = createAction(
  "RECEIVE_SELECTED_AVAILABLE_PACKAGE_VERSIONS",
  resolve => {
    return (versions: GetAvailablePackageVersionsResponse) => resolve(versions);
  },
);

// No reset action

// ** Error actions **
// for handling the erros thrown by the rest of the actions

// Create action
export const createErrorPackage = createAction("CREATE_ERROR_PACKAGE", resolve => {
  return (err: Error) => resolve(err);
});

// Reset action
export const clearErrorPackage = createAction("CLEAR_ERROR_PACKAGE");

const allActions = [
  requestAvailablePackageSummaries,
  receiveAvailablePackageSummaries,
  resetAvailablePackageSummaries,
  requestSelectedAvailablePackageDetail,
  receiveSelectedAvailablePackageDetail,
  resetSelectedAvailablePackageDetail,
  requestSelectedAvailablePackageVersions,
  receiveSelectedAvailablePackageVersions,
  createErrorPackage,
  clearErrorPackage,
];

export type PackagesAction = ActionType<typeof allActions[number]>;

export function fetchAvailablePackageSummaries(
  cluster: string,
  namespace: string,
  repos: string,
  page: number,
  size: number,
  query?: string,
): ThunkAction<Promise<void>, IStoreState, null, PackagesAction> {
  return async dispatch => {
    dispatch(requestAvailablePackageSummaries(page));
    try {
      const response = await PackagesService.getAvailablePackageSummaries(
        cluster,
        namespace,
        repos,
        page,
        size,
        query,
      );
      dispatch(receiveAvailablePackageSummaries({ response, page }));
    } catch (e: any) {
      dispatch(createErrorPackage(new FetchError(e.message)));
    }
  };
}

export function fetchAvailablePackageVersions(
  availablePackageReference?: AvailablePackageReference,
): ThunkAction<Promise<void>, IStoreState, null, PackagesAction> {
  return async dispatch => {
    dispatch(requestSelectedAvailablePackageVersions());
    try {
      const response = await PackagesService.getAvailablePackageVersions(availablePackageReference);
      dispatch(receiveSelectedAvailablePackageVersions(response));
    } catch (e: any) {
      dispatch(createErrorPackage(new FetchError(e.message)));
    }
  };
}

export function fetchAndSelectAvailablePackageDetail(
  availablePackageReference?: AvailablePackageReference,
  version?: string,
): ThunkAction<Promise<void>, IStoreState, null, PackagesAction> {
  return async dispatch => {
    try {
      dispatch(requestSelectedAvailablePackageDetail());
      const response = await PackagesService.getAvailablePackageDetail(
        availablePackageReference,
        version,
      );
      if (response.availablePackageDetail?.version?.pkgVersion) {
        dispatch(receiveSelectedAvailablePackageDetail(response.availablePackageDetail));
      } else {
        dispatch(createErrorPackage(new FetchError("could not find package version")));
      }
    } catch (e: any) {
      dispatch(createErrorPackage(new FetchError(e.message)));
    }
  };
}
