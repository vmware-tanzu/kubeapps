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
    deployChart: (version: IChartVersion, releaseName: string, namespace: string) =>
      dispatch(actions.charts.deployChart(version, releaseName, namespace)),
    fetchChartVersionsAndSelectVersion: (id: string, version?: string) =>
      dispatch(actions.charts.fetchChartVersionsAndSelectVersion(id, version)),
    push: (location: string) => dispatch(push(location)),
    selectChartVersionAndGetReadme: (version: IChartVersion) =>
      dispatch(actions.charts.selectChartVersionAndGetReadme(version)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(ChartView);
