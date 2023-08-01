// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import InfoCard from "components/InfoCard";
import { InstalledPackageSummary } from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import placeholder from "icons/placeholder.svg";
import { Tooltip } from "react-tooltip";
import * as url from "shared/url";
import { getAppStatusLabel, getPluginIcon, getPluginName } from "shared/utils";
import "./AppListItem.css";

export interface IAppListItemProps {
  app: InstalledPackageSummary;
  cluster: string;
}

function AppListItem(props: IAppListItemProps) {
  const { app } = props;
  const icon = app.iconUrl ?? placeholder;
  const appStatus = getAppStatusLabel(app.status?.reason);
  const appReady = app.status?.ready ?? false;
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
    app.latestVersion?.pkgVersion !== app.currentVersion?.pkgVersion
  ) {
    tooltipContent = (
      <>
        A new package version is available: <strong>{app.latestVersion?.pkgVersion}</strong>
      </>
    );
  }

  const tooltip = tooltipContent ? (
    <div className="color-icon-info">
      <span data-tooltip-id={`${app.name}-update-tooltip`}>
        <CdsIcon shape="circle-arrow" size="md" solid={true} />
      </span>
      <Tooltip id={`${app.name}-update-tooltip`} place="top-end" className="small-tooltip">
        {tooltipContent}
      </Tooltip>
    </div>
  ) : (
    <></>
  );

  const pkgPluginName = getPluginName(app.installedPackageRef?.plugin);

  return app?.installedPackageRef ? (
    <InfoCard
      key={app.installedPackageRef?.identifier}
      link={url.app.apps.get(app.installedPackageRef)}
      title={app.name}
      icon={icon}
      info={
        <div>
          <span>Namespace: {app.installedPackageRef.context?.namespace}</span>
          <br />
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
      tag1Class={appReady ? "label-success" : "label-warning"}
      tag2Content={pkgPluginName}
      tag2Class={"label-info-secondary"}
      tooltip={tooltip}
      bgIcon={getPluginIcon(app.installedPackageRef?.plugin)}
    />
  ) : (
    <></>
  );
}

export default AppListItem;
