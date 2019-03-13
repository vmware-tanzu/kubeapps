import * as React from "react";

import { IChartState, IChartVersion } from "../../shared/types";
import { ErrorSelector } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import ChartHeader from "./ChartHeader";
import ChartMaintainers from "./ChartMaintainers";
import ChartReadme from "./ChartReadme";
import ChartVersionsList from "./ChartVersionsList";
import "./ChartView.css";

interface IChartViewProps {
  chartID: string;
  fetchChartVersionsAndSelectVersion: (id: string, version?: string) => void;
  isFetching: boolean;
  selected: IChartState["selected"];
  selectChartVersion: (version: IChartVersion) => any;
  resetChartVersion: () => any;
  getChartReadme: (version: string) => any;
  namespace: string;
  version: string | undefined;
}

class ChartView extends React.Component<IChartViewProps> {
  public componentDidMount() {
    const { chartID, fetchChartVersionsAndSelectVersion, version } = this.props;
    fetchChartVersionsAndSelectVersion(chartID, version);
  }

  public componentWillReceiveProps(nextProps: IChartViewProps) {
    const { selectChartVersion, version } = this.props;
    const { versions } = this.props.selected;
    if (nextProps.version !== version) {
      const cv = versions.find(v => v.attributes.version === nextProps.version);
      if (cv) {
        selectChartVersion(cv);
      } else {
        throw new Error("could not find chart");
      }
    }
  }

  public componentWillUnmount() {
    this.props.resetChartVersion();
  }

  public render() {
    const { isFetching, getChartReadme, namespace, chartID } = this.props;
    const { version, readme, error, readmeError, versions } = this.props.selected;
    if (error) {
      return <ErrorSelector error={error} resource={`Chart ${chartID}`} />;
    }
    if (isFetching || !version) {
      return <LoadingWrapper />;
    }
    const chartAttrs = version.relationships.chart.data;
    return (
      <section className="ChartView padding-b-big">
        <ChartHeader
          id={`${chartAttrs.repo.name}/${chartAttrs.name}`}
          description={chartAttrs.description}
          icon={chartAttrs.icon}
          repo={chartAttrs.repo.name}
          version={version}
          namespace={namespace}
        />
        <main>
          <div className="container container-fluid">
            <div className="row">
              <div className="col-9 ChartView__readme-container">
                <ChartReadme
                  getChartReadme={getChartReadme}
                  readme={readme}
                  hasError={!!readmeError}
                  version={version.attributes.version}
                />
              </div>
              <div className="col-3 ChartView__sidebar-container">
                <aside className="ChartViewSidebar bg-light margin-v-big padding-h-normal padding-b-normal">
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
                        <ul className="remove-style padding-l-reset margin-b-reset">
                          <li>
                            <a href={chartAttrs.home} target="_blank">
                              {chartAttrs.home}
                            </a>
                          </li>
                        </ul>
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
