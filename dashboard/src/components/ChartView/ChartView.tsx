import { useEffect } from "react";
import { Dispatch } from "redux";

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import ChartSummary from "components/Catalog/ChartSummary";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { useDispatch } from "react-redux";
import { Link } from "react-router-dom";
import { app } from "shared/url";
import { IChartState, IChartVersion } from "../../shared/types";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import ChartHeader from "./ChartHeader";
import ChartReadme from "./ChartReadme";

export interface IChartViewProps {
  chartID: string;
  chartNamespace: string;
  isFetching: boolean;
  selected: IChartState["selected"];
  namespace: string;
  cluster: string;
  version: string | undefined;
  kubeappsNamespace: string;
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
  kubeappsNamespace,
}: IChartViewProps) {
  const dispatch = useDispatch();
  const { version, readme, error, readmeError, versions } = selected;
  useEffect(() => {
    dispatch(
      actions.charts.fetchChartVersionsAndSelectVersion(
        cluster,
        chartNamespace,
        chartID,
        versionStr,
      ),
    );
    return () => {
      dispatch(actions.charts.resetChartVersion());
    };
  }, [cluster, chartNamespace, chartID, versionStr, dispatch]);

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
      <div>
        <ChartHeader
          chartAttrs={chartAttrs}
          versions={versions}
          onSelect={selectVersion}
          deployButton={
            <Link
              to={app.apps.new(
                cluster,
                namespace,
                version,
                version.attributes.version,
                kubeappsNamespace,
              )}
            >
              <CdsButton status="primary">
                <CdsIcon shape="deploy" /> Deploy
              </CdsButton>
            </Link>
          }
          selectedVersion={selected.version?.attributes.version}
        />
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
              cluster={cluster}
              namespace={chartNamespace}
              chartID={chartID}
            />
            <div className="after-readme-button">
              <Link
                to={app.apps.new(
                  cluster,
                  namespace,
                  version,
                  version.attributes.version,
                  kubeappsNamespace,
                )}
              >
                <CdsButton status="primary">
                  <CdsIcon shape="deploy" /> Deploy
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
