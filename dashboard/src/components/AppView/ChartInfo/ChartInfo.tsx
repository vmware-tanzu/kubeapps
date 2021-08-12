import { InstalledPackageDetail } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import ChartUpdateInfo from "./ChartUpdateInfo";

interface IChartInfoProps {
  app: InstalledPackageDetail;
  cluster: string;
}

function ChartInfo({ app, cluster }: IChartInfoProps) {
  return (
    <section className="left-menu">
      {app && (
        <section className="left-menu-subsection" aria-labelledby="chartinfo-versions">
          <h5 className="left-menu-subsection-title" id="chartinfo-versions">
            Versions
          </h5>
          <div>
            {/* TODO(agamez): use currentAppVersion when available */}
            {/* {app.currentAppVersion && (
              <div>
                App Version: <strong>{app.currentAppVersion}</strong>
              </div>
            )} */}
            <span>
              Package Version: <strong>{app.currentPkgVersion}</strong>
            </span>
          </div>
          <ChartUpdateInfo app={app} cluster={cluster} />
        </section>
      )}
      {/* TODO(agamez): use shortDescription when available */}
      {/* {app?.shortDescription(
        <section className="left-menu-subsection" aria-labelledby="chartinfo-description">
          <h5 className="left-menu-subsection-title" id="chartinfo-description">
            Description
          </h5>
          <span>{app.shortDescription}</span>
        </section>
      )} */}
    </section>
  );
}

export default ChartInfo;
