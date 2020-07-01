import * as React from "react";

import ErrorBoundaryContainer from "containers/ErrorBoundaryContainer";
import Footer from "../../containers/FooterContainer";

import "./Layout.css";

interface ILayoutProps {
  children: JSX.Element;
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
        <Footer />
      </section>
    );
  }
}

export default Layout;
