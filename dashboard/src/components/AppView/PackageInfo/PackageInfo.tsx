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
        <section className="left-menu-subsection" aria-labelledby="chartinfo-versions">
          <h5 className="left-menu-subsection-title" id="chartinfo-versions">
            Versions
          </h5>
          <div>
            {installedPackageDetail.currentVersion?.appVersion && (
              <div>
                App Version: <strong>{installedPackageDetail.currentVersion?.appVersion}</strong>
              </div>
            )}
            <span>
              Package Version: <strong>{installedPackageDetail.currentVersion?.pkgVersion}</strong>
            </span>
          </div>
          <PackageUpdateInfo installedPackageDetail={installedPackageDetail} />
        </section>
      )}
      {availablePackageDetail && (
        <>
          <section className="left-menu-subsection" aria-labelledby="chartinfo-description">
            <h5 className="left-menu-subsection-title" id="chartinfo-description">
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
