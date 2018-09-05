import * as React from "react";

import "./PageHeader.css";

interface IPageHeaderProps {
  children: Array<boolean | JSX.Element> | JSX.Element;
}

class PageHeader extends React.Component<IPageHeaderProps> {
  public render() {
    return (
      <header className="PageHeader">
        <div className="row padding-t-big padding-b-small collapse-b-phone-land align-center">
          {this.props.children}
        </div>
      </header>
    );
  }
}

export default PageHeader;
