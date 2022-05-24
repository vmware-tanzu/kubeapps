// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Icon from "components/Icon/Icon";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { getPluginIcon, getPluginPackageName } from "shared/utils";
import "./PageHeader.css";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";

export interface IPageHeaderProps {
  title: string;
  titleSize?: "lg" | "md";
  icon?: any;
  subtitle?: JSX.Element;
  filter?: JSX.Element;
  plugin?: Plugin;
  operator?: boolean;
  buttons?: JSX.Element[];
  version?: JSX.Element;
}
function PageHeader({
  title,
  titleSize = "lg",
  icon,
  subtitle,
  filter,
  buttons,
  plugin,
  version,
  operator,
}: IPageHeaderProps) {
  return (
    <header className="kubeapps-header">
      <div className="kubeapps-header-content">
        <Row>
          <Column>
            <div className="kubeapps-title-section">
              <div className="img-container">{icon && <Icon icon={icon} />}</div>
              <div className="kubeapps-title-block">
                {titleSize === "lg" ? <h1>{title}</h1> : <h3>{title}</h3>}
                {subtitle && <div className="kubeapps-header-subtitle">{subtitle}</div>}
                {plugin && (
                  <div className="kubeapps-header-subtitle">
                    <img src={getPluginIcon(plugin)} alt="package-icon" />
                    <span>{getPluginPackageName(plugin)}</span>
                  </div>
                )}
                {operator && (
                  <div className="kubeapps-header-subtitle">
                    <img src={getPluginIcon("operator")} alt="olm-icon" />
                    <span>{getPluginPackageName("operator")}</span>
                  </div>
                )}
              </div>
              {filter}
            </div>
          </Column>
          <Column>
            <div className="control-buttons">
              {version && <div className="header-version">{version}</div>}
              {buttons ? (
                buttons.map((button, i) => (
                  <div className="header-button" key={`control-button-${i}`}>
                    {button}
                  </div>
                ))
              ) : (
                <></>
              )}
            </div>
          </Column>
        </Row>
      </div>
    </header>
  );
}

export default PageHeader;
