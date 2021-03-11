import { IRelease } from "shared/types";
import "./ChartInfo.css";
import ChartUpdateInfo from "./ChartUpdateInfo";

interface IChartInfoProps {
  app: IRelease;
  cluster: string;
}

function ChartInfo({ app, cluster }: IChartInfoProps) {
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
          <ChartUpdateInfo app={app} cluster={cluster} />
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
    </section>
  );
}

export default ChartInfo;
