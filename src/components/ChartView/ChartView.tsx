import * as React from "react";
import { RouterAction } from "react-router-redux";

import { IChartState, IChartVersion } from "../../shared/types";
import ChartDeployButton from "./ChartDeployButton";
import ChartHeader from "./ChartHeader";
import ChartMaintainers from "./ChartMaintainers";
import ChartReadme from "./ChartReadme";
import ChartVersionsList from "./ChartVersionsList";
import "./ChartView.css";

interface IChartViewProps {
  chartID: string;
  fetchChartVersionsAndSelectVersion: (id: string, version?: string) => Promise<{}>;
  deployChart: (
    version: IChartVersion,
    releaseName: string,
    namespace: string,
    values: string,
  ) => Promise<{}>;
  push: (location: string) => RouterAction;
  isFetching: boolean;
  selected: IChartState["selected"];
  selectChartVersionAndGetFiles: (version: IChartVersion) => Promise<{}>;
  version: string | undefined;
}

class ChartView extends React.Component<IChartViewProps> {
  public componentDidMount() {
    const { chartID, fetchChartVersionsAndSelectVersion, version } = this.props;
    fetchChartVersionsAndSelectVersion(chartID, version);
  }

  public componentWillReceiveProps(nextProps: IChartViewProps) {
    const { selectChartVersionAndGetFiles, version } = this.props;
    const { versions } = this.props.selected;
    if (nextProps.version !== version) {
      const cv = versions.find(v => v.attributes.version === nextProps.version);
      if (cv) {
        selectChartVersionAndGetFiles(cv);
      } else {
        throw new Error("could not find chart");
      }
    }
  }

  public render() {
    const { isFetching, deployChart, push } = this.props;
    const { version, readme, versions, values } = this.props.selected;
    if (isFetching || !version) {
      return <div>Loading</div>;
    }
    const chartAttrs = version.relationships.chart.data;
    return (
      <section className="ChartView padding-b-big">
        <ChartHeader
          id={`${chartAttrs.repo.name}/${chartAttrs.name}`}
          description={chartAttrs.description}
          icon={chartAttrs.icon}
          repo={chartAttrs.repo.name}
          appVersion={version.attributes.app_version}
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
                    <ChartDeployButton
                      push={push}
                      version={version}
                      deployChart={deployChart}
                      values={values || ""}
                    />
                  </div>
                  <div className="ChartViewSidebar__section">
                    <h2>Chart Versions</h2>
                    <ChartVersionsList selected={version} versions={versions} />
                  </div>
                  {version.attributes.app_version && (
                    <div className="ChartViewSidebar__section">
                      <h2>App Version</h2>
                      <div>{version.attributes.app_version}</div>
                    </div>
                  )}
                  {chartAttrs.home && (
                    <div className="ChartViewSidebar__section">
                      <h2>Home</h2>
                      <div>
                        <a href={chartAttrs.home} target="_blank">
                          {chartAttrs.home}
                        </a>
                      </div>
                    </div>
                  )}
                  <div className="ChartViewSidebar__section">
                    <h2>Maintainers</h2>
                    <ChartMaintainers
                      maintainers={chartAttrs.maintainers}
                      githubIDAsNames={this.isKubernetesCharts(chartAttrs.repo.url)}
                    />
                  </div>
                  {chartAttrs.sources.length > 0 && (
                    <div className="ChartViewSidebar__section">
                      <h2>Related</h2>
                      <div className="ChartSources">
                        <ul className="remove-style padding-l-reset margin-b-reset">
                          {chartAttrs.sources.map((s, i) => (
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
