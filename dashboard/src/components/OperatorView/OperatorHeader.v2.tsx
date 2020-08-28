import Column from "components/js/Column";
import Row from "components/js/Row";
import PageHeader from "components/PageHeader/PageHeader.v2";
import olmIcon from "icons/operator-framework.svg";
import * as React from "react";
import placeholder from "../../placeholder.png";

interface IOperatorHeaderProps {
  id: string;
  version: string;
  provider: string;
  icon?: string;
  children?: any;
}

export default function OperatorHeader(props: IOperatorHeaderProps) {
  const { id, icon, version, provider, children } = props;
  return (
    <PageHeader>
      <div className="kubeapps-header-content">
        <Row>
          <Column span={7}>
            <Row>
              <img src={icon || placeholder} alt="app-icon" />
              <div className="kubeapps-title-block">
                <h3>
                  {id} by {provider}
                </h3>
                <div className="kubeapps-header-subtitle">
                  <img src={olmIcon} alt="olm-icon" />
                  <span>Operator</span>
                </div>
              </div>
            </Row>
          </Column>
          <Column span={5}>
            <div className="control-buttons">
              <div className="header-version">
                <label className="header-version-label">Operator Version: {version}</label>
              </div>
              {children}
            </div>
          </Column>
        </Row>
      </div>
    </PageHeader>
  );
}
