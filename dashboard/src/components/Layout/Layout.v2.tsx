import * as React from "react";

import ErrorBoundaryContainer from "containers/ErrorBoundaryContainer";

import "./Layout.css";

interface ILayoutProps {
  headerComponent: React.ComponentClass<any> | React.StatelessComponent<any>;
}

class Layout extends React.Component<ILayoutProps> {
  public render() {
    const HeaderComponent = this.props.headerComponent;
    return (
      <section className="Layout">
        <HeaderComponent />
        <main>
          <div className="container">
            <ErrorBoundaryContainer>{this.props.children}</ErrorBoundaryContainer>
          </div>
        </main>
      </section>
    );
  }
}

export default Layout;
