import Column from "components/js/Column";
import Row from "components/js/Row";
import Tooltip from "components/js/Tooltip";
import PageHeader from "components/PageHeader/PageHeader.v2";
import React from "react";
import { IChartAttributes, IChartVersion } from "shared/types";
import helmIcon from "../../icons/helm.svg";
import placeholder from "../../placeholder.png";

import "./ChartHeader.v2.css";

interface IChartHeaderProps {
  chartAttrs: IChartAttributes;
  versions: IChartVersion[];
  onSelect: (event: React.ChangeEvent<HTMLSelectElement>) => void;
  releaseName?: string;
  currentVersion?: string;
  children?: JSX.Element;
}

export default function ChartHeader({
  chartAttrs,
  versions,
  onSelect,
  releaseName,
  currentVersion,
  children,
}: IChartHeaderProps) {
  return (
    <PageHeader>
      <div className="kubeapps-header-content">
        <Row>
          <Column span={7}>
            <Row>
              <img
                src={chartAttrs.icon ? `api/assetsvc/${chartAttrs.icon}` : placeholder}
                alt="app-icon"
              />
              <div className="kubeapps-title-block">
                <h3>
                  {releaseName
                    ? `${releaseName} (${chartAttrs.repo.name}/${chartAttrs.name})`
                    : `${chartAttrs.repo.name}/${chartAttrs.name}`}
                </h3>
                <div className="kubeapps-header-subtitle">
                  <img src={helmIcon} alt="helm-icon" />
                  <span>Helm Chart</span>
                </div>
              </div>
            </Row>
          </Column>
          <Column span={5}>
            <div className="control-buttons">
              <div className="header-version">
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
                <div className="clr-select-wrapper">
                  <select
                    name="chart-versions"
                    className="clr-page-size-select"
                    onChange={onSelect}
                    defaultValue={
                      currentVersion ||
                      (versions.length ? versions[0].attributes.version : undefined)
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
                {children}
              </div>
            </div>
          </Column>
        </Row>
      </div>
    </PageHeader>
  );
}
