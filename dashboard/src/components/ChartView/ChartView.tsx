import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import AvailablePackageDetailExcerpt from "components/Catalog/AvailablePackageDetailExcerpt";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import LoadingWrapper from "components/LoadingWrapper";
import { push } from "connected-react-router";
import { AvailablePackageReference } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import * as ReactRouter from "react-router";
import { Link } from "react-router-dom";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import { app } from "shared/url";
import ChartHeader from "./ChartHeader";
import ChartReadme from "./ChartReadme";

interface IRouteParams {
  cluster: string;
  namespace: string;
  repo: string;
  global: string;
  pluginName: string;
  pluginVersion: string;
  id: string;
  version?: string;
}

export default function ChartView() {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const {
    cluster,
    namespace,
    repo,
    global,
    pluginName,
    pluginVersion,
    id,
    version: queryVersion,
  } = ReactRouter.useParams() as IRouteParams;
  const {
    config,
    charts: { isFetching, selected },
  } = useSelector((state: IStoreState) => state);
  const { availablePackageDetail, versions, pkgVersion, readmeError, error, readme } = selected;

  const packageId = `${repo}/${id}`;
  const chartNamespace = global === "global" ? config.kubeappsNamespace : namespace;
  const chartCluster = global === "global" ? config.kubeappsCluster : cluster;
  const kubeappsNamespace = config.kubeappsNamespace;

  const location = ReactRouter.useLocation();

  const [pluginObj] = useState({ name: pluginName, version: pluginVersion } as Plugin);

  // Fetch the selected/latest version on the initial load
  useEffect(() => {
    dispatch(
      actions.charts.fetchChartVersion(
        {
          context: { cluster: chartCluster, namespace: chartNamespace },
          plugin: pluginObj,
          identifier: packageId,
        } as AvailablePackageReference,
        queryVersion,
      ),
    );
    return;
  }, [dispatch, packageId, chartNamespace, chartCluster, queryVersion, pluginObj]);

  // Fetch all versions
  useEffect(() => {
    dispatch(
      actions.charts.fetchChartVersions({
        context: { cluster: chartCluster, namespace: chartNamespace },
        plugin: { name: pluginName, version: pluginVersion } as Plugin,
        identifier: packageId,
      } as AvailablePackageReference),
    );
  }, [dispatch, packageId, chartNamespace, chartCluster, pluginName, pluginVersion]);

  // Select version handler
  const selectVersion = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const versionRegex = /\/versions\/(.*)/;
    if (versionRegex.test(location.pathname)) {
      // If the current URL already has the version, replace it
      dispatch(push(location.pathname.replace(versionRegex, `/versions/${event.target.value}`)));
    } else {
      // Otherwise, append the version
      dispatch(push(location.pathname.concat(`/versions/${event.target.value}`)));
    }
  };

  if (error) {
    return <Alert theme="danger">Unable to fetch package: {error.message}</Alert>;
  }
  if (isFetching || !availablePackageDetail) {
    return <LoadingWrapper loaded={false} />;
  }

  // If the chart does not have a README, redirect to deployment form
  if (!readme && config.skipAvailablePackageDetails) {
    return (
      <ReactRouter.Redirect
        to={app.apps.new(
          cluster,
          namespace,
          availablePackageDetail,
          pkgVersion!,
          kubeappsNamespace,
          pluginObj,
        )}
      />
    );
  }

  return (
    <section>
      <div>
        <ChartHeader
          chartAttrs={availablePackageDetail}
          versions={versions}
          onSelect={selectVersion}
          deployButton={
            <Link
              to={app.apps.new(
                cluster,
                namespace,
                availablePackageDetail,
                pkgVersion!,
                kubeappsNamespace,
                pluginObj,
              )}
            >
              <CdsButton status="primary">
                <CdsIcon shape="deploy" /> Deploy
              </CdsButton>
            </Link>
          }
          selectedVersion={pkgVersion}
        />
      </div>

      <section>
        <Row>
          <Column span={3}>
            <AvailablePackageDetailExcerpt pkg={availablePackageDetail} />
          </Column>
          <Column span={9}>
            <ChartReadme readme={readme} error={readmeError} />
            <div className="after-readme-button">
              <Link
                to={app.apps.new(
                  cluster,
                  namespace,
                  availablePackageDetail,
                  pkgVersion!,
                  kubeappsNamespace,
                  pluginObj,
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
