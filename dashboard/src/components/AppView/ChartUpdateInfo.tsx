import {
  checkCircleIcon,
  circleArrowIcon,
  ClarityIcons,
  exclamationTriangleIcon,
} from "@clr/core/icon-shapes";
import * as React from "react";
import { CdsIcon } from "../Clarity/clarity";

import { IRelease } from "shared/types";

ClarityIcons.addIcons(exclamationTriangleIcon, checkCircleIcon, circleArrowIcon);

interface IChartInfoProps {
  app: IRelease;
}

export default function ChartUpdateInfo(props: IChartInfoProps) {
  const { app } = props;
  // If update is not set yet we cannot know if there is
  // an update available or not
  if (app.updateInfo) {
    let updateContent: JSX.Element | null = null;
    if (app.updateInfo.error) {
      updateContent = (
        <div className="color-icon-danger">
          <CdsIcon shape="exclamation-triange" size="md" solid={true} /> Update check failed.{" "}
          {app.updateInfo.error.message}
        </div>
      );
    } else {
      if (app.updateInfo.upToDate) {
        updateContent = (
          <div className="color-icon-success">
            <CdsIcon shape="check-circle" size="md" solid={true} /> Up to date
          </div>
        );
      } else {
        const update =
          app.chart &&
          app.chart.metadata &&
          app.chart.metadata.appVersion !== app.updateInfo.appLatestVersion ? (
            // A new version for the app is available
            <span>
              A new version for {app.chart.metadata.name} is available:{" "}
              {app.updateInfo.appLatestVersion}.
            </span>
          ) : (
            // Just a new chart version
            <span>A new chart version is available: {app.updateInfo.chartLatestVersion}.</span>
          );
        updateContent = (
          <div className="color-icon-info">
            <CdsIcon shape="circle-arrow" size="md" solid={true} /> {update}
          </div>
        );
      }
    }
    return (
      <section className="chartinfo-subsection" aria-labelledby="chartinfo-update-info">
        <h5 className="chartinfo-subsection-title" id="chartinfo-update-info">
          Update Info
        </h5>
        {updateContent}
      </section>
    );
  }
  return <div />;
}
