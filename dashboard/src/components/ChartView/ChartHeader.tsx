import Tooltip from "components/js/Tooltip";
import PageHeader from "components/PageHeader/PageHeader";
import React from "react";
import { IChartAttributes, IChartVersion } from "shared/types";
import placeholder from "../../placeholder.png";
import ChartVersionSelector from "./ChartVersionSelector";

import "./ChartHeader.css";

interface IChartHeaderProps {
  chartAttrs: IChartAttributes;
  versions: IChartVersion[];
  onSelect: (event: React.ChangeEvent<HTMLSelectElement>) => void;
  releaseName?: string;
  currentVersion?: string;
  selectedVersion?: string;
  deployButton?: JSX.Element;
}

export default function ChartHeader({
  chartAttrs,
  versions,
  onSelect,
  releaseName,
  currentVersion,
  deployButton,
  selectedVersion,
}: IChartHeaderProps) {
  let icon = placeholder;
  if (chartAttrs?.icon) {
    icon = chartAttrs.icon.startsWith("/") ? chartAttrs.icon : "/".concat(chartAttrs.icon);
    icon = "api/assetsvc".concat(icon);
  }
  return (
    <PageHeader
      title={
        releaseName
          ? `${releaseName} (${chartAttrs.repo.name}/${chartAttrs.name})`
          : `${chartAttrs.repo.name}/${chartAttrs.name}`
      }
      titleSize="md"
      icon={icon}
      helm={true}
      version={
        <>
          <label className="header-version-label" htmlFor="chart-versions">
            Chart Version{" "}
            <Tooltip
              label="chart-versions-tooltip"
              id="chart-versions-tooltip"
              position="bottom-left"
              iconProps={{ solid: true, size: "sm" }}
            >
              Chart and App versions can be increased independently.{" "}
              <a
                href="https://helm.sh/docs/topics/charts/#charts-and-versioning"
                target="_blank"
                rel="noopener noreferrer"
              >
                More info here
              </a>
              .{" "}
            </Tooltip>
          </label>
          <ChartVersionSelector
            versions={versions}
            onSelect={onSelect}
            selectedVersion={selectedVersion}
            currentVersion={currentVersion}
            chartAttrs={chartAttrs}
          />
        </>
      }
      buttons={deployButton ? [deployButton] : undefined}
    />
  );
}
