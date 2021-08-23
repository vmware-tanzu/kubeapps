import Tooltip from "components/js/Tooltip";
import PageHeader from "components/PageHeader/PageHeader";
import { trimStart } from "lodash";
import React from "react";
import { IChartAttributes, IChartVersion } from "shared/types";
import placeholder from "../../placeholder.png";
import "./ChartHeader.css";
import ChartVersionSelector from "./ChartVersionSelector";

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
  return (
    <PageHeader
      title={
        releaseName
          ? `${releaseName} (${chartAttrs.repo.name}/${decodeURIComponent(chartAttrs.name)})`
          : `${chartAttrs.repo.name}/${decodeURIComponent(chartAttrs.name)}`
      }
      titleSize="md"
      icon={chartAttrs.icon ? `api/assetsvc/${trimStart(chartAttrs.icon, "/")}` : placeholder}
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
