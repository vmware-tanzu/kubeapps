import * as React from "react";
import { AlertTriangle } from "react-feather";

import "./ErrorAlertHeader.css";

interface IErrorHeaderProps {
  children: string | JSX.Element | Array<string | JSX.Element>;
  icon?: React.Component;
}

class ErrorPageHeader extends React.Component<IErrorHeaderProps> {
  public render() {
    const { children } = this.props;
    const Icon = this.props.icon || AlertTriangle;
    return (
      <header>
        <div className="margin-b-big">
          <h5 className="type-regular">
            <span className="error__icon margin-r-small">
              <Icon />
            </span>
            {children}
          </h5>
        </div>
      </header>
    );
  }
}

export default ErrorPageHeader;
