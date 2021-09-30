import {
  AvailablePackageDetail,
  AvailablePackageReference,
  GetAvailablePackageVersionsResponse,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { ThunkAction } from "redux-thunk";
import { ActionType, deprecated } from "typesafe-actions";
import PackagesService from "../shared/PackagesService";
import { FetchError, IReceiveChartsActionPayload, IStoreState } from "../shared/types";

const { createAction } = deprecated;

export const requestCharts = createAction("REQUEST_CHARTS", resolve => {
  return (page?: number) => resolve(page);
});

export const receiveCharts = createAction("RECEIVE_CHARTS", resolve => {
  return (payload: IReceiveChartsActionPayload) => resolve(payload);
});

export const receiveChartVersions = createAction("RECEIVE_CHART_VERSIONS", resolve => {
  return (versions: GetAvailablePackageVersionsResponse) => resolve(versions);
});

export const errorChart = createAction("ERROR_CHART", resolve => {
  return (err: Error) => resolve(err);
});

export const clearErrorChart = createAction("CLEAR_ERROR_CHART");

export const selectChartVersion = createAction("SELECT_CHART_VERSION", resolve => {
  return (selectedPackage: AvailablePackageDetail) => resolve({ selectedPackage });
});

export const requestDeployedChartVersion = createAction("REQUEST_DEPLOYED_CHART_VERSION");

export const receiveDeployedChartVersion = createAction(
  "RECEIVE_DEPLOYED_CHART_VERSION",
  resolve => {
    return (chartVersion: AvailablePackageDetail, values?: string, schema?: string) =>
      resolve({ chartVersion, values, schema });
  },
);

export const resetChartVersion = createAction("RESET_CHART_VERSION");

export const resetRequestCharts = createAction("RESET_REQUEST_CHARTS");

const allActions = [
  requestCharts,
  errorChart,
  clearErrorChart,
  receiveCharts,
  receiveChartVersions,
  selectChartVersion,
  requestDeployedChartVersion,
  receiveDeployedChartVersion,
  resetChartVersion,
  resetRequestCharts,
];

export type ChartsAction = ActionType<typeof allActions[number]>;

export function fetchCharts(
  cluster: string,
  namespace: string,
  repos: string,
  page: number,
  size: number,
  query?: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    dispatch(requestCharts(page));
    try {
      const response = await PackagesService.getAvailablePackageSummaries(
        cluster,
        namespace,
        repos,
        page,
        size,
        query,
      );
      dispatch(receiveCharts({ response, page }));
    } catch (e: any) {
      dispatch(errorChart(new FetchError(e.message)));
    }
  };
}

export function fetchChartVersions(
  availablePackageReference?: AvailablePackageReference,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    dispatch(requestCharts());
    try {
      const response = await PackagesService.getAvailablePackageVersions(availablePackageReference);
      dispatch(receiveChartVersions(response));
    } catch (e: any) {
      dispatch(errorChart(new FetchError(e.message)));
    }
  };
}

export function fetchChartVersion(
  availablePackageReference?: AvailablePackageReference,
  version?: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    try {
      const response = await PackagesService.getAvailablePackageDetail(
        availablePackageReference,
        version,
      );
      if (response.availablePackageDetail?.version?.pkgVersion) {
        dispatch(selectChartVersion(response.availablePackageDetail));
      } else {
        dispatch(errorChart(new FetchError("could not find package version")));
      }
    } catch (e: any) {
      dispatch(errorChart(new FetchError(e.message)));
    }
  };
}

export function getDeployedChartVersion(
  availablePackageReference?: AvailablePackageReference,
  version?: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    try {
      dispatch(requestDeployedChartVersion());
      const response = await PackagesService.getAvailablePackageDetail(
        availablePackageReference,
        version,
      );
      if (response.availablePackageDetail) {
        dispatch(
          receiveDeployedChartVersion(
            response.availablePackageDetail,
            response.availablePackageDetail.defaultValues,
            response.availablePackageDetail.valuesSchema,
          ),
        );
      }
    } catch (e: any) {
      dispatch(errorChart(new FetchError(e.message)));
    }
  };
}
