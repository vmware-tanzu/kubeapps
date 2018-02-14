import * as React from "react";

import Footer from "../Footer";
import Header from "../Header";

import "./Layout.css";

class Layout extends React.Component {
  public render() {
    return (
      <section className="Layout">
        <Header />
        <main>
          <div className="container">{this.props.children}</div>
        </main>
        <Footer />
      </section>
    );
  }
}

export default Layout;
