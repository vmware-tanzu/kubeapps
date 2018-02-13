import * as React from "react";

interface IAppHeaderProps {
  releasename: string;
}

class AppHeader extends React.Component<IAppHeaderProps> {
  public render() {
    const { releasename } = this.props;
    return (
      <header>
        <div className="AppView__heading">
          <h1>{releasename}</h1>
        </div>
        <hr />
      </header>
    );
  }
}

export default AppHeader;
