import { connect } from "react-redux";
import { Dispatch } from "redux";

import { push } from "react-router-redux";
import actions from "../../actions";
import ChartView from "../../components/ChartView";
import { IChart, IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      repo: string;
      id: string;
    };
  };
}

function mapStateToProps({ charts }: IStoreState, { match: { params } }: IRouteProps) {
  return {
    chart: charts.selectedChart,
    chartID: `${params.repo}/${params.id}`,
    isFetching: charts.isFetching,
    version: charts.selectedVersion,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    deployChart: (chart: IChart, releaseName: string, namespace: string) =>
      dispatch(actions.charts.deployChart(chart, releaseName, namespace)),
    getChart: (id: string) => dispatch(actions.charts.getChart(id)),
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(ChartView);
