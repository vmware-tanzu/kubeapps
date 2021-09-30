import { PackageAppVersion } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import React from "react";

interface IPackageVersionSelectorProps {
  versions: PackageAppVersion[];
  onSelect: (event: React.ChangeEvent<HTMLSelectElement>) => void;
  currentVersion?: string;
  selectedVersion?: string;
}

export default function PackageVersionSelector({
  versions,
  onSelect,
  currentVersion,
  selectedVersion,
}: IPackageVersionSelectorProps) {
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
