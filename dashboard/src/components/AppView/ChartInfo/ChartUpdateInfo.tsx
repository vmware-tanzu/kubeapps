import { CdsIcon } from "@cds/react/icon";
import Alert from "components/js/Alert";
import { InstalledPackageDetail } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { Link } from "react-router-dom";
import { app as appURL } from "shared/url";
interface IChartInfoProps {
  cluster: string;
  app: InstalledPackageDetail;
  plugin: Plugin;
}

export default function ChartUpdateInfo({ app, cluster, plugin }: IChartInfoProps) {
  const namespace = app.installedPackageRef?.context?.namespace || "";
  let alertContent;
  if (
    app.latestVersion?.appVersion &&
    app.currentVersion?.appVersion &&
    app.currentVersion?.appVersion !== app.latestVersion?.appVersion
  ) {
    // There is a new application version
    alertContent = (
      <>
        A new app version is available: <strong>{app.latestVersion?.appVersion}</strong>.{" "}
      </>
    );
  } else if (
    app.latestVersion?.pkgVersion &&
    app.currentVersion?.pkgVersion &&
    app.latestVersion?.pkgVersion !== app.currentVersion?.pkgVersion
  ) {
    // There is a new package version
    alertContent = (
      <>
        A new package version is available: <strong>{app.latestVersion?.pkgVersion}</strong>.{" "}
      </>
    );
  }
  // App is up to date
  return alertContent ? (
    <Alert>
      {alertContent}
      <Link to={appURL.apps.upgrade(cluster, namespace, app.name, plugin)}>Update Now</Link>
    </Alert>
  ) : (
    <div className="color-icon-success">
      <CdsIcon shape="check-circle" size="md" solid={true} /> Up to date
    </div>
  );
}
