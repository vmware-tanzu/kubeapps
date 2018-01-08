import * as React from 'react';

import Footer from './Footer';
import Header from './Header';
import Sidebar from './Sidebar';

class Layout extends React.Component {
  render() {
    return (
      <section className="Layout">
        <Header />
        <main>
          <div className="container container-fluid padding-reset">
            <div className="row">
              <div className="col-1">
                <Sidebar />
              </div>
              <div className="col-10">
                {this.props.children}
              </div>
            </div>
          </div>
        </main>
        <Footer />
      </section>
    );
  }
}

export default Layout;
