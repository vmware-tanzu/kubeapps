import * as React from "react";

import Column from "components/js/Column";
import Row from "components/js/Row";
import olmIcon from "icons/operator-framework.svg";
import helmIcon from "../../icons/helm.svg";
import "./PageHeader.v2.css";

export interface IPageHeaderProps {
  title: string;
  titleSize?: "lg" | "md";
  icon?: any;
  filter?: JSX.Element;
  helm?: boolean;
  operator?: boolean;
  buttons?: JSX.Element[];
  version?: JSX.Element;
}
function PageHeader({
  title,
  titleSize = "lg",
  icon,
  filter,
  buttons,
  helm,
  version,
  operator,
}: IPageHeaderProps) {
  return (
    <header className="kubeapps-header">
      <div className="kubeapps-header-content">
        <Row>
          <Column span={7}>
            <div className="kubeapps-title-section">
              {icon && <img src={icon} alt="app-icon" />}
              <div className="kubeapps-title-block">
                {titleSize === "lg" ? <h1>{title}</h1> : <h3>{title}</h3>}
                {helm && (
                  <div className="kubeapps-header-subtitle">
                    <img src={helmIcon} alt="helm-icon" />
                    <span>Helm Chart</span>
                  </div>
                )}
                {operator && (
                  <div className="kubeapps-header-subtitle">
                    <img src={olmIcon} alt="olm-icon" />
                    <span>Operator</span>
                  </div>
                )}
              </div>
              {filter}
            </div>
          </Column>
          <Column span={5}>
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
