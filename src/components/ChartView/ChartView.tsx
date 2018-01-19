import * as React from "react";
import { RouterAction } from "react-router-redux";

import { IChart, IChartState } from "../../shared/types";
import ChartDeployButton from "./ChartDeployButton";
import ChartHeader from "./ChartHeader";
import ChartReadme from "./ChartReadme";
import ChartVersionsList from "./ChartVersionsList";
import "./ChartView.css";

interface IChartViewProps {
  chartID: string;
  getChart: (id: string) => Promise<{}>;
  deployChart: (chart: IChart, releaseName: string, namespace: string) => Promise<{}>;
  push: (location: string) => RouterAction;
  isFetching: boolean;
  selected: IChartState["selected"];
}

class ChartView extends React.Component<IChartViewProps> {
  public componentDidMount() {
    const { chartID, getChart } = this.props;
    getChart(chartID);
  }

  public render() {
    const { isFetching, deployChart, push } = this.props;
    const { chart, readme, versions } = this.props.selected;
    if (isFetching || !chart) {
      return <div>Loading</div>;
    }
    return (
      <section className="ChartView">
        <ChartHeader
          id={chart.id}
          description={chart.attributes.description}
          icon={chart.attributes.icon}
          repo={chart.attributes.repo.name}
          appVersion={chart.relationships.latestChartVersion.data.app_version}
        />
        <main>
          <div className="container">
            <div className="row">
              <div className="col-8 ChartView__readme-container">
                <ChartReadme markdown={readme} />
              </div>
              {/* TODO: fix when upgrading to bitnami-ui v3 - col-4 does not fit correctly in v2 */}
              <div className="col-3">
                <aside className="ChartViewSidebar bg-light margin-v-big padding-h-normal">
                  <div className="ChartViewSidebar__section">
                    <h2>Usage</h2>
                    <ChartDeployButton push={push} chart={chart} deployChart={deployChart} />
                  </div>
                  <div className="ChartViewSidebar__section">
                    <h2>Chart Versions</h2>
                    <ChartVersionsList versions={versions} />
                  </div>
                  <div className="ChartViewSidebar__section">
                    <h2>App Version</h2>
                    <div>{chart.relationships.latestChartVersion.data.app_version}</div>
                  </div>
                  <div className="ChartViewSidebar__section">
                    <h2>Home</h2>
                    <div>{chart.attributes.home}</div>
                  </div>
                  <div className="ChartViewSidebar__section">
                    <h2>Maintainers</h2>
                    {chart.attributes.maintainers.map((m, i) => <div key={i}>{m.name}</div>)}
                  </div>
                  <div className="ChartViewSidebar__section">
                    <h2>Related</h2>
                    {chart.attributes.sources.map((s, i) => <div key={i}>{s}</div>)}
                  </div>
                </aside>
              </div>
            </div>
          </div>
        </main>
      </section>
    );
  }
}

export default ChartView;
