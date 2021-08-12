import Tooltip from "components/js/Tooltip";
import { InstalledPackageSummary } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import * as semver from "semver";
import helmIcon from "../../icons/helm.svg";
import placeholder from "../../placeholder.png";
import * as url from "../../shared/url";
import InfoCard from "../InfoCard/InfoCard";
import "./AppListItem.css";

export interface IAppListItemProps {
  app: InstalledPackageSummary;
  cluster: string;
}

function AppListItem(props: IAppListItemProps) {
  const { app, cluster } = props;
  const icon = app.iconUrl ?? placeholder;
  const appStatus = app.status?.userReason?.toLocaleLowerCase();
  let tooltip = <></>;
  // TODO(agamez): API currently checks for pkg version updates, not app.version
  // TODO(agamez): do we want to display all possible updates (just app, just pkg or both) or just one one?
  if (
    app.latestMatchingPkgVersion &&
    semver.gt(app.latestMatchingPkgVersion, app.currentPkgVersion)
  ) {
    tooltip = (
      <div className="color-icon-info">
        <Tooltip
          label="update-tooltip"
          id={`${app.name}-update-tooltip`}
          icon="circle-arrow"
          position="top-left"
          iconProps={{ solid: true, size: "md", color: "blue" }}
        >
          A new matching package version is available:{" "}
          <strong>{app.latestMatchingPkgVersion}</strong>{" "}
          <em>(now using {app.currentPkgVersion})</em>
        </Tooltip>
      </div>
    );
  } else if (app.latestPkgVersion && semver.gt(app.latestPkgVersion, app.currentPkgVersion)) {
    tooltip = (
      <div className="color-icon-info">
        <Tooltip
          label="update-tooltip"
          id={`${app.name}-update-tooltip`}
          icon="circle-arrow"
          position="top-left"
          iconProps={{ solid: true, size: "md" }}
        >
          A new package version is available: <strong>{app.latestPkgVersion}</strong>{" "}
          <em>(now using {app.currentPkgVersion})</em>
        </Tooltip>
      </div>
    );
  }
  return (
    <InfoCard
      key={app.installedPackageRef?.identifier}
      link={url.app.apps.get(cluster, app.installedPackageRef?.context?.namespace || "", app.name)}
      title={app.name}
      icon={icon}
      info={
        <div>
          <span>
            App: {app.pkgDisplayName}{" "}
            {app.currentAppVersion ? `v${app.currentAppVersion.replace(/^v/, "")}` : ""}
          </span>
          <br />
          <span>Package: {app.currentPkgVersion}</span>
        </div>
      }
      description={app.shortDescription}
      tag1Content={appStatus}
      tag1Class={appStatus === "deployed" ? "label-success" : "label-warning"}
      tooltip={tooltip}
      bgIcon={helmIcon}
    />
  );
}

export default AppListItem;
