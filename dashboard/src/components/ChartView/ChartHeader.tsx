import Tooltip from "components/js/Tooltip";
import PageHeader from "components/PageHeader/PageHeader";
import {
  AvailablePackageDetail,
  GetAvailablePackageVersionsResponse_PackageAppVersion,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import React from "react";
import placeholder from "../../placeholder.png";
import "./ChartHeader.css";
import ChartVersionSelector from "./ChartVersionSelector";

interface IChartHeaderProps {
  chartAttrs: AvailablePackageDetail;
  versions: GetAvailablePackageVersionsResponse_PackageAppVersion[];
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
        // TODO(agamez): get the repo name once available
        // https://github.com/kubeapps/kubeapps/issues/3165#issuecomment-884574732
        releaseName
          ? `${releaseName} (${
              chartAttrs.availablePackageRef?.identifier.split("/")[0]
            }/${decodeURIComponent(chartAttrs.name)})`
          : `${chartAttrs.availablePackageRef?.identifier.split("/")[0]}/${decodeURIComponent(
              chartAttrs.name,
            )}`
      }
      titleSize="md"
      icon={chartAttrs.iconUrl ? chartAttrs.iconUrl : placeholder}
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
