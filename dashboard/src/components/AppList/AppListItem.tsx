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
  let tooltipContent;

  if (
    app.latestVersion?.appVersion &&
    app.currentVersion?.appVersion &&
    app.currentVersion?.appVersion !== app.latestVersion?.appVersion
  ) {
    tooltipContent = (
      <>
        A new app version is available: <strong>{app.latestVersion?.appVersion}</strong>
      </>
    );
  } else if (
    app.latestVersion?.pkgVersion &&
    app.currentVersion?.pkgVersion &&
    semver.gt(app.latestVersion?.pkgVersion, app.currentVersion?.pkgVersion)
  ) {
    tooltipContent = (
      <>
        A new package version is available: <strong>{app.latestVersion?.pkgVersion}</strong>
      </>
    );
  }

  const tooltip = tooltipContent ? (
    <div className="color-icon-info">
      <Tooltip
        label="update-tooltip"
        id={`${app.name}-update-tooltip`}
        icon="circle-arrow"
        position="top-left"
        iconProps={{ solid: true, size: "md" }}
      >
        {tooltipContent}
      </Tooltip>
    </div>
  ) : (
    <></>
  );

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
            {app.currentVersion?.appVersion
              ? `v${app.currentVersion.appVersion.replace(/^v/, "")}`
              : ""}
          </span>
          <br />
          <span>Package: {app.currentVersion?.pkgVersion}</span>
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
