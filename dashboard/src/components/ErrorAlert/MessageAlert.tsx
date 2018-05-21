import * as React from "react";
import { Info } from "react-feather";

import ErrorPageHeader from "./ErrorAlertHeader";

interface IMessageAlertPageProps {
  header: string;
  children?: JSX.Element;
}

class MessageAlertPage extends React.Component<IMessageAlertPageProps> {
  public render() {
    const { children, header } = this.props;
    return (
      <div className="alert margin-c margin-t-bigger">
        <ErrorPageHeader icon={Info}>{header}</ErrorPageHeader>
        {children && <div className="message__content margin-l-enormous">{children}</div>}
      </div>
    );
  }
}

export default MessageAlertPage;
