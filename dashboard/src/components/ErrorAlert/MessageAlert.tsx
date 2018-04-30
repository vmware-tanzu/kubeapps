import * as React from "react";

import MessageAlertHeader from "./MessageAlertHeader";

interface IMessageAlertPageProps {
  header: string;
  children?: JSX.Element;
}

class MessageAlertPage extends React.Component<IMessageAlertPageProps> {
  public render() {
    const { children, header } = this.props;
    return (
      <div className="alert margin-c">
        <MessageAlertHeader>{header}</MessageAlertHeader>
        {children && <div className="message__content margin-l-enormous">{children}</div>}
      </div>
    );
  }
}

export default MessageAlertPage;
