import { CdsIcon } from "@cds/react/icon";
// import Alert from "components/js/Alert";
import { InstalledPackageDetail } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
// import { Link } from "react-router-dom";
// import * as semver from "semver";
// import { app as appURL } from "shared/url";

interface IChartInfoProps {
  cluster: string;
  app: InstalledPackageDetail;
}

export default function ChartUpdateInfo({ app, cluster }: IChartInfoProps) {
  // const namespace = app.installedPackageRef?.context?.namespace || "";

  // TODO(agamez): add latestPkgVersion and latestAppVersion when available in installedPackageDetail
  // if (semver.gt(app.latestAppVersion, app.currentAppVersion)) {
  //   // There is a new application version
  //   return (
  //     <Alert>
  //       A new app version is available: <strong>{app.latestAppVersion}</strong>.{" "}
  //       <Link to={appURL.apps.upgrade(cluster, namespace, app.name)}>Update Now</Link>
  //     </Alert>
  //   );
  // }
  // if (semver.gt(app.latestPkgVersion, app.currentPkgVersion)) {
  //   // There is a new package version
  //   return (
  //     <Alert>
  //       A new package version is available: <strong>{app.latestPkgVersion}</strong>.{" "}
  //       <Link to={appURL.apps.upgrade(cluster, namespace, app.name)}>Update Now</Link>
  //     </Alert>
  //   );
  // }
  // App is up to date
  return (
    <div className="color-icon-success">
      <CdsIcon shape="check-circle" size="md" solid={true} /> Up to date
    </div>
  );
}
