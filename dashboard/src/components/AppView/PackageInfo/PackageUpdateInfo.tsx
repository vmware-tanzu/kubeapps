// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import AlertGroup from "components/AlertGroup";
import { InstalledPackageDetail } from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import { Link } from "react-router-dom";
import { app as appURL } from "shared/url";
interface IPackageUpdateInfoProps {
  installedPackageDetail: InstalledPackageDetail;
}

export default function PackageUpdateInfo({ installedPackageDetail }: IPackageUpdateInfoProps) {
  let alertContent;
  if (
    installedPackageDetail.latestVersion?.appVersion &&
    installedPackageDetail.currentVersion?.appVersion &&
    installedPackageDetail.currentVersion?.appVersion !==
      installedPackageDetail.latestVersion?.appVersion
  ) {
    // There is a new application version
    alertContent = (
      <>
        A new app version is available:{" "}
        <strong>{installedPackageDetail.latestVersion?.appVersion}</strong>.{" "}
      </>
    );
  } else if (
    installedPackageDetail.latestVersion?.pkgVersion &&
    installedPackageDetail.currentVersion?.pkgVersion &&
    installedPackageDetail.latestVersion?.pkgVersion !==
      installedPackageDetail.currentVersion?.pkgVersion
  ) {
    // There is a new package version
    alertContent = (
      <>
        A new package version is available:{" "}
        <strong>{installedPackageDetail.latestVersion?.pkgVersion}</strong>.{" "}
      </>
    );
  }
  // App is up to date
  return alertContent && installedPackageDetail?.installedPackageRef ? (
    <AlertGroup status="info" closable={false}>
      {alertContent}
      <br />
      <Link
        to={appURL.apps.upgradeTo(
          installedPackageDetail.installedPackageRef,
          installedPackageDetail.latestVersion?.pkgVersion,
        )}
      >
        Update Now
      </Link>
    </AlertGroup>
  ) : (
    <div className="color-icon-success">
      <CdsIcon shape="check-circle" size="md" solid={true} /> Up to date
    </div>
  );
}
