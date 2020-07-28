import * as React from "react";
import { IRelease } from "shared/types";
import "./ChartInfo.v2.css";
import ChartUpdateInfo from "./ChartUpdateInfo";

interface IChartInfoProps {
  app: IRelease;
}

function ChartInfo({ app }: IChartInfoProps) {
  const metadata = app.chart && app.chart.metadata;
  return (
    <section className="left-menu">
      {metadata && (
        <section className="left-menu-subsection" aria-labelledby="chartinfo-versions">
          <h5 className="left-menu-subsection-title" id="chartinfo-versions">
            Versions
          </h5>
          <div>
            {metadata.appVersion && (
              <div>
                App Version: <strong>{metadata.appVersion}</strong>
              </div>
            )}
            <span>
              Chart Version: <strong>{metadata.version}</strong>
            </span>
          </div>
        </section>
      )}
      {metadata && (
        <section className="left-menu-subsection" aria-labelledby="chartinfo-description">
          <h5 className="left-menu-subsection-title" id="chartinfo-description">
            Description
          </h5>
          <span>{metadata.description}</span>
        </section>
      )}
      <ChartUpdateInfo app={app} />
    </section>
  );
}

export default ChartInfo;
