import {
  AvailablePackageDetail,
  GetAvailablePackageVersionsResponse_PackageAppVersion,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import React from "react";

interface IChartHeaderProps {
  chartAttrs: AvailablePackageDetail;
  versions: GetAvailablePackageVersionsResponse_PackageAppVersion[];
  onSelect: (event: React.ChangeEvent<HTMLSelectElement>) => void;
  releaseName?: string;
  currentVersion?: string;
  selectedVersion?: string;
  deployButton?: JSX.Element;
}

export default function ChartVersionSelector({
  versions,
  onSelect,
  currentVersion,
  selectedVersion,
}: IChartHeaderProps) {
  return (
    <div className="clr-select-wrapper">
      <select
        name="chart-versions"
        className="clr-page-size-select"
        onChange={onSelect}
        value={selectedVersion || currentVersion || (versions.length ? versions[0].pkgVersion : "")}
      >
        {versions.map(v => {
          return (
            <option key={`chart-version-selector-${v.pkgVersion}`} value={v.pkgVersion}>
              {v.pkgVersion} / App Version {v.appVersion}
              {currentVersion === v.pkgVersion ? " (current)" : ""}
            </option>
          );
        })}
      </select>
    </div>
  );
}
