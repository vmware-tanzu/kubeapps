import * as React from "react";

interface IMessageHeaderProps {
  children: string | JSX.Element | Array<string | JSX.Element>;
}

class MessagePageHeader extends React.Component<IMessageHeaderProps> {
  public render() {
    const { children } = this.props;
    return (
      <header>
        <div className="margin-b-big">
          <h5 className="type-regular">{children}</h5>
        </div>
      </header>
    );
  }
}

export default MessagePageHeader;
