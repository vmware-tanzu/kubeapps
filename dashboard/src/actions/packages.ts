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

export const requestAvailablePackageSummaries = createAction(
  "REQUEST_AVAILABLE_PACKAGE_SUMMARIES",
  resolve => {
    return (page?: number) => resolve(page);
  },
);

export const receiveAvailablePackageSummaries = createAction(
  "RECEIVE_AVAILABLE_PACKAGE_SUMMARIES",
  resolve => {
    return (payload: IReceiveAvailablePackageSummariesActionPayload) => resolve(payload);
  },
);

export const receiveAvailablePackageVersions = createAction(
  "RECEIVE_AVAILABLE_PACKAGE_VERSIONS",
  resolve => {
    return (versions: GetAvailablePackageVersionsResponse) => resolve(versions);
  },
);

export const errorPackage = createAction("ERROR_PACKAGE", resolve => {
  return (err: Error) => resolve(err);
});

export const clearErrorPackage = createAction("CLEAR_ERROR_PACKAGE");

export const receiveSelectedAvailablePackageDetail = createAction(
  "SELECT_AVAILABLE_PACKAGE_DETAIL",
  resolve => {
    return (selectedPackage: AvailablePackageDetail) => resolve({ selectedPackage });
  },
);

export const requestDeployedAvailablePackageDetail = createAction(
  "REQUEST_DEPLOYED_AVAILABLE_PACKAGE_DETAIL",
);

export const receiveDeployedAvailablePackageDetail = createAction(
  "RECEIVE_DEPLOYED_AVAILABLE_PACKAGE_DETAIL",
  resolve => {
    return (availablePackageDetail: AvailablePackageDetail) => resolve({ availablePackageDetail });
  },
);

export const resetChartVersion = createAction("RESET_CHART_VERSION");

export const resetAvailablePackageSummaries = createAction("RESET_AVAILABLE_PACKAGE_SUMMARIES");

const allActions = [
  requestAvailablePackageSummaries,
  errorPackage,
  clearErrorPackage,
  receiveAvailablePackageSummaries,
  receiveAvailablePackageVersions,
  receiveSelectedAvailablePackageDetail,
  requestDeployedAvailablePackageDetail,
  receiveDeployedAvailablePackageDetail,
  resetChartVersion,
  resetAvailablePackageSummaries,
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
      dispatch(errorPackage(new FetchError(e.message)));
    }
  };
}

export function fetchAvailablePackageVersions(
  availablePackageReference?: AvailablePackageReference,
): ThunkAction<Promise<void>, IStoreState, null, PackagesAction> {
  return async dispatch => {
    dispatch(requestAvailablePackageSummaries());
    try {
      const response = await PackagesService.getAvailablePackageVersions(availablePackageReference);
      dispatch(receiveAvailablePackageVersions(response));
    } catch (e: any) {
      dispatch(errorPackage(new FetchError(e.message)));
    }
  };
}

export function fetchAndSelectAvailablePackageDetail(
  availablePackageReference?: AvailablePackageReference,
  version?: string,
): ThunkAction<Promise<void>, IStoreState, null, PackagesAction> {
  return async dispatch => {
    try {
      const response = await PackagesService.getAvailablePackageDetail(
        availablePackageReference,
        version,
      );
      if (response.availablePackageDetail?.version?.pkgVersion) {
        dispatch(receiveSelectedAvailablePackageDetail(response.availablePackageDetail));
      } else {
        dispatch(errorPackage(new FetchError("could not find package version")));
      }
    } catch (e: any) {
      dispatch(errorPackage(new FetchError(e.message)));
    }
  };
}

export function fetchDeployedAvailablePackageDetail(
  availablePackageReference?: AvailablePackageReference,
  version?: string,
): ThunkAction<Promise<void>, IStoreState, null, PackagesAction> {
  return async dispatch => {
    try {
      dispatch(requestDeployedAvailablePackageDetail());
      const response = await PackagesService.getAvailablePackageDetail(
        availablePackageReference,
        version,
      );
      if (response.availablePackageDetail) {
        dispatch(receiveDeployedAvailablePackageDetail(response.availablePackageDetail));
      }
    } catch (e: any) {
      dispatch(errorPackage(new FetchError(e.message)));
    }
  };
}
