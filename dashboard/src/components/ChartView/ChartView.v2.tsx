import React, { useEffect } from "react";
import { Dispatch } from "redux";

import actions from "actions";
import ChartSummary from "components/Catalog/ChartSummary";
import { CdsButton, CdsIcon } from "components/Clarity/clarity";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { useDispatch } from "react-redux";
import { Link } from "react-router-dom";
import { app } from "shared/url";
import { IChartState, IChartVersion } from "../../shared/types";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper.v2";
import ChartHeader from "./ChartHeader.v2";
import ChartReadme from "./ChartReadme.v2";

export interface IChartViewProps {
  chartID: string;
  chartNamespace: string;
  isFetching: boolean;
  selected: IChartState["selected"];
  namespace: string;
  cluster: string;
  version: string | undefined;
}

function callSelectChartVersion(ver: string, versions: IChartVersion[], dispatch: Dispatch) {
  const cv = versions.find(v => v.attributes.version === ver);
  if (cv) {
    dispatch(actions.charts.selectChartVersion(cv));
  }
}

function ChartView({
  chartID,
  chartNamespace,
  version: versionStr,
  selected,
  isFetching,
  cluster,
  namespace,
}: IChartViewProps) {
  const dispatch = useDispatch();
  const { version, readme, error, readmeError, versions } = selected;
  useEffect(() => {
    dispatch(
      actions.charts.fetchChartVersionsAndSelectVersion(chartNamespace, chartID, versionStr),
    );
    return () => {
      dispatch(actions.charts.resetChartVersion());
    };
  }, [chartNamespace, chartID, versionStr, dispatch]);

  useEffect(() => {
    callSelectChartVersion(versionStr || "", versions, dispatch);
  }, [versions, versionStr, dispatch]);

  if (error) {
    return <Alert theme="danger">Unable to fetch chart: {error.message}</Alert>;
  }
  if (isFetching || !version) {
    return <LoadingWrapper loaded={false} />;
  }
  const chartAttrs = version.relationships.chart.data;
  const selectVersion = (event: React.ChangeEvent<HTMLSelectElement>) =>
    callSelectChartVersion(event.target.value, versions, dispatch);

  return (
    <section>
      <div className="header-button">
        <ChartHeader chartAttrs={chartAttrs} versions={versions} onSelect={selectVersion}>
          <Link to={app.apps.new(cluster, namespace, version, version.attributes.version)}>
            <CdsButton status="primary">
              <CdsIcon shape="deploy" inverse={true} /> Deploy
            </CdsButton>
          </Link>
        </ChartHeader>
      </div>

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
