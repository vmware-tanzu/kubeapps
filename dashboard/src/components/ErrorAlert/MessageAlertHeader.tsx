import * as React from "react";
import { Info } from "react-feather";

import "./MessageAlertHeader.css";

interface IMessageHeaderProps {
  children: string | JSX.Element | Array<string | JSX.Element>;
  icon?: React.Component;
}

class MessagePageHeader extends React.Component<IMessageHeaderProps> {
  public render() {
    const { children } = this.props;
    const Icon = this.props.icon || Info;
    return (
      <header>
        <div className="margin-b-big">
          <h5 className="type-regular">
            <span className="message__icon margin-r-normal">
              <Icon />
            </span>
            {children}
          </h5>
        </div>
      </header>
    );
  }
}

export default MessagePageHeader;
