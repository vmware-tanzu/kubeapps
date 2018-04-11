import * as React from "react";

import Footer from "../Footer";

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
          <div className="container">{this.props.children}</div>
        </main>
        <Footer />
      </section>
    );
  }
}

export default Layout;
