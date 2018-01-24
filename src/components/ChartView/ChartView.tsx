import * as React from "react";
import { RouterAction } from "react-router-redux";

import { IChart, IChartState } from "../../shared/types";
import ChartDeployButton from "./ChartDeployButton";
import ChartHeader from "./ChartHeader";
import ChartMaintainers from "./ChartMaintainers";
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
      <section className="ChartView padding-b-big">
        <ChartHeader
          id={chart.id}
          description={chart.attributes.description}
          icon={chart.attributes.icon}
          repo={chart.attributes.repo.name}
          appVersion={chart.relationships.latestChartVersion.data.app_version}
        />
        <main>
          <div className="container container-fluid">
            <div className="row">
              <div className="col-9 ChartView__readme-container">
                <ChartReadme markdown={readme} />
              </div>
              <div className="col-3 ChartView__sidebar-container">
                <aside className="ChartViewSidebar bg-light margin-v-big padding-h-normal padding-b-normal">
                  <div className="ChartViewSidebar__section">
                    <h2>Usage</h2>
                    <ChartDeployButton push={push} chart={chart} deployChart={deployChart} />
                  </div>
                  <div className="ChartViewSidebar__section">
                    <h2>Chart Versions</h2>
                    <ChartVersionsList versions={versions} />
                  </div>
                  {chart.relationships.latestChartVersion.data.app_version && (
                    <div className="ChartViewSidebar__section">
                      <h2>App Version</h2>
                      <div>{chart.relationships.latestChartVersion.data.app_version}</div>
                    </div>
                  )}
                  {chart.attributes.home && (
                    <div className="ChartViewSidebar__section">
                      <h2>Home</h2>
                      <div>
                        <a href={chart.attributes.home} target="_blank">
                          {chart.attributes.home}
                        </a>
                      </div>
                    </div>
                  )}
                  <div className="ChartViewSidebar__section">
                    <h2>Maintainers</h2>
                    <ChartMaintainers
                      maintainers={chart.attributes.maintainers}
                      githubIDAsNames={this.isKubernetesCharts(chart.attributes.repo.url)}
                    />
                  </div>
                  {chart.attributes.sources.length > 0 && (
                    <div className="ChartViewSidebar__section">
                      <h2>Related</h2>
                      <div className="ChartSources">
                        <ul className="remove-style padding-l-reset margin-b-reset">
                          {chart.attributes.sources.map((s, i) => (
                            <li key={i}>
                              <a href={s} target="_blank">
                                {s}
                              </a>
                            </li>
                          ))}
                        </ul>
                      </div>
                    </div>
                  )}
                </aside>
              </div>
            </div>
          </div>
        </main>
      </section>
    );
  }

  private isKubernetesCharts(repoURL: string) {
    return (
      repoURL === "https://kubernetes-charts.storage.googleapis.com" ||
      repoURL === "https://kubernetes-charts-incubator.storage.googleapis.com"
    );
  }
}

export default ChartView;
