import * as React from "react";
import { Info } from "react-feather";

import ErrorPageHeader from "./ErrorAlertHeader";

interface IMessageAlertPageProps {
  header?: string;
  level?: string;
  children?: JSX.Element;
}

class MessageAlertPage extends React.Component<IMessageAlertPageProps> {
  public render() {
    const { level, children, header } = this.props;
    return (
      <div className={`alert ${level ? `alert-${level}` : ""} margin-c margin-t-bigger`}>
        {header ? <ErrorPageHeader icon={Info}>{header}</ErrorPageHeader> : null}
        {children && (
          <div className={`message__content ${header ? "margin-l-enormous" : ""}`}>{children}</div>
        )}
      </div>
    );
  }
}

export default MessageAlertPage;
