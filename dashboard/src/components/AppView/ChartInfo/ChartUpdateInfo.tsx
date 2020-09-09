import { CdsIcon } from "@clr/react/icon";
import * as React from "react";

import Alert from "components/js/Alert";
import { hapi } from "shared/hapi/release";
import { IChartUpdateInfo, IRelease } from "shared/types";

interface IChartInfoProps {
  app: IRelease;
}

function getUpdateInfo(updateInfo: IChartUpdateInfo, chartMetadata: hapi.chart.IMetadata) {
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
        A new app version is available: <strong>{updateInfo.appLatestVersion}</strong>.
      </Alert>
    );
  }
  // There is a new chart version
  return (
    <Alert>
      A new chart version is available: <strong>{updateInfo.chartLatestVersion}</strong>.
    </Alert>
  );
}

export default function ChartUpdateInfo(props: IChartInfoProps) {
  const { app } = props;
  let updateInfo = null;
  // If update is not set yet we cannot know if there is
  // an update available or not
  if (app.updateInfo && app.chart?.metadata) {
    updateInfo = getUpdateInfo(app.updateInfo, app.chart.metadata);
  }
  return updateInfo;
}
