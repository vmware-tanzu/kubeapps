import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import ChartView from "../../components/ChartView";
import { IChartVersion, IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      repo: string;
      id: string;
      version?: string;
    };
  };
}

function mapStateToProps({ charts }: IStoreState, { match: { params } }: IRouteProps) {
  return {
    chartID: chartID(params),
    isFetching: charts.isFetching,
    selected: charts.selected,
    version: params.version,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>, { match: { params } }: IRouteProps) {
  return {
    fetchChartVersionsAndSelectVersion: (id: string, version?: string) =>
      dispatch(actions.charts.fetchChartVersionsAndSelectVersion(id, version)),
    getChartReadme: (version: string) =>
      dispatch(actions.charts.getChartReadme(chartID(params), version)),
    resetChartVersion: () => dispatch(actions.charts.resetChartVersion()),
    selectChartVersion: (version: IChartVersion) =>
      dispatch(actions.charts.selectChartVersion(version)),
  };
}

function chartID(params: IRouteProps["match"]["params"]) {
  return `${params.repo}/${params.id}`;
}

export default connect(mapStateToProps, mapDispatchToProps)(ChartView);
