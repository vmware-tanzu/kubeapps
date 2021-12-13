import {
  AvailablePackageDetail,
  InstalledPackageDetail,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import PackageUpdateInfo from "./PackageUpdateInfo";
interface IPackageInfoProps {
  installedPackageDetail: InstalledPackageDetail;
  availablePackageDetail?: AvailablePackageDetail;
}

function PackageInfo({ installedPackageDetail, availablePackageDetail }: IPackageInfoProps) {
  return (
    <section className="left-menu">
      {installedPackageDetail && (
        <>
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
                    <strong>{installedPackageDetail.reconciliationOptions.interval} seconds</strong>
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
