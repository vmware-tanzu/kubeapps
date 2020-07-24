import React, { useEffect } from "react";

import ChartSummary from "components/Catalog/ChartSummary";
import { CdsButton, CdsIcon } from "components/Clarity/clarity";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import Tooltip from "components/js/Tooltip";
import PageHeader from "components/PageHeader/PageHeader.v2";
import { Link } from "react-router-dom";
import { app } from "shared/url";
import helmIcon from "../../icons/helm.svg";
import placeholder from "../../placeholder.png";
import { IChartState, IChartVersion } from "../../shared/types";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper.v2";
import ChartReadme from "./ChartReadme.v2";
import "./ChartView.v2.css";

export interface IChartViewProps {
  chartID: string;
  chartNamespace: string;
  fetchChartVersionsAndSelectVersion: (namespace: string, id: string, version?: string) => void;
  isFetching: boolean;
  selected: IChartState["selected"];
  selectChartVersion: (version: IChartVersion) => any;
  resetChartVersion: () => any;
  namespace: string;
  cluster: string;
  version: string | undefined;
}

function callSelectChartVersion(
  ver: string,
  versions: IChartVersion[],
  selectChartVersion: (version: IChartVersion) => any,
) {
  const cv = versions.find(v => v.attributes.version === ver);
  if (cv) {
    selectChartVersion(cv);
  }
}

function ChartView({
  chartID,
  chartNamespace,
  fetchChartVersionsAndSelectVersion,
  version: versionStr,
  selected,
  selectChartVersion,
  isFetching,
  cluster,
  namespace,
  resetChartVersion,
}: IChartViewProps) {
  const { version, readme, error, readmeError, versions } = selected;
  useEffect(() => {
    fetchChartVersionsAndSelectVersion(chartNamespace, chartID, versionStr);
    return resetChartVersion;
  }, [fetchChartVersionsAndSelectVersion, chartNamespace, chartID, versionStr, resetChartVersion]);

  useEffect(() => {
    callSelectChartVersion(versionStr || "", versions, selectChartVersion);
  }, [versions, versionStr, selectChartVersion, resetChartVersion]);

  const selectVersion = (event: React.ChangeEvent<HTMLSelectElement>) =>
    callSelectChartVersion(event.target.value, versions, selectChartVersion);

  if (error) {
    return <Alert theme="danger">Unable to fetch chart: {error.message}</Alert>;
  }
  if (isFetching || !version) {
    return <LoadingWrapper loaded={false} />;
  }
  const chartAttrs = version.relationships.chart.data;
  return (
    <section>
      <PageHeader>
        <div className="kubeapps-header-content">
          <Row>
            <Column span={7}>
              <Row>
                <img
                  src={chartAttrs.icon ? `api/assetsvc/${chartAttrs.icon}` : placeholder}
                  alt="app-icon"
                />
                <div className="kubeapps-title-block">
                  <h3>{`${chartAttrs.repo.name}/${chartAttrs.name}`}</h3>
                  <div className="kubeapps-header-subtitle">
                    <img src={helmIcon} alt="helm-icon" />
                    <span>Helm Chart</span>
                  </div>
                </div>
              </Row>
            </Column>
            <Column span={5}>
              <div className="control-buttons">
                <div className="chart-version-selector">
                  <label className="chart-version-selector-label" htmlFor="chart-versions">
                    Chart Version{" "}
                    <Tooltip
                      label="chart-versions-tooltip"
                      id="chart-versions-tooltip"
                      position="bottom-left"
                      iconProps={{ solid: true, size: "sm" }}
                    >
                      Chart and App versions can be increased independently.{" "}
                      <a
                        href="https://helm.sh/docs/topics/charts/#charts-and-versioning"
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        More info here
                      </a>
                      .{" "}
                    </Tooltip>
                  </label>
                  <div className="clr-select-wrapper">
                    <select
                      name="chart-versions"
                      className="clr-page-size-select"
                      onChange={selectVersion}
                    >
                      {versions.map(v => {
                        return (
                          <option
                            key={`chart-version-selector-${v.attributes.version}`}
                            value={v.attributes.version}
                          >
                            {v.attributes.version} / App Version {v.attributes.app_version}
                          </option>
                        );
                      })}
                    </select>
                  </div>
                </div>
                <div className="header-button">
                  <Link to={app.apps.new(cluster, namespace, version, version.attributes.version)}>
                    <CdsButton status="primary">
                      <CdsIcon shape="deploy" inverse={true} /> Deploy
                    </CdsButton>
                  </Link>
                </div>
              </div>
            </Column>
          </Row>
        </div>
      </PageHeader>
      <section>
        <Row>
          <Column span={3}>
            <ChartSummary version={version} chartAttrs={chartAttrs} />
          </Column>
          <Column span={9}>
            <ChartReadme
              readme={readme}
              error={readmeError}
              version={version.attributes.version}
              namespace={chartNamespace}
              chartID={chartID}
            />
            <div className="after-readme-button">
              <Link to={app.apps.new(cluster, namespace, version, version.attributes.version)}>
                <CdsButton status="primary">
                  <CdsIcon shape="deploy" inverse={true} /> Deploy
                </CdsButton>
              </Link>
            </div>
          </Column>
        </Row>
      </section>
    </section>
  );
}

export default ChartView;
