import { CdsIcon } from "components/Clarity/clarity";
import * as React from "react";

import { hapi } from "shared/hapi/release";
import { IChartUpdateInfo, IRelease } from "shared/types";

interface IChartInfoProps {
  app: IRelease;
}

function getUpdateInfo(updateInfo: IChartUpdateInfo, chartMetadata: hapi.chart.IMetadata) {
  if (updateInfo.error) {
    // Unable to get info, return error
    return (
      <div className="color-icon-danger">
        <CdsIcon shape="exclamation-triange" size="md" solid={true} /> Update check failed.{" "}
        {updateInfo.error.message}
      </div>
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
      <div className="color-icon-info">
        <CdsIcon shape="circle-arrow" size="md" solid={true} /> A new version for{" "}
        {chartMetadata.name} is available: {updateInfo.appLatestVersion}.
      </div>
    );
  }
  // There is a new chart version
  return (
    <div className="color-icon-info">
      <CdsIcon shape="circle-arrow" size="md" solid={true} /> A new chart version is available:{" "}
      {updateInfo.chartLatestVersion}.
    </div>
  );
}

export default function ChartUpdateInfo(props: IChartInfoProps) {
  const { app } = props;
  let updateInfo = null;
  // If update is not set yet we cannot know if there is
  // an update available or not
  if (app.updateInfo && app.chart?.metadata) {
    updateInfo = (
      <section className="left-menu-subsection" aria-labelledby="chartinfo-update-info">
        <h5 className="left-menu-subsection-title" id="chartinfo-update-info">
          Update Info
        </h5>
        {getUpdateInfo(app.updateInfo, app.chart.metadata)}
      </section>
    );
  }
  return updateInfo;
}
