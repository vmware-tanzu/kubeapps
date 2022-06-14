// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  AvailablePackageDetail,
  InstalledPackageDetail,
  InstalledPackageStatus_StatusReason,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import PackageUpdateInfo from "./PackageUpdateInfo";
import { getAppStatusLabel } from "shared/utils";

interface IPackageInfoProps {
  installedPackageDetail: InstalledPackageDetail;
  availablePackageDetail?: AvailablePackageDetail;
}

function PackageInfo({ installedPackageDetail, availablePackageDetail }: IPackageInfoProps) {
  return (
    <section className="left-menu">
      {installedPackageDetail && (
        <>
          {installedPackageDetail.status && (
            <section className="left-menu-subsection" aria-labelledby="packageinfo-versions">
              <div>
                Status: <strong>{getAppStatusLabel(installedPackageDetail.status.reason)}</strong>
              </div>
              {installedPackageDetail.status.reason !==
                InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED && (
                <div>{installedPackageDetail.status.userReason}</div>
              )}
            </section>
          )}
          <section className="left-menu-subsection" aria-labelledby="packageinfo-versions">
            <h5 className="left-menu-subsection-title" id="packageinfo-versions">
              Versions
            </h5>
            <div>
              {installedPackageDetail.currentVersion?.appVersion && (
                <div>
                  App Version: <strong>{installedPackageDetail.currentVersion?.appVersion}</strong>
                </div>
              )}
              <span>
                Package Version:{" "}
                <strong>{installedPackageDetail.currentVersion?.pkgVersion}</strong>
              </span>
            </div>
            <PackageUpdateInfo installedPackageDetail={installedPackageDetail} />
          </section>
          {installedPackageDetail.reconciliationOptions && (
            <section className="left-menu-subsection" aria-labelledby="packageinfo-reconciliation">
              <h5 className="left-menu-subsection-title" id="packageinfo-reconciliation">
                Reconciliation Options
              </h5>
              <div>
                <>
                  {" "}
                  <div>
                    Service Account:{" "}
                    <strong>
                      {installedPackageDetail.reconciliationOptions.serviceAccountName}
                    </strong>
                  </div>
                  <div>
                    Interval:{" "}
                    <strong>{installedPackageDetail.reconciliationOptions.interval}</strong>
                  </div>
                </>
              </div>
            </section>
          )}
        </>
      )}
      {availablePackageDetail && (
        <>
          <section className="left-menu-subsection" aria-labelledby="packageinfo-description">
            <h5 className="left-menu-subsection-title" id="packageinfo-description">
              Description
            </h5>
            <span>{availablePackageDetail.shortDescription}</span>
          </section>
        </>
      )}
    </section>
  );
}

export default PackageInfo;
