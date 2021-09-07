import {
  AvailablePackageDetail,
  InstalledPackageDetail,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import ChartUpdateInfo from "./ChartUpdateInfo";
interface IChartInfoProps {
  app: InstalledPackageDetail;
  appDetails?: AvailablePackageDetail;
  cluster: string;
  plugin: Plugin;
}

function ChartInfo({ app, appDetails, cluster, plugin }: IChartInfoProps) {
  return (
    <section className="left-menu">
      {app && (
        <section className="left-menu-subsection" aria-labelledby="chartinfo-versions">
          <h5 className="left-menu-subsection-title" id="chartinfo-versions">
            Versions
          </h5>
          <div>
            {app.currentVersion?.appVersion && (
              <div>
                App Version: <strong>{app.currentVersion?.appVersion}</strong>
              </div>
            )}
            <span>
              Package Version: <strong>{app.currentVersion?.pkgVersion}</strong>
            </span>
          </div>
          <ChartUpdateInfo app={app} cluster={cluster} plugin={plugin} />
        </section>
      )}
      {appDetails && (
        <>
          <section className="left-menu-subsection" aria-labelledby="chartinfo-description">
            <h5 className="left-menu-subsection-title" id="chartinfo-description">
              Description
            </h5>
            <span>{appDetails.shortDescription}</span>
          </section>
        </>
      )}
    </section>
  );
}

export default ChartInfo;
