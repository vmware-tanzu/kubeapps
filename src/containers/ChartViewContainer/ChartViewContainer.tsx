import { connect } from "react-redux";
import { Dispatch } from "redux";

import { push } from "react-router-redux";
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
    chartID: `${params.repo}/${params.id}`,
    isFetching: charts.isFetching,
    selected: charts.selected,
    version: params.version,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    deployChart: (version: IChartVersion, releaseName: string, namespace: string, values: string) =>
      dispatch(actions.charts.deployChart(version, releaseName, namespace, values)),
    fetchChartVersionsAndSelectVersion: (id: string, version?: string) =>
      dispatch(actions.charts.fetchChartVersionsAndSelectVersion(id, version)),
    push: (location: string) => dispatch(push(location)),
    selectChartVersionAndGetFiles: (version: IChartVersion) =>
      dispatch(actions.charts.selectChartVersionAndGetFiles(version)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(ChartView);
