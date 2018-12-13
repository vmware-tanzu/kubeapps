import * as React from "react";

import ErrorBoundary from "../ErrorBoundary";
import Footer from "../Footer";

import "./Layout.css";

interface ILayoutProps {
  headerComponent: React.ComponentClass<any> | React.StatelessComponent<any>;
  kubeappsVersion: string;
}

class Layout extends React.Component<ILayoutProps> {
  public render() {
    const HeaderComponent = this.props.headerComponent;
    return (
      <section className="Layout">
        <HeaderComponent />
        <main>
          <div className="container">
            <ErrorBoundary>{this.props.children}</ErrorBoundary>
          </div>
        </main>
        <Footer kubeappsVersion={this.props.kubeappsVersion} />
      </section>
    );
  }
}

export default Layout;
