import * as React from "react";

import ErrorBoundaryContainer from "containers/ErrorBoundaryContainer";

import "./Layout.v2.css";

interface ILayoutProps {
  headerComponent: React.ComponentClass<any> | React.StatelessComponent<any>;
}

class Layout extends React.Component<ILayoutProps> {
  public render() {
    const HeaderComponent = this.props.headerComponent;
    return (
      <section className="layout">
        <HeaderComponent />
        <main>
          <div className="container kubeapps-main-container">
            <div className="content-area">
              <ErrorBoundaryContainer>{this.props.children}</ErrorBoundaryContainer>
            </div>
          </div>
        </main>
      </section>
    );
  }
}

export default Layout;
