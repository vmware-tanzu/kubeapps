import React from "react";
import { IChartAttributes, IChartVersion } from "shared/types";

interface IChartHeaderProps {
  chartAttrs: IChartAttributes;
  versions: IChartVersion[];
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
        value={
          selectedVersion ||
          currentVersion ||
          (versions.length ? versions[0].attributes.version : "")
        }
      >
        {versions.map(v => {
          return (
            <option
              key={`chart-version-selector-${v.attributes.version}`}
              value={v.attributes.version}
            >
              {v.attributes.version} / App Version {v.attributes.app_version}
              {currentVersion === v.attributes.version ? " (current)" : ""}
            </option>
          );
        })}
      </select>
    </div>
  );
}
