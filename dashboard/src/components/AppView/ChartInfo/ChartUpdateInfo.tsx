import { CdsIcon } from "@cds/react/icon";

import Alert from "components/js/Alert";
import { Link } from "react-router-dom";
import { hapi } from "shared/hapi/release";
import { IChartUpdateInfo, IRelease } from "shared/types";
import { app as appURL } from "shared/url";

interface IChartInfoProps {
  cluster: string;
  app: IRelease;
}

function getUpdateInfo(
  name: string,
  namespace: string,
  cluster: string,
  updateInfo: IChartUpdateInfo,
  chartMetadata: hapi.chart.IMetadata,
) {
  if (updateInfo.error) {
    // Unable to get info, return error
    return (
      <Alert theme="danger">
        <CdsIcon shape="exclamation-triange" size="md" solid={true} /> Update check failed.{" "}
        {updateInfo.error.message}
      </Alert>
    );
  }
  if (updateInfo.upToDate) {
    // App is up to date
    return (
      <div className="color-icon-success">
        <CdsIcon shape="check-circle" size="md" solid={true} /> Up to date
      </div>
    );
  }
  if (chartMetadata.appVersion !== updateInfo.appLatestVersion) {
    // There is a new application version
    return (
      <Alert>
        A new app version is available: <strong>{updateInfo.appLatestVersion}</strong>.{" "}
        <Link to={appURL.apps.upgrade(cluster, namespace, name)}>Update Now</Link>
      </Alert>
    );
  }
  // There is a new chart version
  return (
    <Alert>
      A new chart version is available: <strong>{updateInfo.chartLatestVersion}</strong>.{" "}
      <Link to={appURL.apps.upgrade(cluster, namespace, name)}>Update Now</Link>
    </Alert>
  );
}

export default function ChartUpdateInfo({ app, cluster }: IChartInfoProps) {
  let updateInfo = null;
  // If update is not set yet we cannot know if there is
  // an update available or not
  if (app.updateInfo && app.chart?.metadata) {
    updateInfo = getUpdateInfo(
      app.name,
      app.namespace,
      cluster,
      app.updateInfo,
      app.chart.metadata,
    );
  }
  return updateInfo;
}
